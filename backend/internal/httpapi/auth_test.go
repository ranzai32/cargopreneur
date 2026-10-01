package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func TestAuthFlow(t *testing.T) {
	users := newFakeUsers()
	router := testRouter(users)

	body, rec := registerBody(t, router, map[string]string{
		"businessName": "Harbor Shop",
		"contactName":  "Ada Lopez",
		"email":        "Ada@Harbor.Example",
		"password":     "long-enough",
		"region":       "Grand Rapids, MI",
	}, http.StatusCreated)

	if _, ok := body["password"]; ok {
		t.Fatal("response included a password")
	}
	if _, ok := body["passwordHash"]; ok {
		t.Fatal("response included a password hash")
	}
	if body["email"] != "ada@harbor.example" || body["role"] != "business" || body["businessName"] != "Harbor Shop" {
		t.Fatalf("user = %#v", body)
	}

	cookie := sessionCookieFrom(t, rec)
	me, _ := request(t, router, http.MethodGet, "/auth/me", nil, cookie, http.StatusOK)
	if me["email"] != "ada@harbor.example" {
		t.Fatalf("me = %#v", me)
	}

	logout := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(cookie)
	router.ServeHTTP(logout, req)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", logout.Code)
	}
	cleared := logout.Header().Get("Set-Cookie")
	if cleared == "" || !containsAll(cleared, "session=", "HttpOnly", "Max-Age=0") && !containsAll(cleared, "session=", "HttpOnly", "Max-Age=-1") {
		t.Fatalf("clear cookie = %q", cleared)
	}

	_, _ = request(t, router, http.MethodGet, "/auth/me", nil, nil, http.StatusUnauthorized)

	login, loginRec := request(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "ada@harbor.example",
		"password": "long-enough",
	}, nil, http.StatusOK)
	if login["id"] != body["id"] {
		t.Fatalf("login id = %#v", login)
	}
	loginCookie := sessionCookieFrom(t, loginRec)
	if loginCookie.HttpOnly != true || loginCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie httpOnly=%v sameSite=%v", loginCookie.HttpOnly, loginCookie.SameSite)
	}
}

func TestRegisterValidationAndConflict(t *testing.T) {
	router := testRouter(newFakeUsers())
	errorsBody, _ := request(t, router, http.MethodPost, "/auth/register", map[string]string{
		"businessName": "",
		"email":        "nope",
		"password":     "short",
	}, nil, http.StatusUnprocessableEntity)
	fields, _ := errorsBody["fieldErrors"].(map[string]any)
	if fields["email"] == nil || fields["password"] == nil || fields["region"] == nil {
		t.Fatalf("fieldErrors = %#v", errorsBody)
	}

	_, _ = registerBody(t, router, map[string]string{
		"businessName": "Harbor Shop",
		"contactName":  "Ada Lopez",
		"email":        "ada@harbor.example",
		"password":     "long-enough",
		"region":       "Grand Rapids, MI",
	}, http.StatusCreated)
	conflict, _ := request(t, router, http.MethodPost, "/auth/register", map[string]string{
		"businessName": "Other Shop",
		"contactName":  "Bea",
		"email":        "ADA@harbor.example",
		"password":     "long-enough",
		"region":       "Columbus, OH",
	}, nil, http.StatusConflict)
	conflictFields, _ := conflict["fieldErrors"].(map[string]any)
	if conflictFields["email"] == nil {
		t.Fatalf("conflict = %#v", conflict)
	}
}

func TestLoginHidesWhichFieldFailed(t *testing.T) {
	router := testRouter(newFakeUsers())
	_, _ = registerBody(t, router, map[string]string{
		"businessName": "Harbor Shop",
		"contactName":  "Ada Lopez",
		"email":        "ada@harbor.example",
		"password":     "long-enough",
		"region":       "Grand Rapids, MI",
	}, http.StatusCreated)

	missing, _ := request(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email": "missing@harbor.example", "password": "long-enough",
	}, nil, http.StatusUnauthorized)
	wrong, _ := request(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email": "ada@harbor.example", "password": "wrong-password",
	}, nil, http.StatusUnauthorized)
	if missing["message"] != wrong["message"] {
		t.Fatalf("messages differ: %#v vs %#v", missing, wrong)
	}
}

