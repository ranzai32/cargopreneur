package httpapi

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

func TestAdminRoutesRequireAdmin(t *testing.T) {
	users := newFakeUsers()
	router, tokens := adminRouter(sampleCatalog(), newMemOrders(), users, nil)
	_, business := signedIn(t, users, tokens, domain.RoleBusiness, "business-admin@harbor.example")
	missing := uuid.NewString()
	calls := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPost, "/admin/suppliers", map[string]any{"name": "Northwind", "region": "OH"}},
		{http.MethodPost, "/admin/suppliers/" + missing + "/verify", nil},
		{http.MethodPost, "/admin/products", map[string]any{"name": "Gloves"}},
		{http.MethodGet, "/admin/suppliers", nil},
		{http.MethodGet, "/admin/orders", nil},
		{http.MethodPost, "/admin/orders/" + missing + "/cancel", nil},
		{http.MethodPost, "/admin/participations/" + missing + "/ship", nil},
		{http.MethodPost, "/admin/participations/" + missing + "/deliver", nil},
	}
	for _, call := range calls {
		body, _ := request(t, router, call.method, call.path, call.body, nil, http.StatusUnauthorized)
		if body["message"] != "Authentication required" {
			t.Fatalf("%s %s auth = %#v", call.method, call.path, body)
		}
		forbidden, _ := request(t, router, call.method, call.path, call.body, business, http.StatusForbidden)
		if forbidden["message"] != "You do not have access to this" {
			t.Fatalf("%s %s forbidden = %#v", call.method, call.path, forbidden)
		}
	}
}

func TestAdminCreatesAndVerifiesSupplier(t *testing.T) {
	catalog := sampleCatalog()
	users := newFakeUsers()
	router, tokens := adminRouter(catalog, newMemOrders(), users, nil)
	_, admin := signedIn(t, users, tokens, domain.RoleAdmin, "admin-suppliers@harbor.example")

	invalid, _ := request(t, router, http.MethodPost, "/admin/suppliers", map[string]any{
		"name":   " ",
		"region": "",
	}, admin, http.StatusUnprocessableEntity)
	if invalid["message"] != "Check the highlighted fields" {
		t.Fatalf("message = %#v", invalid)
	}
	fields := invalid["fieldErrors"].(map[string]any)
	if fields["name"] != "Name is required" || fields["region"] != "Region is required" {
		t.Fatalf("fields = %#v", fields)
	}

	created, _ := request(t, router, http.MethodPost, "/admin/suppliers", map[string]any{
		"name":        " Northwind Goods ",
		"description": " Paper goods. ",
		"region":      " Columbus, OH ",
	}, admin, http.StatusCreated)
	assertSupplierShape(t, created)
	if created["name"] != "Northwind Goods" || created["description"] != "Paper goods." || created["region"] != "Columbus, OH" || created["isVerified"] != false || created["productCount"].(float64) != 0 {
		t.Fatalf("created = %#v", created)
	}
	supplierID := created["id"].(string)

	publicBefore, _ := requestList(t, router, "/suppliers", nil, http.StatusOK)
	if names := supplierNames(publicBefore); !sameStrings(names, []string{"Harbor Mill Supply"}) {
		t.Fatalf("public before verify = %#v", names)
	}
	listed, _ := requestList(t, router, "/admin/suppliers", admin, http.StatusOK)
	found := findSupplier(t, listed, supplierID)
	if found["isVerified"] != false {
		t.Fatalf("admin list = %#v", found)
	}
	if names := supplierNames(listed); !sameStrings(names, []string{"Harbor Mill Supply", "Northwind Goods", "QuickLot Imports"}) {
		t.Fatalf("admin suppliers = %#v", names)
	}

	verified, _ := request(t, router, http.MethodPost, "/admin/suppliers/"+supplierID+"/verify", nil, admin, http.StatusOK)
	if verified["isVerified"] != true || verified["name"] != "Northwind Goods" {
		t.Fatalf("verified = %#v", verified)
	}
	publicAfter, _ := requestList(t, router, "/suppliers", nil, http.StatusOK)
	if names := supplierNames(publicAfter); !sameStrings(names, []string{"Harbor Mill Supply", "Northwind Goods"}) {
		t.Fatalf("public after verify = %#v", names)
	}
	missing, _ := request(t, router, http.MethodPost, "/admin/suppliers/"+uuid.NewString()+"/verify", nil, admin, http.StatusNotFound)
	if missing["message"] != "Supplier not found" {
		t.Fatalf("missing = %#v", missing)
	}
}

