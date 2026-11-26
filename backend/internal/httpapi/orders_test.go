package httpapi

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func TestOrderListFiltersAndViewer(t *testing.T) {
	orders := newMemOrders()
	users := newFakeUsers()
	router, tokens := orderRouter(orders, users)
	owner, ownerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "owner-orders@harbor.example")
	_, strangerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "stranger-orders@harbor.example")

	openID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	confirmedID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	cancelledID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3")
	expiredID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4")
	future := time.Date(2099, 1, 15, 23, 59, 59, 0, time.UTC)
	orders.seed(storedOrder{
		id: openID, productID: towelID, status: domain.OrderOpen, moq: 20, deadline: future,
		createdAt: time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: owner.ID, quantity: 5}},
	})
	orders.seed(storedOrder{
		id: confirmedID, productID: mailerID, status: domain.OrderConfirmed, moq: 10, deadline: future,
		createdAt: time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: uuid.New(), quantity: 10}},
	})
	orders.seed(storedOrder{
		id: cancelledID, productID: cleanerID, status: domain.OrderCancelled, moq: 12, deadline: future,
		createdAt: time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: uuid.New(), quantity: 2}},
	})
	orders.seed(storedOrder{
		id: expiredID, productID: towelID, status: domain.OrderOpen, moq: 40, deadline: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
		createdAt: time.Date(2026, 11, 1, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: uuid.New(), quantity: 3}},
	})

	listed, _ := requestList(t, router, "/orders", nil, http.StatusOK)
	if !sameStrings(orderIDs(listed), []string{openID.String(), confirmedID.String(), cancelledID.String(), expiredID.String()}) {
		t.Fatalf("all = %#v", orderIDs(listed))
	}
	for _, item := range listed {
		order := item.(map[string]any)
		if _, ok := order["yourQuantity"]; !ok || order["yourQuantity"] != nil {
			t.Fatalf("anonymous yourQuantity = %#v", order["yourQuantity"])
		}
		assertOrderShape(t, order)
	}
	if listed[3].(map[string]any)["status"] != domain.OrderCancelled {
		t.Fatalf("expired status = %#v", listed[3])
	}

	requestList(t, router, "/orders", &http.Cookie{Name: sessionCookie, Value: "not-a-session"}, http.StatusOK)
	ownerList, _ := requestList(t, router, "/orders?status=all", ownerCookie, http.StatusOK)
	openOrder := findOrder(t, ownerList, openID.String())
	if intField(t, openOrder, "yourQuantity") != 5 || intField(t, openOrder, "collectedQty") != 5 || intField(t, openOrder, "participantCount") != 1 {
		t.Fatalf("owner order = %#v", openOrder)
	}
	if findOrder(t, ownerList, confirmedID.String())["yourQuantity"] != nil {
		t.Fatal("stranger participation was filled for the owner")
	}
	strangerList, _ := requestList(t, router, "/orders", strangerCookie, http.StatusOK)
	if findOrder(t, strangerList, openID.String())["yourQuantity"] != nil {
		t.Fatal("another user's quantity was returned")
	}

	openOnly, _ := requestList(t, router, "/orders?status=open", nil, http.StatusOK)
	if !sameStrings(orderIDs(openOnly), []string{openID.String()}) {
		t.Fatalf("open = %#v", orderIDs(openOnly))
	}
	confirmedOnly, _ := requestList(t, router, "/orders?status=confirmed", nil, http.StatusOK)
	if !sameStrings(orderIDs(confirmedOnly), []string{confirmedID.String()}) {
		t.Fatalf("confirmed = %#v", orderIDs(confirmedOnly))
	}
	cancelledOnly, _ := requestList(t, router, "/orders?status=cancelled", nil, http.StatusOK)
	if !sameStrings(orderIDs(cancelledOnly), []string{cancelledID.String(), expiredID.String()}) {
		t.Fatalf("cancelled = %#v", orderIDs(cancelledOnly))
	}

	invalid, _ := request(t, router, http.MethodGet, "/orders?status=paused", nil, nil, http.StatusBadRequest)
	if invalid["message"] != "Invalid query" || invalid["fieldErrors"].(map[string]any)["status"] != "Use open, confirmed, cancelled, or all" {
		t.Fatalf("invalid status = %#v", invalid)
	}

	fetched, _ := request(t, router, http.MethodGet, "/orders/"+expiredID.String(), nil, nil, http.StatusOK)
	if fetched["status"] != domain.OrderCancelled || fetched["yourQuantity"] != nil {
		t.Fatalf("expired get = %#v", fetched)
	}
	ownerFetched, _ := request(t, router, http.MethodGet, "/orders/"+openID.String(), nil, ownerCookie, http.StatusOK)
	if intField(t, ownerFetched, "yourQuantity") != 5 {
		t.Fatalf("owner get = %#v", ownerFetched)
	}
	product := ownerFetched["product"].(map[string]any)
	if product["retailPrice"].(float64) != 28 || product["priceTiers"].([]any)[0].(map[string]any)["unitPrice"].(float64) != 19.5 {
		t.Fatalf("product prices = %#v", product)
	}
	missing, _ := request(t, router, http.MethodGet, "/orders/"+uuid.NewString(), nil, nil, http.StatusNotFound)
	if missing["message"] != "Order not found" {
		t.Fatalf("missing = %#v", missing)
	}
	request(t, router, http.MethodGet, "/orders/not-an-id", nil, nil, http.StatusBadRequest)

	closed, _ := request(t, router, http.MethodPost, "/orders/"+expiredID.String()+"/join", map[string]any{"quantity": 8}, ownerCookie, http.StatusBadRequest)
	if closed["message"] != "This order is no longer open" {
		t.Fatalf("join expired = %#v", closed)
	}
	afterJoin, _ := request(t, router, http.MethodGet, "/orders/"+expiredID.String(), nil, nil, http.StatusOK)
	if afterJoin["status"] != domain.OrderCancelled || intField(t, afterJoin, "collectedQty") != 3 {
		t.Fatalf("expired after join = %#v", afterJoin)
	}
}

