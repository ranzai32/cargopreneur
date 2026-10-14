package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func TestCatalogFiltersSortsAndPages(t *testing.T) {
	catalog := sampleCatalog()
	router, _ := catalogRouter(catalog, newFakeUsers())

	body, _ := request(t, router, http.MethodGet, "/products", nil, nil, http.StatusOK)
	if int(body["total"].(float64)) != 3 || body["pageSize"].(float64) != 12 {
		t.Fatalf("list = %#v", body)
	}
	if names := productNames(body); !sameStrings(names, []string{"Shop towels", "Kraft mailers", "Glass cleaner"}) {
		t.Fatalf("newest = %#v", names)
	}
	first := body["items"].([]any)[0].(map[string]any)
	tiers := first["priceTiers"].([]any)
	if len(tiers) != 3 || tiers[0].(map[string]any)["minQty"].(float64) != 40 {
		t.Fatalf("tiers = %#v", tiers)
	}

	body, _ = request(t, router, http.MethodGet, "/products?q=TOWEL", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Shop towels"}) {
		t.Fatalf("search = %#v", names)
	}
	body, _ = request(t, router, http.MethodGet, "/products?q=pallet", nil, nil, http.StatusOK)
	if int(body["total"].(float64)) != 0 {
		t.Fatal("description search matched a hidden product")
	}
	body, _ = request(t, router, http.MethodGet, "/products?q=concentrate", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Glass cleaner"}) {
		t.Fatalf("description search = %#v", names)
	}
	body, _ = request(t, router, http.MethodGet, "/products?q=MAIL", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Kraft mailers"}) {
		t.Fatalf("tag search = %#v", names)
	}

	body, _ = request(t, router, http.MethodGet, "/products?category=packaging", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Kraft mailers"}) {
		t.Fatalf("category = %#v", names)
	}
	body, _ = request(t, router, http.MethodGet, "/products?region=grand%20rapids,%20mi", nil, nil, http.StatusOK)
	if int(body["total"].(float64)) != 3 {
		t.Fatalf("region total = %#v", body["total"])
	}
	body, _ = request(t, router, http.MethodGet, "/products?supplierId="+harborID.String(), nil, nil, http.StatusOK)
	if int(body["total"].(float64)) != 3 {
		t.Fatalf("supplier total = %#v", body["total"])
	}

	body, _ = request(t, router, http.MethodGet, "/products?sort=name", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Glass cleaner", "Kraft mailers", "Shop towels"}) {
		t.Fatalf("name sort = %#v", names)
	}
	body, _ = request(t, router, http.MethodGet, "/products?sort=price_asc", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Kraft mailers", "Glass cleaner", "Shop towels"}) {
		t.Fatalf("price asc = %#v", names)
	}
	body, _ = request(t, router, http.MethodGet, "/products?sort=price_desc", nil, nil, http.StatusOK)
	if names := productNames(body); !sameStrings(names, []string{"Shop towels", "Glass cleaner", "Kraft mailers"}) {
		t.Fatalf("price desc = %#v", names)
	}

	body, _ = request(t, router, http.MethodGet, "/products?page=2&pageSize=2", nil, nil, http.StatusOK)
	if int(body["total"].(float64)) != 3 || !sameStrings(productNames(body), []string{"Glass cleaner"}) {
		t.Fatalf("page = %#v", body)
	}
}

func TestCatalogRejectsInvalidQuery(t *testing.T) {
	router, _ := catalogRouter(sampleCatalog(), newFakeUsers())
	body, _ := request(t, router, http.MethodGet, "/products?sort=cheapest&page=0&pageSize=51&supplierId=nope", nil, nil, http.StatusBadRequest)
	fields := body["fieldErrors"].(map[string]any)
	for _, field := range []string{"sort", "page", "pageSize", "supplierId"} {
		if fields[field] == nil {
			t.Fatalf("missing %s in %#v", field, body)
		}
	}
	request(t, router, http.MethodGet, "/products/"+uuid.NewString(), nil, nil, http.StatusNotFound)
	request(t, router, http.MethodGet, "/products/not-an-id", nil, nil, http.StatusBadRequest)
	request(t, router, http.MethodGet, "/suppliers?includeUnverified=maybe", nil, nil, http.StatusBadRequest)
}