func TestAdminProductValidation(t *testing.T) {
	users := newFakeUsers()
	router, tokens := adminRouter(sampleCatalog(), newMemOrders(), users, nil)
	_, admin := signedIn(t, users, tokens, domain.RoleAdmin, "admin-products@harbor.example")

	invalid, _ := request(t, router, http.MethodPost, "/admin/products", map[string]any{
		"supplierId":  "nope",
		"name":        " ",
		"category":    "",
		"unit":        "",
		"tags":        "gloves",
		"retailPrice": -1,
		"priceTiers":  []any{},
	}, admin, http.StatusUnprocessableEntity)
	if invalid["message"] != "Check the highlighted fields" {
		t.Fatalf("message = %#v", invalid)
	}
	fields := invalid["fieldErrors"].(map[string]any)
	if fields["supplierId"] != "Must be a valid id" || fields["name"] != "Name is required" || fields["category"] != "Category is required" || fields["unit"] != "Unit is required" || fields["tags"] != "Enter tags as a list" || fields["retailPrice"] != "Enter a retail price of 0 or more" || fields["priceTiers"] != "Add at least one price tier" {
		t.Fatalf("fields = %#v", fields)
	}

	priced, _ := request(t, router, http.MethodPost, "/admin/products", map[string]any{
		"supplierId":  harborID.String(),
		"name":        "Gloves",
		"category":    "Workshop",
		"unit":        "box",
		"retailPrice": "12",
		"priceTiers":  []any{map[string]any{"minQty": 10, "unitPrice": 4}},
	}, admin, http.StatusUnprocessableEntity)
	if priced["fieldErrors"].(map[string]any)["retailPrice"] != "Enter a retail price of 0 or more" {
		t.Fatalf("string price = %#v", priced)
	}

	unknown, _ := request(t, router, http.MethodPost, "/admin/products", map[string]any{
		"supplierId":  uuid.NewString(),
		"name":        "Gloves",
		"category":    "Workshop",
		"unit":        "box",
		"retailPrice": 10,
		"priceTiers": []any{
			map[string]any{"minQty": 0, "unitPrice": -2},
			map[string]any{"minQty": 5, "unitPrice": 1},
			map[string]any{"minQty": 5, "unitPrice": 2},
		},
	}, admin, http.StatusUnprocessableEntity)
	tierFields := unknown["fieldErrors"].(map[string]any)
	if tierFields["supplierId"] != "Choose a supplier" || tierFields["priceTiers.0.minQty"] != "Enter a minimum quantity of at least 1" || tierFields["priceTiers.0.unitPrice"] != "Enter a unit price of 0 or more" || tierFields["priceTiers"] != "Each minimum quantity can only be used once" {
		t.Fatalf("tiers = %#v", tierFields)
	}

	created, _ := request(t, router, http.MethodPost, "/admin/products", map[string]any{
		"supplierId":  harborID.String(),
		"name":        " Nitrile gloves ",
		"description": " Box of 100 ",
		"category":    " Workshop supplies ",
		"tags":        []string{"gloves", " safety "},
		"unit":        " box ",
		"retailPrice": 14.8,
		"priceTiers": []any{
			map[string]any{"minQty": 48, "unitPrice": 8.4},
			map[string]any{"minQty": 24, "unitPrice": 0},
		},
	}, admin, http.StatusCreated)
	assertProductShape(t, created)
	if created["name"] != "Nitrile gloves" || created["category"] != "Workshop supplies" || created["unit"] != "box" || created["retailPrice"].(float64) != 14.8 {
		t.Fatalf("product = %#v", created)
	}
	supplier := created["supplier"].(map[string]any)
	if supplier["name"] != "Harbor Mill Supply" || supplier["isVerified"] != true {
		t.Fatalf("supplier = %#v", supplier)
	}
	tiers := created["priceTiers"].([]any)
	if tiers[0].(map[string]any)["minQty"].(float64) != 24 || tiers[0].(map[string]any)["unitPrice"].(float64) != 0 || tiers[1].(map[string]any)["minQty"].(float64) != 48 {
		t.Fatalf("tiers = %#v", tiers)
	}
	tags := created["tags"].([]any)
	if len(tags) != 2 || tags[0] != "gloves" || tags[1] != "safety" {
		t.Fatalf("tags = %#v", tags)
	}

	free, _ := request(t, router, http.MethodPost, "/admin/products", map[string]any{
		"supplierId":  quickLotID.String(),
		"name":        "Sample bulbs",
		"category":    "Lighting",
		"unit":        "pack",
		"retailPrice": 0,
		"priceTiers":  []any{map[string]any{"minQty": 1, "unitPrice": 0}},
	}, admin, http.StatusCreated)
	if free["retailPrice"].(float64) != 0 || free["supplier"].(map[string]any)["isVerified"] != false {
		t.Fatalf("free = %#v", free)
	}
}