func TestCreateOrderConfirmsOrStaysOpen(t *testing.T) {
	orders := newMemOrders()
	users := newFakeUsers()
	router, tokens := orderRouter(orders, users)
	_, cookie := signedIn(t, users, tokens, domain.RoleBusiness, "creator-orders@harbor.example")

	confirmed, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": towelID.String(),
		"moq":       10,
		"deadline":  "2099-06-01T15:04:05Z",
		"quantity":  10,
	}, cookie, http.StatusCreated)
	if confirmed["status"] != domain.OrderConfirmed || intField(t, confirmed, "collectedQty") != 10 || intField(t, confirmed, "participantCount") != 1 || intField(t, confirmed, "yourQuantity") != 10 {
		t.Fatalf("confirmed = %#v", confirmed)
	}
	assertOrderShape(t, confirmed)
	deadline, err := time.Parse(time.RFC3339, confirmed["deadline"].(string))
	if err != nil || !deadline.Equal(time.Date(2099, 6, 1, 15, 4, 5, 0, time.UTC)) {
		t.Fatalf("deadline = %#v", confirmed["deadline"])
	}

	opened, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": towelID.String(),
		"moq":       10,
		"deadline":  "2099-01-15",
		"quantity":  4,
	}, cookie, http.StatusCreated)
	if opened["status"] != domain.OrderOpen || intField(t, opened, "collectedQty") != 4 || intField(t, opened, "yourQuantity") != 4 {
		t.Fatalf("open = %#v", opened)
	}
	openedDeadline, err := time.Parse(time.RFC3339, opened["deadline"].(string))
	if err != nil || !openedDeadline.Equal(time.Date(2099, 1, 15, 23, 59, 59, 0, time.UTC)) {
		t.Fatalf("date deadline = %#v", opened["deadline"])
	}
	if opened["product"].(map[string]any)["name"] != "Shop towels" {
		t.Fatalf("product = %#v", opened["product"])
	}
}

func TestCreateOrderRejectsInvalidFields(t *testing.T) {
	users := newFakeUsers()
	router, tokens := orderRouter(newMemOrders(), users)
	_, cookie := signedIn(t, users, tokens, domain.RoleBusiness, "invalid-orders@harbor.example")

	body, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": "nope",
		"moq":       1.5,
		"deadline":  "tomorrow",
		"quantity":  0,
	}, cookie, http.StatusUnprocessableEntity)
	if body["message"] != "Check the highlighted fields" {
		t.Fatalf("message = %#v", body)
	}
	fields := body["fieldErrors"].(map[string]any)
	if fields["productId"] != "Must be a valid id" || fields["moq"] != "Enter a minimum of at least 1" || fields["deadline"] != "Enter a valid date" || fields["quantity"] != "Enter the quantity you need" {
		t.Fatalf("fields = %#v", fields)
	}

	missing, _ := request(t, router, http.MethodPost, "/orders", map[string]any{}, cookie, http.StatusUnprocessableEntity)
	missingFields := missing["fieldErrors"].(map[string]any)
	for _, key := range []string{"productId", "moq", "deadline", "quantity"} {
		if missingFields[key] == nil {
			t.Fatalf("missing %s in %#v", key, missing)
		}
	}
	if missingFields["deadline"] != "Choose a deadline" {
		t.Fatalf("deadline = %#v", missingFields["deadline"])
	}

	past, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": towelID.String(),
		"moq":       10,
		"deadline":  "2020-01-01",
		"quantity":  2,
	}, cookie, http.StatusUnprocessableEntity)
	if past["fieldErrors"].(map[string]any)["deadline"] != "Choose a deadline that has not passed" {
		t.Fatalf("past = %#v", past)
	}
}

func TestCreateOrderRejectsUnknownProducts(t *testing.T) {
	users := newFakeUsers()
	router, tokens := orderRouter(newMemOrders(), users)
	_, business := signedIn(t, users, tokens, domain.RoleBusiness, "business-orders@harbor.example")
	_, admin := signedIn(t, users, tokens, domain.RoleAdmin, "admin-orders@harbor.example")

	unknown, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": uuid.NewString(),
		"moq":       10,
		"deadline":  "2099-01-15",
		"quantity":  2,
	}, business, http.StatusBadRequest)
	if unknown["fieldErrors"].(map[string]any)["productId"] != "Choose a product from the catalog" {
		t.Fatalf("unknown = %#v", unknown)
	}
	hidden, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": bulbID.String(),
		"moq":       10,
		"deadline":  "2099-01-15",
		"quantity":  2,
	}, business, http.StatusBadRequest)
	if hidden["fieldErrors"].(map[string]any)["productId"] != "Choose a product from the catalog" {
		t.Fatalf("hidden = %#v", hidden)
	}
	created, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": bulbID.String(),
		"moq":       10,
		"deadline":  "2099-01-15",
		"quantity":  2,
	}, admin, http.StatusCreated)
	if created["product"].(map[string]any)["name"] != "LED bulbs" {
		t.Fatalf("admin order = %#v", created["product"])
	}
}

