package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/domain"

	"github.com/google/uuid"
)

type AdminHandler struct {
	catalog  CatalogStore
	orders   OrderStore
	notifier domain.Notifier
}

func (h AdminHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var body domain.CreateSupplierBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	input, fieldErrors := domain.ParseCreateSupplier(body)
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	supplier, err := h.catalog.CreateSupplier(r.Context(), input)
	if err != nil {
		slog.Error("create supplier", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusCreated, supplier.JSON())
}

func (h AdminHandler) VerifySupplier(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	supplier, err := h.catalog.VerifySupplier(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Supplier not found", nil)
		return
	}
	if err != nil {
		slog.Error("verify supplier", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, supplier.JSON())
}

func (h AdminHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.catalog.ListSuppliers(r.Context(), true)
	if err != nil {
		slog.Error("list admin suppliers", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.SupplierJSON, 0, len(suppliers))
	for _, supplier := range suppliers {
		encoded = append(encoded, supplier.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body domain.CreateProductBody
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	input, fieldErrors := domain.ParseCreateProduct(body)
	if input.SupplierID != uuid.Nil {
		_, err := h.catalog.FindSupplier(r.Context(), input.SupplierID, true)
		if errors.Is(err, domain.ErrNotFound) {
			if fieldErrors == nil {
				fieldErrors = map[string]string{}
			}
			fieldErrors["supplierId"] = "Choose a supplier"
		} else if err != nil {
			slog.Error("find supplier", "error", err)
			writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
			return
		}
	}
	if len(fieldErrors) > 0 {
		writeValidation(w, fieldErrors)
		return
	}
	product, err := h.catalog.CreateProduct(r.Context(), input)
	if errors.Is(err, domain.ErrNotFound) {
		writeValidation(w, map[string]string{"supplierId": "Choose a supplier"})
		return
	}
	if err != nil {
		slog.Error("create product", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusCreated, product.JSON())
}

func (h AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	orders, err := h.orders.ListOrders(ctx, "", &user.ID)
	dispatchNotices(ctx, h.notifier)
	if err != nil {
		slog.Error("list admin orders", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.OrderJSON, 0, len(orders))
	for _, order := range orders {
		encoded = append(encoded, order.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h AdminHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ctx := domain.WithNoticeLog(r.Context())
	order, err := h.orders.CancelOrder(ctx, id, &user.ID)
	dispatchNotices(ctx, h.notifier)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Order not found", nil)
		return
	}
	if errors.Is(err, domain.ErrOrderNotCancellable) {
		writeError(w, http.StatusBadRequest, "This order can no longer be cancelled", nil)
		return
	}
	if err != nil {
		slog.Error("cancel order", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, order.JSON())
}

func (h AdminHandler) ShipParticipation(w http.ResponseWriter, r *http.Request) {
	h.updateShipment(w, r, true)
}

func (h AdminHandler) DeliverParticipation(w http.ResponseWriter, r *http.Request) {
	h.updateShipment(w, r, false)
}

func (h AdminHandler) updateShipment(w http.ResponseWriter, r *http.Request, ship bool) {
	user, ok := userFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authentication required", nil)
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
		order, err = h.orders.ShipParticipation(ctx, id, &user.ID)
	} else {
		order, err = h.orders.DeliverParticipation(ctx, id, &user.ID)
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
		slog.Error("update shipment", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, order.JSON())
}