func TestAdminCancelsOpenOrder(t *testing.T) {
	orders := newMemOrders()
	sink := &noticeSink{}
	users := newFakeUsers()
	router, tokens := adminRouter(sampleCatalog(), orders, users, sink)
	adminUser, admin := signedIn(t, users, tokens, domain.RoleAdmin, "admin-orders@harbor.example")
	member := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1")
	other := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2")
	future := time.Date(2099, 1, 15, 23, 59, 59, 0, time.UTC)
	openID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc1")
	joinedID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc2")
	confirmedID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc3")
	cancelledID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc4")
	paidID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc5")
	completedID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc6")
	orders.seed(storedOrder{
		id: openID, productID: towelID, status: domain.OrderOpen, moq: 20, deadline: future,
		createdAt: time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: member, quantity: 4}},
	})
	orders.seed(storedOrder{
		id: joinedID, productID: mailerID, status: domain.OrderOpen, moq: 20, deadline: future,
		createdAt: time.Date(2026, 11, 19, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: adminUser.ID, quantity: 6}, {userID: other, quantity: 2}},
	})
	orders.seed(storedOrder{
		id: confirmedID, productID: cleanerID, status: domain.OrderConfirmed, moq: 10, deadline: future,
		createdAt: time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: member, quantity: 10}},
	})
	orders.seed(storedOrder{
		id: cancelledID, productID: towelID, status: domain.OrderCancelled, moq: 12, deadline: future,
		createdAt: time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: member, quantity: 1}},
	})
	orders.seed(storedOrder{
		id: paidID, productID: towelID, status: "paid", moq: 12, deadline: future,
		createdAt: time.Date(2026, 11, 16, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: member, quantity: 12}},
	})
	orders.seed(storedOrder{
		id: completedID, productID: towelID, status: "completed", moq: 12, deadline: future,
		createdAt: time.Date(2026, 11, 15, 15, 0, 0, 0, time.UTC),
		parts:     []participation{{userID: member, quantity: 12}},
	})

	listed, _ := requestList(t, router, "/admin/orders", admin, http.StatusOK)
	joined := findOrder(t, listed, joinedID.String())
	if intField(t, joined, "yourQuantity") != 6 {
		t.Fatalf("admin quantity = %#v", joined["yourQuantity"])
	}
	if findOrder(t, listed, openID.String())["yourQuantity"] != nil {
		t.Fatal("another participant quantity was filled for the admin")
	}

	cancelled, _ := request(t, router, http.MethodPost, "/admin/orders/"+openID.String()+"/cancel", nil, admin, http.StatusOK)
	assertOrderShape(t, cancelled)
	if cancelled["status"] != domain.OrderCancelled || cancelled["yourQuantity"] != nil || intField(t, cancelled, "collectedQty") != 4 {
		t.Fatalf("cancelled = %#v", cancelled)
	}
	if len(sink.items) != 1 || sink.items[0].Kind != domain.NoticeOrderCancelled || sink.items[0].OrderID != openID || len(sink.items[0].UserIDs) != 1 || sink.items[0].UserIDs[0] != member {
		t.Fatalf("notice = %#v", sink.items)
	}

	again, _ := request(t, router, http.MethodPost, "/admin/orders/"+openID.String()+"/cancel", nil, admin, http.StatusBadRequest)
	if again["message"] != "This order can no longer be cancelled" || len(sink.items) != 1 {
		t.Fatalf("again = %#v notices = %d", again, len(sink.items))
	}

	confirmed, _ := request(t, router, http.MethodPost, "/admin/orders/"+confirmedID.String()+"/cancel", nil, admin, http.StatusOK)
	if confirmed["status"] != domain.OrderCancelled || len(sink.items) != 2 || sink.items[1].Kind != domain.NoticeOrderCancelled || sink.items[1].OrderID != confirmedID {
		t.Fatalf("confirmed cancel = %#v notices = %#v", confirmed, sink.items)
	}

	for _, id := range []uuid.UUID{cancelledID, paidID, completedID} {
		refused, _ := request(t, router, http.MethodPost, "/admin/orders/"+id.String()+"/cancel", nil, admin, http.StatusBadRequest)
		if refused["message"] != "This order can no longer be cancelled" {
			t.Fatalf("refuse %s = %#v", id, refused)
		}
	}
	if len(sink.items) != 2 {
		t.Fatalf("extra notices = %#v", sink.items)
	}
	missing, _ := request(t, router, http.MethodPost, "/admin/orders/"+uuid.NewString()+"/cancel", nil, admin, http.StatusNotFound)
	if missing["message"] != "Order not found" {
		t.Fatalf("missing = %#v", missing)
	}
}