func TestJoinOrderAggregatesAndRejects(t *testing.T) {
	users := newFakeUsers()
	router, tokens := orderRouter(newMemOrders(), users)
	_, creator := signedIn(t, users, tokens, domain.RoleBusiness, "join-creator@harbor.example")
	_, joiner := signedIn(t, users, tokens, domain.RoleBusiness, "join-member@harbor.example")
	_, closer := signedIn(t, users, tokens, domain.RoleBusiness, "join-closer@harbor.example")
	_, late := signedIn(t, users, tokens, domain.RoleBusiness, "join-late@harbor.example")

	created, _ := request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": towelID.String(),
		"moq":       100,
		"deadline":  "2099-01-15",
		"quantity":  40,
	}, creator, http.StatusCreated)
	id := created["id"].(string)

	duplicate, _ := request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 10}, creator, http.StatusConflict)
	if duplicate["message"] != "You already joined this order" || duplicate["fieldErrors"].(map[string]any)["quantity"] != "You already joined this order" {
		t.Fatalf("creator join = %#v", duplicate)
	}

	partial, _ := request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 30}, joiner, http.StatusOK)
	if partial["status"] != domain.OrderOpen || intField(t, partial, "collectedQty") != 70 || intField(t, partial, "participantCount") != 2 || intField(t, partial, "yourQuantity") != 30 {
		t.Fatalf("partial = %#v", partial)
	}
	again, _ := request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 5}, joiner, http.StatusConflict)
	if again["message"] != "You already joined this order" {
		t.Fatalf("second join = %#v", again)
	}

	confirmed, _ := request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 40}, closer, http.StatusOK)
	if confirmed["status"] != domain.OrderConfirmed || intField(t, confirmed, "collectedQty") != 110 || intField(t, confirmed, "participantCount") != 3 || intField(t, confirmed, "yourQuantity") != 40 {
		t.Fatalf("confirmed join = %#v", confirmed)
	}
	closed, _ := request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 5}, late, http.StatusBadRequest)
	if closed["message"] != "This order is no longer open" {
		t.Fatalf("closed = %#v", closed)
	}

	invalidQty, _ := request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 0}, late, http.StatusUnprocessableEntity)
	if invalidQty["message"] != "Check the highlighted fields" || invalidQty["fieldErrors"].(map[string]any)["quantity"] != "Enter a quantity of at least 1" {
		t.Fatalf("quantity = %#v", invalidQty)
	}
	request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{}, late, http.StatusUnprocessableEntity)
	request(t, router, http.MethodPost, "/orders/"+id+"/join", map[string]any{"quantity": 1.5}, late, http.StatusUnprocessableEntity)
	missing, _ := request(t, router, http.MethodPost, "/orders/"+uuid.NewString()+"/join", map[string]any{"quantity": 1}, late, http.StatusNotFound)
	if missing["message"] != "Order not found" {
		t.Fatalf("missing = %#v", missing)
	}
	request(t, router, http.MethodPost, "/orders/not-an-id/join", map[string]any{"quantity": 1}, late, http.StatusBadRequest)
}

func TestPayOrderRecordsSimulatedPayment(t *testing.T) {
	orders := newMemOrders()
	users := newFakeUsers()
	notices := &captureNotifier{}
	router, tokens := orderRouterWith(orders, users, notices)
	payer, payerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "payer@harbor.example")
	partner, partnerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "partner@harbor.example")
	_, outsiderCookie := signedIn(t, users, tokens, domain.RoleBusiness, "outsider@harbor.example")

	future := time.Date(2099, 3, 1, 23, 59, 59, 0, time.UTC)
	confirmedID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1")
	openID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2")
	cancelledID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3")
	orders.seed(storedOrder{
		id: confirmedID, productID: towelID, status: domain.OrderConfirmed, moq: 10, deadline: future,
		createdAt: time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC),
		parts: []participation{
			{userID: payer.ID, quantity: 6},
			{userID: partner.ID, quantity: 4},
		},
	})
	orders.seed(storedOrder{
		id: openID, productID: towelID, status: domain.OrderOpen, moq: 20, deadline: future,
		createdAt: time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: payer.ID, quantity: 2}},
	})
	orders.seed(storedOrder{
		id: cancelledID, productID: mailerID, status: domain.OrderCancelled, moq: 8, deadline: future,
		createdAt: time.Date(2026, 11, 12, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: payer.ID, quantity: 1}},
	})

	denied, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/pay", map[string]any{"shippingAddress": "12 Dock Street", "deliveryMethod": "freight"}, nil, http.StatusUnauthorized)
	if denied["message"] != "Authentication required" {
		t.Fatalf("auth = %#v", denied)
	}
	missing, _ := request(t, router, http.MethodPost, "/orders/"+uuid.NewString()+"/pay", map[string]any{"shippingAddress": "12 Dock Street", "deliveryMethod": "freight"}, payerCookie, http.StatusNotFound)
	if missing["message"] != "Order not found" {
		t.Fatalf("missing = %#v", missing)
	}
	early, _ := request(t, router, http.MethodPost, "/orders/"+openID.String()+"/pay", map[string]any{"shippingAddress": "12 Dock Street", "deliveryMethod": "freight"}, payerCookie, http.StatusBadRequest)
	if early["message"] != "This order is not ready for payment" {
		t.Fatalf("open = %#v", early)
	}
	stillOpen, _ := request(t, router, http.MethodGet, "/orders/"+openID.String(), nil, payerCookie, http.StatusOK)
	if stillOpen["status"] != domain.OrderOpen || stillOpen["yourPaymentStatus"] != domain.PaymentUnpaid || stillOpen["allPaid"] != false {
		t.Fatalf("open after pay = %#v", stillOpen)
	}
	cancelled, _ := request(t, router, http.MethodPost, "/orders/"+cancelledID.String()+"/pay", map[string]any{"shippingAddress": "12 Dock Street", "deliveryMethod": "freight"}, payerCookie, http.StatusBadRequest)
	if cancelled["message"] != "This order is not ready for payment" {
		t.Fatalf("cancelled = %#v", cancelled)
	}
	outsider, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/pay", map[string]any{"shippingAddress": "9 River Road", "deliveryMethod": "freight"}, outsiderCookie, http.StatusForbidden)
	if outsider["message"] != "You are not part of this order" {
		t.Fatalf("outsider = %#v", outsider)
	}
	blank, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/pay", map[string]any{"shippingAddress": "   ", "deliveryMethod": "freight"}, payerCookie, http.StatusUnprocessableEntity)
	if blank["message"] != "Check the highlighted fields" || blank["fieldErrors"].(map[string]any)["shippingAddress"] != "Enter a shipping address" {
		t.Fatalf("blank = %#v", blank)
	}
	if len(notices.list()) != 0 {
		t.Fatalf("notices before payment = %#v", notices.list())
	}

	paid, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/pay", map[string]any{"shippingAddress": "  12 Dock Street  ", "deliveryMethod": "freight"}, payerCookie, http.StatusOK)
	assertOrderShape(t, paid)
	if paid["status"] != domain.OrderConfirmed || paid["yourPaymentStatus"] != domain.PaymentPaid || paid["allPaid"] != false || intField(t, paid, "yourQuantity") != 6 || intField(t, paid, "collectedQty") != 10 {
		t.Fatalf("first pay = %#v", paid)
	}
	if paid["yourDeliveryMethod"] != "freight" || paid["yourShippingFee"] != 30.7 {
		t.Fatalf("delivery = %#v %#v", paid["yourDeliveryMethod"], paid["yourShippingFee"])
	}
	if saved := savedAddress(t, orders, confirmedID, payer.ID); saved != "12 Dock Street" {
		t.Fatalf("address = %q", saved)
	}
	recorded := notices.list()
	if len(recorded) != 1 || recorded[0].Kind != domain.NoticePaymentReceived || recorded[0].OrderID != confirmedID || len(recorded[0].UserIDs) != 1 || recorded[0].UserIDs[0] != payer.ID {
		t.Fatalf("notice = %#v", recorded)
	}

	again, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/pay", map[string]any{"shippingAddress": "12 Dock Street", "deliveryMethod": "freight"}, payerCookie, http.StatusConflict)
	if again["message"] != "This order is already paid" {
		t.Fatalf("second pay = %#v", again)
	}
	if len(notices.list()) != 1 {
		t.Fatalf("notices after second pay = %#v", notices.list())
	}

	finished, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/pay", map[string]any{"shippingAddress": "4 Market Street", "deliveryMethod": "express"}, partnerCookie, http.StatusOK)
	assertOrderShape(t, finished)
	if finished["status"] != domain.OrderPaid || finished["yourPaymentStatus"] != domain.PaymentPaid || finished["allPaid"] != true || intField(t, finished, "yourQuantity") != 4 {
		t.Fatalf("all paid = %#v", finished)
	}
	payerView, _ := request(t, router, http.MethodGet, "/orders/"+confirmedID.String(), nil, payerCookie, http.StatusOK)
	if payerView["status"] != domain.OrderPaid || payerView["yourPaymentStatus"] != domain.PaymentPaid || payerView["allPaid"] != true || intField(t, payerView, "yourQuantity") != 6 {
		t.Fatalf("payer view = %#v", payerView)
	}
	strangerView, _ := request(t, router, http.MethodGet, "/orders/"+confirmedID.String(), nil, outsiderCookie, http.StatusOK)
	if strangerView["yourPaymentStatus"] != nil || strangerView["allPaid"] != true || strangerView["yourQuantity"] != nil {
		t.Fatalf("stranger view = %#v", strangerView)
	}
	recorded = notices.list()
	if len(recorded) != 2 || recorded[1].Kind != domain.NoticePaymentReceived || len(recorded[1].UserIDs) != 1 || recorded[1].UserIDs[0] != partner.ID {
		t.Fatalf("notices = %#v", recorded)
	}
}

