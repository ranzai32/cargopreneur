package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/auth"
	"cargoprener/backend/internal/domain"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CatalogStore interface {
	ListProducts(ctx context.Context, query domain.ProductQuery) ([]domain.Product, int, error)
	FindProduct(ctx context.Context, id uuid.UUID, includeUnverified bool) (domain.Product, error)
	ListSuppliers(ctx context.Context, includeUnverified bool) ([]domain.Supplier, error)
	FindSupplier(ctx context.Context, id uuid.UUID, includeUnverified bool) (domain.Supplier, error)
	ListCategories(ctx context.Context, includeUnverified bool) ([]domain.Category, error)
}

type CatalogHandler struct {
	catalog CatalogStore
	users   UserLookup
	tokens  *auth.Tokens
}

func (h CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query, fieldErrors := domain.ParseProductQuery(r.URL.Query())
	if len(fieldErrors) > 0 {
		writeError(w, http.StatusBadRequest, "Invalid query", fieldErrors)
		return
	}
	query.IncludeUnverified = h.isAdmin(r)

	items, total, err := h.catalog.ListProducts(r.Context(), query)
	if err != nil {
		slog.Error("list products", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.ProductJSON, 0, len(items))
	for _, item := range items {
		encoded = append(encoded, item.JSON())
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":    encoded,
		"total":    total,
		"page":     query.Page,
		"pageSize": query.PageSize,
	})
}

func (h CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	product, err := h.catalog.FindProduct(r.Context(), id, h.isAdmin(r))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Product not found", nil)
		return
	}
	if err != nil {
		slog.Error("find product", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, product.JSON())
}

func (h CatalogHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	requested, message := domain.ParseIncludeUnverified(r.URL.Query().Get("includeUnverified"))
	if message != "" {
		writeError(w, http.StatusBadRequest, "Invalid query", map[string]string{"includeUnverified": message})
		return
	}
	includeUnverified := requested && h.isAdmin(r)
	suppliers, err := h.catalog.ListSuppliers(r.Context(), includeUnverified)
	if err != nil {
		slog.Error("list suppliers", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	encoded := make([]domain.SupplierJSON, 0, len(suppliers))
	for _, supplier := range suppliers {
		encoded = append(encoded, supplier.JSON())
	}
	writeJSON(w, http.StatusOK, encoded)
}

func (h CatalogHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	supplier, err := h.catalog.FindSupplier(r.Context(), id, h.isAdmin(r))
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Supplier not found", nil)
		return
	}
	if err != nil {
		slog.Error("find supplier", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	writeJSON(w, http.StatusOK, supplier.JSON())
}

func (h CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.catalog.ListCategories(r.Context(), h.isAdmin(r))
	if err != nil {
		slog.Error("list categories", "error", err)
		writeError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	if categories == nil {
		categories = []domain.Category{}
	}
	writeJSON(w, http.StatusOK, categories)
}

func (h CatalogHandler) isAdmin(r *http.Request) bool {
	if h.users == nil || h.tokens == nil {
		return false
	}
	user, ok := optionalUser(r, h.users, h.tokens)
	return ok && user.Role == domain.RoleAdmin
}

func pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid query", map[string]string{
			"id": "Must be a valid id",
		})
		return uuid.Nil, false
	}
	return id, true
}
