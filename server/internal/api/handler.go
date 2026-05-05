package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paulsizon/notify/server/internal/db"
)

// Runner triggers an agent run for a subscription. main wires the real impl.
type Runner func(ctx context.Context, subID uuid.UUID) ([]uuid.UUID, error)

type Handler struct {
	DB     *db.DB
	Runner Runner
}

func NewHandler(d *db.DB, runner Runner) *Handler {
	return &Handler{DB: d, Runner: runner}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(accessLog)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	r.Route("/v1", func(r chi.Router) {
		r.Post("/devices", h.registerDevice)

		r.Group(func(r chi.Router) {
			r.Use(h.requireDevice)
			r.Post("/subscriptions", h.createSubscription)
			r.Get("/subscriptions", h.listSubscriptions)
			r.Delete("/subscriptions/{id}", h.deleteSubscription)
			r.Get("/subscriptions/{id}/signals", h.listSignals)
			r.Post("/subscriptions/{id}/run", h.runSubscription)
		})
	})
	return r
}

// --- middleware ---

// statusRecorder wraps ResponseWriter so accessLog can read the status code
// after the handler runs. Without it, http.ResponseWriter exposes no way
// to recover the code that was written.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// accessLog emits one slog line per request — method, path, status, latency,
// remote addr. Without this, a misbehaving client (wrong device id, network
// issue, app pointing at the wrong URL) is invisible: the server only logs
// scheduler activity, not incoming HTTP. Critical for diagnosing "the app
// shows an error but the server log is silent."
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"dur_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}


type ctxKey string

const ctxKeyDeviceID ctxKey = "device_id"

func (h *Handler) requireDevice(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("X-Device-Id")
		if raw == "" {
			writeErr(w, http.StatusUnauthorized, "missing X-Device-Id")
			return
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "invalid X-Device-Id")
			return
		}
		// Verify device exists.
		if _, err := h.DB.GetDevice(r.Context(), id); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeErr(w, http.StatusUnauthorized, "unknown device")
				return
			}
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		ctx := contextWithDeviceID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contextWithDeviceID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyDeviceID, id)
}

func deviceIDFromContext(ctx context.Context) uuid.UUID {
	if v, ok := ctx.Value(ctxKeyDeviceID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// --- handlers ---

func (h *Handler) registerDevice(w http.ResponseWriter, r *http.Request) {
	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.APNsToken == "" {
		writeErr(w, http.StatusBadRequest, "apns_token required")
		return
	}
	id, err := h.DB.UpsertDevice(r.Context(), req.APNsToken)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, RegisterDeviceResponse{DeviceID: id.String()})
}

func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	devID := deviceIDFromContext(r.Context())
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateCreateSubscription(req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	sub, err := h.DB.InsertSubscription(r.Context(), db.SubscriptionInsert{
		DeviceID:       devID,
		Query:          req.Query,
		Type:           req.Type,
		CadenceSeconds: req.CadenceSeconds,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toSubDTO(sub))
}

func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	devID := deviceIDFromContext(r.Context())
	subs, err := h.DB.ListSubscriptionsByDevice(r.Context(), devID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]SubscriptionDTO, 0, len(subs))
	for _, s := range subs {
		out = append(out, toSubDTO(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	devID := deviceIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	sub, err := h.DB.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sub.DeviceID != devID {
		writeErr(w, http.StatusForbidden, "not your subscription")
		return
	}
	if err := h.DB.DeleteSubscription(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listSignals(w http.ResponseWriter, r *http.Request) {
	devID := deviceIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	sub, err := h.DB.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sub.DeviceID != devID {
		writeErr(w, http.StatusForbidden, "not your subscription")
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	before := time.Now().Add(time.Hour) // a bit in the future to include just-inserted rows
	if v := r.URL.Query().Get("before"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			before = t
		}
	}

	sigs, err := h.DB.ListSignalsBySubscription(r.Context(), id, before, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]SignalDTO, 0, len(sigs))
	for _, s := range sigs {
		out = append(out, toSignalDTO(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) runSubscription(w http.ResponseWriter, r *http.Request) {
	devID := deviceIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	sub, err := h.DB.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sub.DeviceID != devID {
		writeErr(w, http.StatusForbidden, "not your subscription")
		return
	}
	if h.Runner == nil {
		writeErr(w, http.StatusServiceUnavailable, "runner not configured")
		return
	}
	newIDs, err := h.Runner(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, RunResponse{NewSignals: len(newIDs)})
}

// --- helpers ---

func validateCreateSubscription(req CreateSubscriptionRequest) error {
	if l := len(req.Query); l < 3 || l > 200 {
		return fmt.Errorf("query length must be 3-200, got %d", l)
	}
	if req.Type != "event" && req.Type != "news" {
		return fmt.Errorf("type must be 'event' or 'news'")
	}
	if req.CadenceSeconds < 300 {
		return fmt.Errorf("cadence_seconds must be >= 300")
	}
	return nil
}

func toSubDTO(s db.Subscription) SubscriptionDTO {
	return SubscriptionDTO{
		ID:             s.ID.String(),
		Query:          s.Query,
		Type:           s.Type,
		CadenceSeconds: s.CadenceSeconds,
		LastRunAt:      s.LastRunAt,
		NextRunAt:      s.NextRunAt,
		CreatedAt:      s.CreatedAt,
	}
}

func toSignalDTO(s db.Signal) SignalDTO {
	return SignalDTO{
		ID:             s.ID.String(),
		SubscriptionID: s.SubscriptionID.String(),
		Title:          s.Title,
		Body:           s.Body,
		URL:            s.URL,
		OccursAt:       s.OccursAt,
		SourceDomains:  s.SourceDomains,
		Confidence:     s.Confidence,
		FirstSeenAt:    s.FirstSeenAt,
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, ErrorResponse{Error: msg})
}