func TestOrderWritesRequireAuth(t *testing.T) {
	router, _ := orderRouter(newMemOrders(), newFakeUsers())
	request(t, router, http.MethodPost, "/orders", map[string]any{
		"productId": towelID.String(),
		"moq":       10,
		"deadline":  "2099-01-15",
		"quantity":  1,
	}, nil, http.StatusUnauthorized)
	denied, _ := request(t, router, http.MethodPost, "/orders/"+uuid.NewString()+"/join", map[string]any{"quantity": 1}, nil, http.StatusUnauthorized)
	if denied["message"] != "Authentication required" {
		t.Fatalf("auth = %#v", denied)
	}
}

func orderRouter(orders *memOrders, users *fakeUsers) (http.Handler, *auth.Tokens) {
	return orderRouterWith(orders, users, domain.NopNotifier{})
}

func orderRouterWith(orders *memOrders, users *fakeUsers, notifier domain.Notifier) (http.Handler, *auth.Tokens) {
	tokens := auth.NewTokens("test-secret-value", time.Hour)
	if users == nil {
		users = newFakeUsers()
	}
	if notifier == nil {
		notifier = domain.NopNotifier{}
	}
	return NewRouter(RouterConfig{
		Users:      users,
		Pinger:     users,
		Passwords:  auth.NewTestHasher(),
		Tokens:     tokens,
		CORSOrigin: "http://localhost:5173",
		Limiter:    NewRateLimiter(100, time.Minute),
		Orders:     orders,
		Notifier:   notifier,
	}), tokens
}

func signedIn(t *testing.T, users *fakeUsers, tokens *auth.Tokens, role, email string) (domain.User, *http.Cookie) {
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
	return user, &http.Cookie{Name: sessionCookie, Value: raw}
}

func assertOrderShape(t *testing.T, order map[string]any) {
	t.Helper()
	want := []string{"allPaid", "collectedQty", "createdAt", "deadline", "disputes", "id", "moq", "participantCount", "participants", "product", "status", "yourDeliveryMethod", "yourDeliveryStatus", "yourPaymentStatus", "yourQuantity", "yourReceiptStatus", "yourShippingFee"}
	if !sameStrings(mapKeys(order), want) {
		t.Fatalf("order keys = %#v", mapKeys(order))
	}
	product, ok := order["product"].(map[string]any)
	if !ok {
		t.Fatalf("product = %#v", order["product"])
	}
	productKeys := []string{"category", "createdAt", "description", "id", "imageUrl", "name", "priceTiers", "rating", "retailPrice", "reviewCount", "supplier", "tags", "unit"}
	if !sameStrings(mapKeys(product), productKeys) {
		t.Fatalf("product keys = %#v", mapKeys(product))
	}
	if _, ok := product["retailPrice"].(float64); !ok {
		t.Fatalf("retailPrice = %#v", product["retailPrice"])
	}
	supplier, ok := product["supplier"].(map[string]any)
	if !ok || !sameStrings(mapKeys(supplier), []string{"id", "isVerified", "name", "region"}) {
		t.Fatalf("supplier = %#v", product["supplier"])
	}
	tiers, ok := product["priceTiers"].([]any)
	if !ok || len(tiers) == 0 {
		t.Fatalf("tiers = %#v", product["priceTiers"])
	}
	if _, ok := tiers[0].(map[string]any)["unitPrice"].(float64); !ok {
		t.Fatalf("unitPrice = %#v", tiers[0])
	}
	parts, ok := order["participants"].([]any)
	if !ok || len(parts) == 0 {
		t.Fatalf("participants = %#v", order["participants"])
	}
	part, ok := parts[0].(map[string]any)
	if !ok || !sameStrings(mapKeys(part), []string{"deliveryMethod", "deliveryStatus", "id", "paymentStatus", "quantity", "receiptStatus", "shippingAddress", "shippingFee"}) {
		t.Fatalf("participation = %#v", parts[0])
	}
}