func TestInvalidJSONAndHealth(t *testing.T) {
	users := newFakeUsers()
	router := testRouter(users)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{"))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}

	health, _ := request(t, router, http.MethodGet, "/health", nil, nil, http.StatusOK)
	if health["status"] != "ok" {
		t.Fatalf("health = %#v", health)
	}
	users.pingErr = io.EOF
	_, _ = request(t, router, http.MethodGet, "/health", nil, nil, http.StatusServiceUnavailable)
}

func TestRequireRole(t *testing.T) {
	users := newFakeUsers()
	tokens := auth.NewTokens("test-secret-value", time.Hour)
	router := NewRouter(RouterConfig{
		Users:        users,
		Pinger:       users,
		Passwords:    auth.NewTestHasher(),
		Tokens:       tokens,
		CORSOrigin:   "http://localhost:5173",
		Limiter:      NewRateLimiter(100, time.Minute),
	})
	protected := requireAuth(users, tokens)(RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	business := mustUser(t, users, domain.RoleBusiness)
	businessToken, err := tokens.Issue(business.ID)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: businessToken})
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("business status = %d", rec.Code)
	}

	admin := mustUser(t, users, domain.RoleAdmin)
	adminToken, err := tokens.Issue(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: adminToken})
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("admin status = %d", rec.Code)
	}

	cors := httptest.NewRecorder()
	preflight := httptest.NewRequest(http.MethodOptions, "/auth/login", nil)
	preflight.Header.Set("Origin", "http://localhost:5173")
	preflight.Header.Set("Access-Control-Request-Method", "POST")
	router.ServeHTTP(cors, preflight)
	if cors.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("cors origin = %q", cors.Header().Get("Access-Control-Allow-Origin"))
	}
	if cors.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("credentials were not allowed")
	}
}

type fakeUsers struct {
	mu           sync.Mutex
	byEmail      map[string]domain.User
	byID         map[uuid.UUID]domain.User
	pingErr error
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		byEmail: map[string]domain.User{},
		byID:    map[uuid.UUID]domain.User{},
	}
}

func (f *fakeUsers) Ping(context.Context) error { return f.pingErr }

func (f *fakeUsers) CreateUser(_ context.Context, input domain.NewUser) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byEmail[input.Email]; ok {
		return domain.User{}, domain.ErrEmailTaken
	}
	role := input.Role
	if role == "" {
		role = domain.RoleBusiness
	}
	user := domain.User{
		ID:           uuid.New(),
		BusinessName: input.BusinessName,
		ContactName:  input.ContactName,
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		Region:       input.Region,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	return user, nil
}

func (f *fakeUsers) FindUserByEmail(_ context.Context, email string) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	user, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (f *fakeUsers) FindUserByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	user, ok := f.byID[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func testRouter(users *fakeUsers) http.Handler {
	return NewRouter(RouterConfig{
		Users:      users,
		Pinger:     users,
		Passwords:  auth.NewTestHasher(),
		Tokens:     auth.NewTokens("test-secret-value", time.Hour),
		CORSOrigin: "http://localhost:5173",
		Limiter:    NewRateLimiter(100, time.Minute),
	})
}

func registerBody(t *testing.T, router http.Handler, payload map[string]string, status int) (map[string]any, *httptest.ResponseRecorder) {
	t.Helper()
	return request(t, router, http.MethodPost, "/auth/register", payload, nil, status)
}

func request(t *testing.T, router http.Handler, method, path string, payload any, cookie *http.Cookie, status int) (map[string]any, *httptest.ResponseRecorder) {
	t.Helper()
	var reader io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, reader)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s status = %d body = %s", method, path, rec.Code, rec.Body.String())
	}
	if rec.Body.Len() == 0 {
		return nil, rec
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body, rec
}

func sessionCookieFrom(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	if rec == nil {
		t.Fatal("no response recorded")
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookie {
			return cookie
		}
	}
	t.Fatalf("session cookie missing: %q", rec.Header().Get("Set-Cookie"))
	return nil
}

func mustUser(t *testing.T, users *fakeUsers, role string) domain.User {
	t.Helper()
	user, err := users.CreateUser(context.Background(), domain.NewUser{
		BusinessName: "Shop",
		ContactName:  "Owner",
		Email:        role + "@harbor.example",
		PasswordHash: "hash",
		Region:       "MI",
		Role:         role,
	})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !bytes.Contains([]byte(value), []byte(part)) {
			return false
		}
	}
	return true
}