func TestCatalogHidesUnverifiedSuppliers(t *testing.T) {
	catalog := sampleCatalog()
	users := newFakeUsers()
	router, tokens := catalogRouter(catalog, users)

	request(t, router, http.MethodGet, "/products/"+bulbID.String(), nil, nil, http.StatusNotFound)
	request(t, router, http.MethodGet, "/suppliers/"+quickLotID.String(), nil, nil, http.StatusNotFound)

	suppliers, _ := requestList(t, router, "/suppliers", nil, http.StatusOK)
	if names := supplierNames(suppliers); !sameStrings(names, []string{"Harbor Mill Supply"}) {
		t.Fatalf("suppliers = %#v", names)
	}
	suppliers, _ = requestList(t, router, "/suppliers?includeUnverified=true", nil, http.StatusOK)
	if names := supplierNames(suppliers); !sameStrings(names, []string{"Harbor Mill Supply"}) {
		t.Fatalf("guest includeUnverified = %#v", names)
	}

	categories, _ := requestList(t, router, "/categories", nil, http.StatusOK)
	if names := categoryNames(categories); !sameStrings(names, []string{"Cleaning", "Packaging", "Workshop supplies"}) {
		t.Fatalf("categories = %#v", names)
	}

	admin := adminCookie(t, users, tokens)
	product, _ := request(t, router, http.MethodGet, "/products/"+bulbID.String(), nil, admin, http.StatusOK)
	if product["name"] != "LED bulbs" {
		t.Fatalf("admin product = %#v", product)
	}
	supplier, _ := request(t, router, http.MethodGet, "/suppliers/"+quickLotID.String(), nil, admin, http.StatusOK)
	if supplier["isVerified"] != false || supplier["productCount"].(float64) != 1 {
		t.Fatalf("admin supplier = %#v", supplier)
	}
	suppliers, _ = requestList(t, router, "/suppliers?includeUnverified=true", admin, http.StatusOK)
	if names := supplierNames(suppliers); !sameStrings(names, []string{"Harbor Mill Supply", "QuickLot Imports"}) {
		t.Fatalf("admin suppliers = %#v", names)
	}
	listed, _ := request(t, router, http.MethodGet, "/products", nil, admin, http.StatusOK)
	if int(listed["total"].(float64)) != 4 {
		t.Fatalf("admin total = %#v", listed["total"])
	}

	business := businessCookie(t, users, tokens)
	request(t, router, http.MethodGet, "/products/"+bulbID.String(), nil, business, http.StatusNotFound)
	suppliers, _ = requestList(t, router, "/suppliers?includeUnverified=true", business, http.StatusOK)
	if len(suppliers) != 1 {
		t.Fatalf("business suppliers = %#v", suppliers)
	}
}

var (
	harborID   = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	quickLotID = uuid.MustParse("66666666-6666-4666-8666-666666666666")
	mailerID   = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	towelID    = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	cleanerID  = uuid.MustParse("77777777-7777-4777-8777-777777777705")
	bulbID     = uuid.MustParse("77777777-7777-4777-8777-777777777707")
)