func findParticipation(t *testing.T, order map[string]any, id string) map[string]any {
	t.Helper()
	parts, ok := order["participants"].([]any)
	if !ok {
		t.Fatalf("participants = %#v", order["participants"])
	}
	for _, item := range parts {
		part := item.(map[string]any)
		if part["id"] == id {
			return part
		}
	}
	t.Fatalf("participation %s missing", id)
	return nil
}

func findOrder(t *testing.T, items []any, id string) map[string]any {
	t.Helper()
	for _, item := range items {
		order := item.(map[string]any)
		if order["id"] == id {
			return order
		}
	}
	t.Fatalf("order %s missing", id)
	return nil
}

func orderIDs(items []any) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.(map[string]any)["id"].(string))
	}
	return ids
}

func intField(t *testing.T, body map[string]any, key string) int {
	t.Helper()
	value, ok := body[key].(float64)
	if !ok {
		t.Fatalf("%s = %#v", key, body[key])
	}
	return int(value)
}

func mapKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type participation struct {
	id              uuid.UUID
	userID          uuid.UUID
	quantity        int
	paymentStatus   string
	paidAt          *time.Time
	shippingAddress string
	deliveryMethod  string
	shippingFee     float64
	deliveryStatus  string
	receiptStatus   string
	shippedAt       *time.Time
	deliveredAt     *time.Time
	receivedAt      *time.Time
}

type storedOrder struct {
	id        uuid.UUID
	productID uuid.UUID
	status    string
	moq       int
	deadline  time.Time
	createdAt time.Time
	parts     []participation
}

type storedDispute struct {
	id           uuid.UUID
	orderID      uuid.UUID
	openedBy     uuid.UUID
	businessName string
	reason       string
	status       string
	outcome      string
	createdAt    time.Time
	resolvedAt   *time.Time
}

type memOrders struct {
	mu       sync.Mutex
	products []domain.Product
	orders   []*storedOrder
	disputes []storedDispute
}

func newMemOrders() *memOrders {
	catalog := sampleCatalog()
	return &memOrders{products: append([]domain.Product(nil), catalog.products...)}
}

func (m *memOrders) seed(order storedOrder) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := order
	copied.parts = append([]participation(nil), order.parts...)
	for i := range copied.parts {
		if copied.parts[i].id == uuid.Nil {
			copied.parts[i].id = uuid.New()
		}
		if copied.parts[i].paymentStatus == "" {
			copied.parts[i].paymentStatus = domain.PaymentUnpaid
		}
		if copied.parts[i].deliveryStatus == "" {
			copied.parts[i].deliveryStatus = domain.DeliveryPending
		}
		if copied.parts[i].receiptStatus == "" {
			copied.parts[i].receiptStatus = domain.ReceiptPending
		}
	}
	m.orders = append(m.orders, &copied)
}

func (m *memOrders) ListOrders(ctx context.Context, status string, viewerID *uuid.UUID) ([]domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	for _, order := range m.orders {
		m.settle(ctx, order, now)
	}
	listed := make([]domain.Order, 0)
	for _, order := range m.orders {
		if status != "" && order.status != status {
			continue
		}
		view, err := m.view(order, viewerID)
		if err != nil {
			return nil, err
		}
		listed = append(listed, view)
	}
	sort.Slice(listed, func(i, j int) bool {
		if !listed[i].CreatedAt.Equal(listed[j].CreatedAt) {
			return listed[i].CreatedAt.After(listed[j].CreatedAt)
		}
		return listed[i].ID.String() > listed[j].ID.String()
	})
	return listed, nil
}

func (m *memOrders) FindOrder(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := m.find(id)
	if order == nil {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	return m.view(order, viewerID)
}

func (m *memOrders) CreateOrder(ctx context.Context, input domain.NewOrder) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.visibleProduct(input.ProductID, input.IncludeUnverified); !ok {
		return domain.Order{}, domain.ErrProductUnavailable
	}
	status := domain.OrderOpen
	if input.Quantity >= input.MOQ {
		status = domain.OrderConfirmed
	}
	order := &storedOrder{
		id:        uuid.New(),
		productID: input.ProductID,
		status:    status,
		moq:       input.MOQ,
		deadline:  input.Deadline,
		createdAt: time.Now().UTC(),
		parts: []participation{{
			id: uuid.New(), userID: input.CreatorID, quantity: input.Quantity,
			paymentStatus: domain.PaymentUnpaid, deliveryStatus: domain.DeliveryPending,
		}},
	}
	m.orders = append(m.orders, order)
	if status == domain.OrderConfirmed {
		domain.AddNotice(ctx, domain.Notice{
			Kind:    domain.NoticeOrderConfirmed,
			OrderID: order.id,
			UserIDs: []uuid.UUID{input.CreatorID},
		})
	}
	return m.view(order, &input.CreatorID)
}

func (m *memOrders) JoinOrder(ctx context.Context, orderID, userID uuid.UUID, quantity int) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := m.find(orderID)
	if order == nil {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	if order.status != domain.OrderOpen {
		return domain.Order{}, domain.ErrOrderClosed
	}
	for _, part := range order.parts {
		if part.userID == userID {
			return domain.Order{}, domain.ErrAlreadyJoined
		}
	}
	order.parts = append(order.parts, participation{
		id: uuid.New(), userID: userID, quantity: quantity,
		paymentStatus: domain.PaymentUnpaid, deliveryStatus: domain.DeliveryPending,
	})
	if collectedQuantity(order) >= order.moq {
		order.status = domain.OrderConfirmed
		domain.AddNotice(ctx, domain.Notice{
			Kind:    domain.NoticeOrderConfirmed,
			OrderID: order.id,
			UserIDs: participantIDs(order),
		})
	}
	return m.view(order, &userID)
}

