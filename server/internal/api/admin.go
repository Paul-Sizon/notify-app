package api

import (
	"crypto/subtle"
	"net/http"
)

// requireAdmin gates a route by `X-Admin-Token: <token>` matching the
// configured Handler.AdminToken via constant-time comparison.
func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Admin-Token")
		if h.AdminToken == "" || got == "" ||
			subtle.ConstantTimeCompare([]byte(got), []byte(h.AdminToken)) != 1 {
			writeErr(w, http.StatusUnauthorized, "admin token required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// adminNukeSubscriptions wipes every subscription across every device.
// Cascades to signals via FK. Intended for hackathon/dev resets.
func (h *Handler) adminNukeSubscriptions(w http.ResponseWriter, r *http.Request) {
	n, err := h.DB.DeleteAllSubscriptions(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": n})
}

// bulkDeleteSubscriptions wipes every subscription belonging to the
// current device (X-Device-Id). Authed via the device middleware.
func (h *Handler) bulkDeleteSubscriptions(w http.ResponseWriter, r *http.Request) {
	devID := deviceIDFromContext(r.Context())
	n, err := h.DB.DeleteSubscriptionsByDevice(r.Context(), devID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": n})
}