func TestAdminTracksDelivery(t *testing.T) {
	orders := newMemOrders()
	notices := &captureNotifier{}
	users := newFakeUsers()
	router, tokens := adminRouter(sampleCatalog(), orders, users, notices)
	payer, payerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "ship-payer@harbor.example")
	partner, partnerCookie := signedIn(t, users, tokens, domain.RoleBusiness, "ship-partner@harbor.example")
	_, businessCookie := signedIn(t, users, tokens, domain.RoleBusiness, "ship-business@harbor.example")
	_, adminCookie := signedIn(t, users, tokens, domain.RoleAdmin, "ship-admin@harbor.example")

	future := time.Date(2099, 4, 1, 23, 59, 59, 0, time.UTC)
	paidID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd1")
	earlyID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd2")
	unpaidOrderID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd3")
	payerPart := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1")
	partnerPart := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2")
	earlyPart := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3")
	unpaidPart := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee4")
	orders.seed(storedOrder{
		id: paidID, productID: towelID, status: domain.OrderPaid, moq: 10, deadline: future,
		createdAt: time.Date(2026, 11, 21, 15, 0, 0, 0, time.UTC),
		parts: []participation{
			{id: payerPart, userID: payer.ID, quantity: 6, paymentStatus: domain.PaymentPaid, shippingAddress: "12 Dock Street", deliveryStatus: domain.DeliveryPending},
			{id: partnerPart, userID: partner.ID, quantity: 4, paymentStatus: domain.PaymentPaid, shippingAddress: "4 Market Street", deliveryStatus: domain.DeliveryPending},
		},
	})
	orders.seed(storedOrder{
		id: earlyID, productID: mailerID, status: domain.OrderConfirmed, moq: 10, deadline: future,
		createdAt: time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC),
		parts: []participation{
			{id: earlyPart, userID: payer.ID, quantity: 6, paymentStatus: domain.PaymentPaid, shippingAddress: "12 Dock Street", deliveryStatus: domain.DeliveryPending},
		},
	})
	orders.seed(storedOrder{
		id: unpaidOrderID, productID: cleanerID, status: domain.OrderPaid, moq: 8, deadline: future,
		createdAt: time.Date(2026, 11, 19, 15, 0, 0, 0, time.UTC),
		parts: []participation{
			{id: unpaidPart, userID: partner.ID, quantity: 8, paymentStatus: domain.PaymentUnpaid, deliveryStatus: domain.DeliveryPending},
		},
	})

	forbidden, _ := request(t, router, http.MethodPost, "/admin/participations/"+payerPart.String()+"/ship", nil, businessCookie, http.StatusForbidden)
	if forbidden["message"] != "You do not have access to this" {
		t.Fatalf("forbidden = %#v", forbidden)
	}
	if len(notices.list()) != 0 {
		t.Fatalf("notices after forbidden = %#v", notices.list())
	}

	earlyDeliver, _ := request(t, router, http.MethodPost, "/admin/participations/"+payerPart.String()+"/deliver", nil, adminCookie, http.StatusBadRequest)
	if earlyDeliver["message"] != "This shipment cannot be updated" {
		t.Fatalf("deliver before ship = %#v", earlyDeliver)
	}
	earlyShip, _ := request(t, router, http.MethodPost, "/admin/participations/"+earlyPart.String()+"/ship", nil, adminCookie, http.StatusBadRequest)
	if earlyShip["message"] != "This shipment cannot be updated" {
		t.Fatalf("ship before paid order = %#v", earlyShip)
	}
	unpaidShip, _ := request(t, router, http.MethodPost, "/admin/participations/"+unpaidPart.String()+"/ship", nil, adminCookie, http.StatusBadRequest)
	if unpaidShip["message"] != "This shipment cannot be updated" {
		t.Fatalf("unpaid ship = %#v", unpaidShip)
	}
	if len(notices.list()) != 0 {
		t.Fatalf("notices after rejected ships = %#v", notices.list())
	}

	shipped, _ := request(t, router, http.MethodPost, "/admin/participations/"+payerPart.String()+"/ship", nil, adminCookie, http.StatusOK)
	assertOrderShape(t, shipped)
	if shipped["status"] != domain.OrderPaid || shipped["yourDeliveryStatus"] != nil || shipped["allPaid"] != true {
		t.Fatalf("shipped order = %#v", shipped)
	}
	payerShipment := findParticipation(t, shipped, payerPart.String())
	if payerShipment["deliveryStatus"] != domain.DeliveryShipped || payerShipment["shippingAddress"] != "12 Dock Street" || payerShipment["paymentStatus"] != domain.PaymentPaid || intField(t, payerShipment, "quantity") != 6 {
		t.Fatalf("payer shipment = %#v", payerShipment)
	}
	if findParticipation(t, shipped, partnerPart.String())["deliveryStatus"] != domain.DeliveryPending {
		t.Fatalf("partner still pending = %#v", shipped["participants"])
	}
	recorded := notices.list()
	if len(recorded) != 1 || recorded[0].Kind != domain.NoticeShipmentSent || recorded[0].OrderID != paidID || len(recorded[0].UserIDs) != 1 || recorded[0].UserIDs[0] != payer.ID {
		t.Fatalf("ship notice = %#v", recorded)
	}

	payerView, _ := request(t, router, http.MethodGet, "/orders/"+paidID.String(), nil, payerCookie, http.StatusOK)
	if payerView["yourDeliveryStatus"] != domain.DeliveryShipped || payerView["yourPaymentStatus"] != domain.PaymentPaid {
		t.Fatalf("payer view = %#v", payerView)
	}
	partnerView, _ := request(t, router, http.MethodGet, "/orders/"+paidID.String(), nil, partnerCookie, http.StatusOK)
	if partnerView["yourDeliveryStatus"] != domain.DeliveryPending {
		t.Fatalf("partner view = %#v", partnerView)
	}
	strangerView, _ := request(t, router, http.MethodGet, "/orders/"+paidID.String(), nil, businessCookie, http.StatusOK)
	if strangerView["yourDeliveryStatus"] != nil || strangerView["yourPaymentStatus"] != nil {
		t.Fatalf("stranger view = %#v", strangerView)
	}

	again, _ := request(t, router, http.MethodPost, "/admin/participations/"+payerPart.String()+"/ship", nil, adminCookie, http.StatusBadRequest)
	if again["message"] != "This shipment cannot be updated" || len(notices.list()) != 1 {
		t.Fatalf("second ship = %#v notices = %#v", again, notices.list())
	}

	delivered, _ := request(t, router, http.MethodPost, "/admin/participations/"+payerPart.String()+"/deliver", nil, adminCookie, http.StatusOK)
	assertOrderShape(t, delivered)
	if delivered["status"] != domain.OrderPaid || findParticipation(t, delivered, payerPart.String())["deliveryStatus"] != domain.DeliveryDelivered {
		t.Fatalf("partial delivery = %#v", delivered)
	}
	recorded = notices.list()
	if len(recorded) != 2 || recorded[1].Kind != domain.NoticeShipmentDelivered || recorded[1].OrderID != paidID || len(recorded[1].UserIDs) != 1 || recorded[1].UserIDs[0] != payer.ID {
		t.Fatalf("deliver notice = %#v", recorded)
	}
	repeatDeliver, _ := request(t, router, http.MethodPost, "/admin/participations/"+payerPart.String()+"/deliver", nil, adminCookie, http.StatusBadRequest)
	if repeatDeliver["message"] != "This shipment cannot be updated" || len(notices.list()) != 2 {
		t.Fatalf("second deliver = %#v", repeatDeliver)
	}

	partnerShipped, _ := request(t, router, http.MethodPost, "/admin/participations/"+partnerPart.String()+"/ship", nil, adminCookie, http.StatusOK)
	if partnerShipped["status"] != domain.OrderPaid || findParticipation(t, partnerShipped, partnerPart.String())["deliveryStatus"] != domain.DeliveryShipped {
		t.Fatalf("partner ship = %#v", partnerShipped)
	}
	completed, _ := request(t, router, http.MethodPost, "/admin/participations/"+partnerPart.String()+"/deliver", nil, adminCookie, http.StatusOK)
	assertOrderShape(t, completed)
	if completed["status"] != domain.OrderCompleted || findParticipation(t, completed, partnerPart.String())["deliveryStatus"] != domain.DeliveryDelivered || findParticipation(t, completed, payerPart.String())["deliveryStatus"] != domain.DeliveryDelivered {
		t.Fatalf("completed = %#v", completed)
	}
	recorded = notices.list()
	if len(recorded) != 4 || recorded[2].Kind != domain.NoticeShipmentSent || recorded[2].UserIDs[0] != partner.ID || recorded[3].Kind != domain.NoticeShipmentDelivered || recorded[3].UserIDs[0] != partner.ID {
		t.Fatalf("notices = %#v", recorded)
	}

	missing, _ := request(t, router, http.MethodPost, "/admin/participations/"+uuid.NewString()+"/ship", nil, adminCookie, http.StatusNotFound)
	if missing["message"] != "Participation not found" {
		t.Fatalf("missing = %#v", missing)
	}
	listed, _ := requestList(t, router, "/admin/orders", adminCookie, http.StatusOK)
	adminPaid := findOrder(t, listed, unpaidOrderID.String())
	if findParticipation(t, adminPaid, unpaidPart.String())["deliveryStatus"] != domain.DeliveryPending || findOrder(t, listed, earlyID.String())["status"] != domain.OrderConfirmed {
		t.Fatalf("unchanged orders = %#v", listed)
	}
}