func (m *memOrders) PayOrder(ctx context.Context, orderID, userID uuid.UUID, shippingAddress, deliveryMethod string) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := m.find(orderID)
	if order == nil {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	if order.status != domain.OrderConfirmed {
		return domain.Order{}, domain.ErrNotReadyForPayment
	}
	index := -1
	for i, part := range order.parts {
		if part.userID == userID {
			index = i
			break
		}
	}
	if index < 0 {
		return domain.Order{}, domain.ErrNotParticipant
	}
	if participationStatus(order.parts[index]) == domain.PaymentPaid {
		return domain.Order{}, domain.ErrAlreadyPaid
	}
	fee, ok := domain.DeliveryFee(deliveryMethod, order.parts[index].quantity)
	if !ok {
		return domain.Order{}, domain.ErrUnknownDelivery
	}
	now := time.Now().UTC()
	order.parts[index].paymentStatus = domain.PaymentPaid
	order.parts[index].paidAt = &now
	order.parts[index].shippingAddress = strings.TrimSpace(shippingAddress)
	order.parts[index].deliveryMethod = deliveryMethod
	order.parts[index].shippingFee = fee
	if allParticipationsPaid(order) {
		order.status = domain.OrderPaid
	}
	domain.AddNotice(ctx, domain.Notice{
		Kind:    domain.NoticePaymentReceived,
		OrderID: order.id,
		UserIDs: []uuid.UUID{userID},
	})
	return m.view(order, &userID)
}

func (m *memOrders) CancelOrder(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := m.find(id)
	if order == nil {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	if order.status != domain.OrderOpen && order.status != domain.OrderConfirmed {
		return domain.Order{}, domain.ErrOrderNotCancellable
	}
	order.status = domain.OrderCancelled
	domain.AddNotice(ctx, domain.Notice{
		Kind:    domain.NoticeOrderCancelled,
		OrderID: order.id,
		UserIDs: participantIDs(order),
	})
	return m.view(order, viewerID)
}

func (m *memOrders) ShipParticipation(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	return m.advanceDelivery(ctx, id, viewerID, domain.DeliveryShipped)
}

func (m *memOrders) DeliverParticipation(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (domain.Order, error) {
	return m.advanceDelivery(ctx, id, viewerID, domain.DeliveryDelivered)
}

func (m *memOrders) ReceiveOrder(ctx context.Context, orderID, userID uuid.UUID) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := m.find(orderID)
	if order == nil {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	index := -1
	for i, part := range order.parts {
		if part.userID == userID {
			index = i
			break
		}
	}
	if index < 0 {
		return domain.Order{}, domain.ErrNotParticipant
	}
	part := &order.parts[index]
	if !domain.ReceiptAllowed(order.status, participationStatus(*part), deliveryStatusOf(*part), receiptStatusOf(*part)) {
		return domain.Order{}, domain.ErrReceiptNotAllowed
	}
	now := time.Now().UTC()
	part.receiptStatus = domain.ReceiptReceived
	part.receivedAt = &now
	if allParticipationsConfirmed(order) {
		order.status = domain.OrderCompleted
	}
	return m.view(order, &userID)
}

func (m *memOrders) advanceDelivery(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID, next string) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order, index := m.findParticipation(id)
	if order == nil || index < 0 {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	part := &order.parts[index]
	if !domain.ShipmentUpdateAllowed(order.status, participationStatus(*part), deliveryStatusOf(*part), next) {
		return domain.Order{}, domain.ErrShipmentNotUpdatable
	}
	now := time.Now().UTC()
	kind := domain.NoticeShipmentSent
	if next == domain.DeliveryShipped {
		part.deliveryStatus = domain.DeliveryShipped
		part.shippedAt = &now
	} else {
		part.deliveryStatus = domain.DeliveryDelivered
		if part.shippedAt == nil {
			part.shippedAt = &now
		}
		part.deliveredAt = &now
		if allParticipationsConfirmed(order) {
			order.status = domain.OrderCompleted
		}
		kind = domain.NoticeShipmentDelivered
	}
	domain.AddNotice(ctx, domain.Notice{
		Kind:    kind,
		OrderID: order.id,
		UserIDs: []uuid.UUID{part.userID},
	})
	return m.view(order, viewerID)
}

func (m *memOrders) settle(ctx context.Context, order *storedOrder, now time.Time) {
	next, changed := domain.NextOrderStatus(order.status, order.deadline, collectedQuantity(order), order.moq, now)
	if !changed {
		return
	}
	order.status = next
	domain.AddNotice(ctx, domain.Notice{
		Kind:    domain.NoticeKindForStatus(next),
		OrderID: order.id,
		UserIDs: participantIDs(order),
	})
}

func participantIDs(order *storedOrder) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(order.parts))
	for _, part := range order.parts {
		ids = append(ids, part.userID)
	}
	return ids
}

func (m *memOrders) find(id uuid.UUID) *storedOrder {
	for _, order := range m.orders {
		if order.id == id {
			return order
		}
	}
	return nil
}

func (m *memOrders) findParticipation(id uuid.UUID) (*storedOrder, int) {
	for _, order := range m.orders {
		for index := range order.parts {
			if order.parts[index].id == id {
				return order, index
			}
		}
	}
	return nil, -1
}

