package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/paulsizon/notify/server/internal/agent"
)

type contextSuggestRequest struct {
	Context string `json:"context"`
}

func (h *Handler) suggestFromContext(w http.ResponseWriter, r *http.Request) {
	var req contextSuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctxText := strings.TrimSpace(req.Context)
	if l := len(ctxText); l < 10 || l > 2000 {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("context length must be 10-2000, got %d", l))
		return
	}

	if h.Suggester == nil {
		writeJSON(w, http.StatusOK, onboardingResponse{
			Suggestions: agent.FallbackSuggestions("", ""),
			Fallback:    true,
		})
		return
	}

	sugs, fallback, err := h.Suggester.SuggestFromContext(r.Context(), ctxText)
	if err != nil {
		writeJSON(w, http.StatusOK, onboardingResponse{
			Suggestions: agent.FallbackSuggestions("", ""),
			Fallback:    true,
		})
		return
	}
	writeJSON(w, http.StatusOK, onboardingResponse{Suggestions: sugs, Fallback: fallback})
}