func sampleCatalog() *memCatalog {
	harbor := domain.SupplierBrief{ID: harborID, Name: "Harbor Mill Supply", Region: "Grand Rapids, MI", IsVerified: true}
	quick := domain.SupplierBrief{ID: quickLotID, Name: "QuickLot Imports", Region: "Newark, NJ", IsVerified: false}
	return &memCatalog{
		suppliers: []domain.Supplier{
			{ID: harbor.ID, Name: harbor.Name, Description: "Packaging.", Region: harbor.Region, IsVerified: true},
			{ID: quick.ID, Name: quick.Name, Description: "Awaiting review.", Region: quick.Region, IsVerified: false},
		},
		products: []domain.Product{
			{
				ID: mailerID, Name: "Kraft mailers", Description: "Carton of boxes.", Category: "Packaging",
				Tags: []string{"mailer"}, Unit: "carton", RetailPrice: 16.5, Supplier: harbor,
				PriceTiers: []domain.PriceTier{{100, 11.2}, {140, 9.6}, {180, 8.4}},
				CreatedAt:  time.Date(2026, 7, 1, 15, 0, 0, 0, time.UTC),
			},
			{
				ID: towelID, Name: "Shop towels", Description: "Blue shop towels.", Category: "Workshop supplies",
				Tags: []string{"towels"}, Unit: "case", RetailPrice: 28, Supplier: harbor,
				PriceTiers: []domain.PriceTier{{40, 19.5}, {80, 16.75}, {120, 14.25}},
				CreatedAt:  time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC),
			},
			{
				ID: cleanerID, Name: "Glass cleaner", Description: "Concentrate for counters.", Category: "Cleaning",
				Tags: []string{"cleaner"}, Unit: "gallon", RetailPrice: 18.25, Supplier: harbor,
				PriceTiers: []domain.PriceTier{{12, 12.4}, {24, 10.8}, {48, 9.35}},
				CreatedAt:  time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC),
			},
			{
				ID: bulbID, Name: "LED bulbs", Description: "Pallet wrap is not in this name.", Category: "Lighting",
				Tags: []string{"bulbs"}, Unit: "pack", RetailPrice: 24, Supplier: quick,
				PriceTiers: []domain.PriceTier{{10, 15.5}, {25, 13.2}, {50, 11.4}},
				CreatedAt:  time.Date(2026, 9, 1, 15, 0, 0, 0, time.UTC),
			},
		},
	}
}

type memCatalog struct {
	suppliers []domain.Supplier
	products  []domain.Product
}

func (m *memCatalog) ListProducts(_ context.Context, query domain.ProductQuery) ([]domain.Product, int, error) {
	matched := make([]domain.Product, 0)
	for _, product := range m.products {
		if !query.IncludeUnverified && !product.Supplier.IsVerified {
			continue
		}
		if query.Q != "" && !productMatches(product, query.Q) {
			continue
		}
		if query.Category != "" && !strings.EqualFold(product.Category, query.Category) {
			continue
		}
		if query.SupplierID != nil && product.Supplier.ID != *query.SupplierID {
			continue
		}
		if query.Region != "" && !strings.EqualFold(product.Supplier.Region, query.Region) {
			continue
		}
		matched = append(matched, product)
	}
	sort.SliceStable(matched, func(i, j int) bool {
		left, right := matched[i], matched[j]
		switch query.Sort {
		case domain.SortName:
			if left.Name != right.Name {
				return left.Name < right.Name
			}
			return left.ID.String() < right.ID.String()
		case domain.SortPriceAsc, domain.SortPriceDesc:
			leftPrice, leftOK := domain.LowestUnitPrice(left.PriceTiers)
			rightPrice, rightOK := domain.LowestUnitPrice(right.PriceTiers)
			if leftOK != rightOK {
				return leftOK
			}
			if leftOK && leftPrice != rightPrice {
				if query.Sort == domain.SortPriceDesc {
					return leftPrice > rightPrice
				}
				return leftPrice < rightPrice
			}
			return left.ID.String() < right.ID.String()
		default:
			if !left.CreatedAt.Equal(right.CreatedAt) {
				return left.CreatedAt.After(right.CreatedAt)
			}
			return left.ID.String() > right.ID.String()
		}
	})
	total := len(matched)
	start := (query.Page - 1) * query.PageSize
	if start > total {
		return []domain.Product{}, total, nil
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (m *memCatalog) FindProduct(_ context.Context, id uuid.UUID, includeUnverified bool) (domain.Product, error) {
	for _, product := range m.products {
		if product.ID != id {
			continue
		}
		if !includeUnverified && !product.Supplier.IsVerified {
			return domain.Product{}, domain.ErrNotFound
		}
		return product, nil
	}
	return domain.Product{}, domain.ErrNotFound
}

func (m *memCatalog) ListSuppliers(_ context.Context, includeUnverified bool) ([]domain.Supplier, error) {
	listed := make([]domain.Supplier, 0)
	for _, supplier := range m.suppliers {
		if !includeUnverified && !supplier.IsVerified {
			continue
		}
		supplier.ProductCount = m.countProducts(supplier.ID)
		listed = append(listed, supplier)
	}
	sort.Slice(listed, func(i, j int) bool { return listed[i].Name < listed[j].Name })
	return listed, nil
}

func (m *memCatalog) FindSupplier(_ context.Context, id uuid.UUID, includeUnverified bool) (domain.Supplier, error) {
	for _, supplier := range m.suppliers {
		if supplier.ID != id {
			continue
		}
		if !includeUnverified && !supplier.IsVerified {
			return domain.Supplier{}, domain.ErrNotFound
		}
		supplier.ProductCount = m.countProducts(supplier.ID)
		return supplier, nil
	}
	return domain.Supplier{}, domain.ErrNotFound
}

func (m *memCatalog) ListCategories(_ context.Context, includeUnverified bool) ([]domain.Category, error) {
	counts := map[string]int{}
	for _, product := range m.products {
		if !includeUnverified && !product.Supplier.IsVerified {
			continue
		}
		counts[product.Category]++
	}
	categories := make([]domain.Category, 0, len(counts))
	for name, count := range counts {
		categories = append(categories, domain.Category{Name: name, Count: count})
	}
	sort.Slice(categories, func(i, j int) bool {
		return strings.ToLower(categories[i].Name) < strings.ToLower(categories[j].Name)
	})
	return categories, nil
}

func (m *memCatalog) countProducts(supplierID uuid.UUID) int {
	count := 0
	for _, product := range m.products {
		if product.Supplier.ID == supplierID {
			count++
		}
	}
	return count
}

func productMatches(product domain.Product, term string) bool {
	term = strings.ToLower(term)
	if strings.Contains(strings.ToLower(product.Name), term) || strings.Contains(strings.ToLower(product.Description), term) {
		return true
	}
	for _, tag := range product.Tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return true
		}
	}
	return false
}

