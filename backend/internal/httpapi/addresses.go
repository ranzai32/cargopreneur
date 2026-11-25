package httpapi

import (
	"log/slog"
	"net/http"

	"cargoprener/backend/internal/places"
)

func addressHandler(source places.TwoGIS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := source.Suggest(r.Context(), r.URL.Query().Get("q"))
		if err != nil {
			slog.Error("address suggest", "error", err)
			writeError(w, http.StatusBadGateway, "Address suggestions are unavailable", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}
