package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

type SupplierAPI interface {
	ListProducts(ctx context.Context, query domain.ProductQuery) ([]domain.Product, int, error)
	FindProduct(ctx context.Context, id uuid.UUID, includeUnverified bool) (domain.Product, error)
	CreateProduct(ctx context.Context, input domain.NewProduct) (domain.Product, error)
	UpdateOwnedProduct(ctx context.Context, supplierID, productID uuid.UUID, input domain.NewProduct) (domain.Product, error)
	DeleteOwnedProduct(ctx context.Context, supplierID, productID uuid.UUID) error
	ListOrdersForSupplier(ctx context.Context, supplierID uuid.UUID, viewerID *uuid.UUID) ([]domain.Order, error)
	ShipForSupplier(ctx context.Context, participationID, supplierID uuid.UUID, viewerID *uuid.UUID) (domain.Order, error)
	DeliverForSupplier(ctx context.Context, participationID, supplierID uuid.UUID, viewerID *uuid.UUID) (domain.Order, error)
	ListSupplierReviews(ctx context.Context, supplierID uuid.UUID) ([]domain.Review, error)
	ListProductReviews(ctx context.Context, productID uuid.UUID) ([]domain.Review, error)
	UpsertSupplierReview(ctx context.Context, supplierID, userID uuid.UUID, rating int, body string) error
	UpsertProductReview(ctx context.Context, productID, userID uuid.UUID, rating int, body string) error
}

type SupplierHandler struct {
	api       SupplierAPI
	notifier  domain.Notifier
	uploadDir string
}

func (h SupplierHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := ownedSupplier(w, r)
	if !ok {
		return
	}
	items, _, err := h.api.ListProducts(r.Context(), domain.ProductQuery{
		SupplierID:        &supplierID,
		Sort:              domain.SortName,
		Page:              1,
		PageSize:          50,
		IncludeUnverified: true,
	})
	if err != nil {
		slog.Error("list supplier products", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.ProductJSON, 0, len(items))
	for _, item := range items {
		encoded = append(encoded, item.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h SupplierHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := ownedSupplier(w, r)
	if !ok {
		return
	}
	var body domain.CreateProductBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	body.SupplierID = supplierID.String()
	input, fieldErrors := domain.ParseCreateProduct(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	product, err := h.api.CreateProduct(r.Context(), input)
	if err != nil {
		slog.Error("create supplier product", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusCreated, product.JSON())
}

func (h SupplierHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := ownedSupplier(w, r)
	if !ok {
		return
	}
	productID, ok := pathID(w, r)
	if !ok {
		return
	}
	var body domain.CreateProductBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	body.SupplierID = supplierID.String()
	input, fieldErrors := domain.ParseCreateProduct(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	product, err := h.api.UpdateOwnedProduct(r.Context(), supplierID, productID, input)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Product not found", nil)
		return
	}
	if err != nil {
		slog.Error("update supplier product", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, product.JSON())
}

func (h SupplierHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	supplierID, ok := ownedSupplier(w, r)
	if !ok {
		return
	}
	productID, ok := pathID(w, r)
	if !ok {
		return
	}
	err := h.api.DeleteOwnedProduct(r.Context(), supplierID, productID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Product not found", nil)
		return
	}
	if errors.Is(err, domain.ErrProductInUse) {
		writeError(w, http.StatusConflict, "This product is on an order", nil)
		return
	}
	if err != nil {
		slog.Error("delete supplier product", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h SupplierHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if _, ok := ownedSupplier(w, r); !ok {
		return
	}
	url, status, message := saveUpload(h.uploadDir, w, r)
	if message != "" {
		writeError(w, status, message, nil)
		return
	}
	writeJSON(w, status, map[string]string{"url": url})
}

func (h SupplierHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	supplierID, user, ok := ownedSupplierUser(w, r)
	if !ok {
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	orders, err := h.api.ListOrdersForSupplier(ctx, supplierID, &user.ID)
	dispatchNotices(ctx, h.notifier)
	if err != nil {
		slog.Error("list supplier orders", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.OrderJSON, 0, len(orders))
	for _, order := range orders {
		encoded = append(encoded, order.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h SupplierHandler) Ship(w http.ResponseWriter, r *http.Request) {
	h.markDelivery(w, r, true)
}

func (h SupplierHandler) Deliver(w http.ResponseWriter, r *http.Request) {
	h.markDelivery(w, r, false)
}

func (h SupplierHandler) markDelivery(w http.ResponseWriter, r *http.Request, ship bool) {
	supplierID, user, ok := ownedSupplierUser(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	var order domain.Order
	var err error
	if ship {
		order, err = h.api.ShipForSupplier(ctx, id, supplierID, &user.ID)
	} else {
		order, err = h.api.DeliverForSupplier(ctx, id, supplierID, &user.ID)
	}
	dispatchNotices(ctx, h.notifier)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Participation not found", nil)
		return
	}
	if errors.Is(err, domain.ErrShipmentNotUpdatable) {
		writeError(w, http.StatusBadRequest, "This shipment cannot be updated", nil)
		return
	}
	if err != nil {
		slog.Error("supplier shipment", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, order.JSON())
}

func (h SupplierHandler) SupplierReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	h.writeReviews(w, r, func() ([]domain.Review, error) {
		return h.api.ListSupplierReviews(r.Context(), id)
	})
}

func (h SupplierHandler) ProductReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	h.writeReviews(w, r, func() ([]domain.Review, error) {
		return h.api.ListProductReviews(r.Context(), id)
	})
}

func (h SupplierHandler) WriteSupplierReview(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	rating, text, ok := reviewInput(w, r)
	if !ok {
		return
	}
	err := h.api.UpsertSupplierReview(r.Context(), id, user.ID, rating, text)
	h.reviewResult(w, err)
}

func (h SupplierHandler) WriteProductReview(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	rating, text, ok := reviewInput(w, r)
	if !ok {
		return
	}
	err := h.api.UpsertProductReview(r.Context(), id, user.ID, rating, text)
	h.reviewResult(w, err)
}

func (h SupplierHandler) writeReviews(w http.ResponseWriter, r *http.Request, load func() ([]domain.Review, error)) {
	items, err := load()
	if err != nil {
		slog.Error("list reviews", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.ReviewJSON, 0, len(items))
	for _, item := range items {
		encoded = append(encoded, item.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h SupplierHandler) reviewResult(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Not found", nil)
		return
	}
	if errors.Is(err, domain.ErrNotParticipant) {
		writeError(w, http.StatusForbidden, "You cannot review your own listing", nil)
		return
	}
	if err != nil {
		slog.Error("save review", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func reviewInput(w http.ResponseWriter, r *http.Request) (int, string, bool) {
	var body domain.ReviewBody
	if err := decodeJSON(w, r, &body); err != nil {
		return 0, "", false
	}
	rating, text, fieldErrors := domain.ParseReview(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return 0, "", false
	}
	return rating, text, true
}

func ownedSupplier(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, _, ok := ownedSupplierUser(w, r)
	return id, ok
}

func ownedSupplierUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, domain.User, bool) {
	user, ok := userFrom(r.Context())
	if !ok || user.Role != domain.RoleSupplier || user.SupplierID == nil {
		writeError(w, http.StatusForbidden, "You do not have access to this", nil)
		return uuid.Nil, domain.User{}, false
	}
	return *user.SupplierID, user, true
}