func (m *memOrders) view(order *storedOrder, viewerID *uuid.UUID) (domain.Order, error) {
	product, ok := m.productByID(order.productID)
	if !ok {
		return domain.Order{}, domain.ErrNotFound
	}
	disputes := make([]domain.Dispute, 0)
	for _, dispute := range m.disputes {
		if dispute.orderID != order.id {
			continue
		}
		disputes = append(disputes, domain.Dispute{
			ID:           dispute.id,
			OrderID:      dispute.orderID,
			OpenedBy:     dispute.openedBy,
			BusinessName: dispute.businessName,
			Reason:       dispute.reason,
			Status:       dispute.status,
			Outcome:      dispute.outcome,
			CreatedAt:    dispute.createdAt,
			ResolvedAt:   dispute.resolvedAt,
		})
	}
	participants := make([]domain.Participation, 0, len(order.parts))
	var yours *int
	var payment *string
	var delivery *string
	var method *string
	var fee *float64
	var receipt *string
	allPaid := len(order.parts) > 0
	for _, part := range order.parts {
		status := participationStatus(part)
		shipped := deliveryStatusOf(part)
		participants = append(participants, domain.Participation{
			ID:              part.id,
			Quantity:        part.quantity,
			ShippingAddress: part.shippingAddress,
			DeliveryMethod:  part.deliveryMethod,
			ShippingFee:     part.shippingFee,
			PaymentStatus:   status,
			DeliveryStatus:  shipped,
			ReceiptStatus:   receiptStatusOf(part),
		})
		if status != domain.PaymentPaid {
			allPaid = false
		}
		if viewerID != nil && part.userID == *viewerID {
			quantity := part.quantity
			yours = &quantity
			statusCopy := status
			payment = &statusCopy
			deliveryCopy := shipped
			delivery = &deliveryCopy
			if part.deliveryMethod != "" {
				methodCopy := part.deliveryMethod
				feeCopy := part.shippingFee
				method = &methodCopy
				fee = &feeCopy
			}
			receiptCopy := receiptStatusOf(part)
			receipt = &receiptCopy
		}
	}
	return domain.Order{
		ID:                 order.id,
		ProductID:          order.productID,
		Status:             order.status,
		MOQ:                order.moq,
		Deadline:           order.deadline,
		CreatedAt:          order.createdAt,
		CollectedQty:       collectedQuantity(order),
		ParticipantCount:   len(order.parts),
		YourQuantity:       yours,
		YourPaymentStatus:  payment,
		YourDeliveryStatus: delivery,
		YourDeliveryMethod: method,
		YourShippingFee:    fee,
		YourReceiptStatus:  receipt,
		AllPaid:            allPaid,
		Participants:       participants,
		Disputes:           disputes,
		Product:            product,
	}, nil
}

func (m *memOrders) OpenDispute(ctx context.Context, orderID, userID uuid.UUID, reason string) (domain.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := m.find(orderID)
	if order == nil {
		return domain.Order{}, domain.ErrNotFound
	}
	m.settle(ctx, order, time.Now().UTC())
	if !domain.DisputeAllowed(order.status) {
		return domain.Order{}, domain.ErrDisputeNotAllowed
	}
	member := false
	for _, part := range order.parts {
		if part.userID == userID {
			member = true
			break
		}
	}
	if !member {
		return domain.Order{}, domain.ErrNotParticipant
	}
	for _, dispute := range m.disputes {
		if dispute.orderID == orderID && dispute.openedBy == userID && dispute.status == domain.DisputeOpen {
			return domain.Order{}, domain.ErrDisputeOpen
		}
	}
	m.disputes = append(m.disputes, storedDispute{
		id:           uuid.New(),
		orderID:      orderID,
		openedBy:     userID,
		businessName: "Business",
		reason:       reason,
		status:       domain.DisputeOpen,
		createdAt:    time.Now().UTC(),
	})
	domain.AddNotice(ctx, domain.Notice{
		Kind:    domain.NoticeDisputeOpened,
		OrderID: orderID,
		UserIDs: participantIDs(order),
	})
	return m.view(order, &userID)
}

func (m *memOrders) ResolveDispute(ctx context.Context, disputeID uuid.UUID, outcome string) (domain.AdminDispute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.disputes {
		if m.disputes[i].id != disputeID {
			continue
		}
		if m.disputes[i].status != domain.DisputeOpen {
			return domain.AdminDispute{}, domain.ErrDisputeClosed
		}
		now := time.Now().UTC()
		m.disputes[i].status = domain.DisputeResolved
		m.disputes[i].outcome = outcome
		m.disputes[i].resolvedAt = &now
		order := m.find(m.disputes[i].orderID)
		productName := ""
		orderStatus := ""
		if order != nil {
			orderStatus = order.status
			if product, ok := m.productByID(order.productID); ok {
				productName = product.Name
			}
		}
		domain.AddNotice(ctx, domain.Notice{
			Kind:    domain.NoticeDisputeResolved,
			OrderID: m.disputes[i].orderID,
			UserIDs: []uuid.UUID{m.disputes[i].openedBy},
		})
		return domain.AdminDispute{
			Dispute: domain.Dispute{
				ID:           m.disputes[i].id,
				OrderID:      m.disputes[i].orderID,
				OpenedBy:     m.disputes[i].openedBy,
				BusinessName: m.disputes[i].businessName,
				Reason:       m.disputes[i].reason,
				Status:       m.disputes[i].status,
				Outcome:      m.disputes[i].outcome,
				CreatedAt:    m.disputes[i].createdAt,
				ResolvedAt:   m.disputes[i].resolvedAt,
			},
			ProductName: productName,
			OrderStatus: orderStatus,
		}, nil
	}
	return domain.AdminDispute{}, domain.ErrNotFound
}

func (m *memOrders) ListDisputes(context.Context) ([]domain.AdminDispute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]domain.AdminDispute, 0, len(m.disputes))
	for _, dispute := range m.disputes {
		order := m.find(dispute.orderID)
		productName := ""
		orderStatus := ""
		if order != nil {
			orderStatus = order.status
			if product, ok := m.productByID(order.productID); ok {
				productName = product.Name
			}
		}
		items = append(items, domain.AdminDispute{
			Dispute: domain.Dispute{
				ID:           dispute.id,
				OrderID:      dispute.orderID,
				OpenedBy:     dispute.openedBy,
				BusinessName: dispute.businessName,
				Reason:       dispute.reason,
				Status:       dispute.status,
				Outcome:      dispute.outcome,
				CreatedAt:    dispute.createdAt,
				ResolvedAt:   dispute.resolvedAt,
			},
			ProductName: productName,
			OrderStatus: orderStatus,
		})
	}
	return items, nil
}

func (m *memOrders) visibleProduct(id uuid.UUID, includeUnverified bool) (domain.Product, bool) {
	product, ok := m.productByID(id)
	if !ok || (!includeUnverified && !product.Supplier.IsVerified) {
		return domain.Product{}, false
	}
	return product, true
}