func adminRouter(catalog *memCatalog, orders *memOrders, users *fakeUsers, notifier domain.Notifier) (http.Handler, *auth.Tokens) {
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
		Catalog:    catalog,
		Orders:     orders,
		Notifier:   notifier,
	}), tokens
}

func assertSupplierShape(t *testing.T, supplier map[string]any) {
	t.Helper()
	if !sameStrings(mapKeys(supplier), []string{"description", "id", "isVerified", "name", "productCount", "region"}) {
		t.Fatalf("supplier keys = %#v", mapKeys(supplier))
	}
}

func assertProductShape(t *testing.T, product map[string]any) {
	t.Helper()
	if !sameStrings(mapKeys(product), []string{"category", "createdAt", "description", "id", "imageUrl", "name", "priceTiers", "retailPrice", "supplier", "tags", "unit"}) {
		t.Fatalf("product keys = %#v", mapKeys(product))
	}
	supplier, ok := product["supplier"].(map[string]any)
	if !ok || !sameStrings(mapKeys(supplier), []string{"id", "isVerified", "name", "region"}) {
		t.Fatalf("supplier = %#v", product["supplier"])
	}
}

func findSupplier(t *testing.T, items []any, id string) map[string]any {
	t.Helper()
	for _, item := range items {
		supplier := item.(map[string]any)
		if supplier["id"] == id {
			return supplier
		}
	}
	t.Fatalf("supplier %s missing", id)
	return nil
}

type noticeSink struct {
	mu    sync.Mutex
	items []domain.Notice
}

func (n *noticeSink) Notify(_ context.Context, notice domain.Notice) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	notice.UserIDs = append([]uuid.UUID(nil), notice.UserIDs...)
	n.items = append(n.items, notice)
	return nil
}