func catalogRouter(catalog *memCatalog, users *fakeUsers) (http.Handler, *auth.Tokens) {
	tokens := auth.NewTokens("test-secret-value", time.Hour)
	if users == nil {
		users = newFakeUsers()
	}
	return NewRouter(RouterConfig{
		Users:      users,
		Pinger:     users,
		Passwords:  auth.NewTestHasher(),
		Tokens:     tokens,
		CORSOrigin: "http://localhost:5173",
		Limiter:    NewRateLimiter(100, time.Minute),
		Catalog:    catalog,
	}), tokens
}

func adminCookie(t *testing.T, users *fakeUsers, tokens *auth.Tokens) *http.Cookie {
	t.Helper()
	return userCookie(t, users, tokens, domain.RoleAdmin, "admin-catalog@harbor.example")
}

func businessCookie(t *testing.T, users *fakeUsers, tokens *auth.Tokens) *http.Cookie {
	t.Helper()
	return userCookie(t, users, tokens, domain.RoleBusiness, "business-catalog@harbor.example")
}

func userCookie(t *testing.T, users *fakeUsers, tokens *auth.Tokens, role, email string) *http.Cookie {
	t.Helper()
	user, err := users.CreateUser(context.Background(), domain.NewUser{
		BusinessName: "Shop",
		ContactName:  "Owner",
		Email:        email,
		PasswordHash: "hash",
		Region:       "MI",
		Role:         role,
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := tokens.Issue(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: sessionCookie, Value: raw}
}

func requestList(t *testing.T, router http.Handler, path string, cookie *http.Cookie, status int) ([]any, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	router.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("GET %s status = %d body = %s", path, rec.Code, rec.Body.String())
	}
	var body []any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body, rec
}

func productNames(body map[string]any) []string {
	items, _ := body["items"].([]any)
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.(map[string]any)["name"].(string))
	}
	return names
}

func supplierNames(items []any) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.(map[string]any)["name"].(string))
	}
	return names
}

func categoryNames(items []any) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.(map[string]any)["name"].(string))
	}
	return names
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