func (m *memOrders) productByID(id uuid.UUID) (domain.Product, bool) {
	for _, product := range m.products {
		if product.ID == id {
			return product, true
		}
	}
	return domain.Product{}, false
}

func participationStatus(part participation) string {
	if part.paymentStatus == "" {
		return domain.PaymentUnpaid
	}
	return part.paymentStatus
}

func deliveryStatusOf(part participation) string {
	if part.deliveryStatus == "" {
		return domain.DeliveryPending
	}
	return part.deliveryStatus
}

func allParticipationsConfirmed(order *storedOrder) bool {
	if len(order.parts) == 0 {
		return false
	}
	for _, part := range order.parts {
		if deliveryStatusOf(part) != domain.DeliveryDelivered || receiptStatusOf(part) != domain.ReceiptReceived {
			return false
		}
	}
	return true
}

func receiptStatusOf(part participation) string {
	if part.receiptStatus == "" {
		return domain.ReceiptPending
	}
	return part.receiptStatus
}

func allParticipationsPaid(order *storedOrder) bool {
	if len(order.parts) == 0 {
		return false
	}
	for _, part := range order.parts {
		if participationStatus(part) != domain.PaymentPaid {
			return false
		}
	}
	return true
}

func savedAddress(t *testing.T, orders *memOrders, orderID, userID uuid.UUID) string {
	t.Helper()
	orders.mu.Lock()
	defer orders.mu.Unlock()
	order := orders.find(orderID)
	if order == nil {
		t.Fatalf("order %s missing", orderID)
	}
	for _, part := range order.parts {
		if part.userID == userID {
			return part.shippingAddress
		}
	}
	t.Fatalf("participation %s missing", userID)
	return ""
}

type captureNotifier struct {
	mu    sync.Mutex
	items []domain.Notice
}

func (c *captureNotifier) Notify(_ context.Context, notice domain.Notice) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	notice.UserIDs = append([]uuid.UUID(nil), notice.UserIDs...)
	c.items = append(c.items, notice)
	return nil
}

func (c *captureNotifier) list() []domain.Notice {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]domain.Notice(nil), c.items...)
}

func TestOpenAndResolveDispute(t *testing.T) {
	orders := newMemOrders()
	users := newFakeUsers()
	router, tokens := adminRouter(nil, orders, users, nil)
	payer, payerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "payer-dispute@harbor.example")
	_, outsiderCookie := signedIn(t, users, tokens, domain.RoleBusiness, "outsider-dispute@harbor.example")
	_, adminCookie := signedIn(t, users, tokens, domain.RoleAdmin, "admin-dispute@harbor.example")
	future := time.Date(2099, 4, 1, 23, 59, 59, 0, time.UTC)
	confirmedID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc1")
	openID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc2")
	orders.seed(storedOrder{
		id: confirmedID, productID: towelID, status: domain.OrderConfirmed, moq: 10, deadline: future,
		createdAt: time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: payer.ID, quantity: 10}},
	})
	orders.seed(storedOrder{
		id: openID, productID: towelID, status: domain.OrderOpen, moq: 20, deadline: future,
		createdAt: time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: payer.ID, quantity: 2}},
	})

	early, _ := request(t, router, http.MethodPost, "/orders/"+openID.String()+"/disputes", map[string]any{"reason": "The towels arrived torn"}, payerCookie, http.StatusBadRequest)
	if early["message"] != "This order cannot be disputed" {
		t.Fatalf("open = %#v", early)
	}
	outsider, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/disputes", map[string]any{"reason": "The towels arrived torn"}, outsiderCookie, http.StatusForbidden)
	if outsider["message"] != "You are not part of this order" {
		t.Fatalf("outsider = %#v", outsider)
	}
	short, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/disputes", map[string]any{"reason": "short"}, payerCookie, http.StatusUnprocessableEntity)
	if short["fieldErrors"].(map[string]any)["reason"] != "Use at least 8 characters" {
		t.Fatalf("short = %#v", short)
	}

	opened, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/disputes", map[string]any{"reason": "The towels arrived torn"}, payerCookie, http.StatusCreated)
	rows := opened["disputes"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["status"] != domain.DisputeOpen || rows[0].(map[string]any)["reason"] != "The towels arrived torn" {
		t.Fatalf("opened = %#v", opened["disputes"])
	}
	again, _ := request(t, router, http.MethodPost, "/orders/"+confirmedID.String()+"/disputes", map[string]any{"reason": "The towels arrived torn"}, payerCookie, http.StatusConflict)
	if again["message"] != "You already have an open dispute on this order" {
		t.Fatalf("again = %#v", again)
	}

	disputeID := rows[0].(map[string]any)["id"].(string)
	blank, _ := request(t, router, http.MethodPost, "/admin/disputes/"+disputeID+"/resolve", map[string]any{"outcome": "  "}, adminCookie, http.StatusUnprocessableEntity)
	if blank["fieldErrors"].(map[string]any)["outcome"] != "Record the outcome" {
		t.Fatalf("blank = %#v", blank)
	}
	resolved, _ := request(t, router, http.MethodPost, "/admin/disputes/"+disputeID+"/resolve", map[string]any{"outcome": "  Replacement sent  "}, adminCookie, http.StatusOK)
	if resolved["status"] != domain.DisputeResolved || resolved["outcome"] != "Replacement sent" || resolved["productName"] == "" {
		t.Fatalf("resolved = %#v", resolved)
	}
	repeat, _ := request(t, router, http.MethodPost, "/admin/disputes/"+disputeID+"/resolve", map[string]any{"outcome": "Again"}, adminCookie, http.StatusBadRequest)
	if repeat["message"] != "This dispute is already resolved" {
		t.Fatalf("repeat = %#v", repeat)
	}
	listed, _ := requestList(t, router, "/admin/disputes", adminCookie, http.StatusOK)
	if len(listed) != 1 || listed[0].(map[string]any)["id"] != disputeID {
		t.Fatalf("listed = %#v", listed)
	}
}

func collectedQuantity(order *storedOrder) int {
	total := 0
	for _, part := range order.parts {
		total += part.quantity
	}
	return total
}
