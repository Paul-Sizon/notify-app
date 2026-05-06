package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paulsizon/notify/server/internal/agent"
)

var validRoles = map[string]bool{
	"developer": true, "founder": true, "designer": true,
	"investor": true, "student": true, "other": true,
}

var validInterests = map[string]bool{
	"concerts": true, "tech_meetups": true, "crypto_web3": true,
	"fintech": true, "startups_vc": true, "ai_ml": true,
	"sports": true, "art_design": true, "food_restaurants": true,
	"politics_policy": true, "gaming": true, "film_tv": true,
}

type onboardingRequest struct {
	City      string   `json:"city"`
	Country   string   `json:"country"`
	Role      string   `json:"role"`
	RoleOther string   `json:"role_other,omitempty"`
	Interests []string `json:"interests"`
}

type onboardingResponse struct {
	Suggestions []agent.Suggestion `json:"suggestions"`
	Fallback    bool               `json:"fallback,omitempty"`
}

func (h *Handler) suggestOnboarding(w http.ResponseWriter, r *http.Request) {
	var req onboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateOnboarding(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Suggester == nil {
		// No LLM configured — serve the local fallback so the client still
		// gets something usable (dev/CI without an OpenAI key).
		writeJSON(w, http.StatusOK, onboardingResponse{
			Suggestions: agent.FallbackSuggestions(req.City, req.Country),
			Fallback:    true,
		})
		return
	}

	sugs, fallback, err := h.Suggester.Suggest(r.Context(), agent.SuggestInput{
		City:      req.City,
		Country:   req.Country,
		Role:      req.Role,
		RoleOther: req.RoleOther,
		Interests: req.Interests,
	})
	if err != nil {
		writeJSON(w, http.StatusOK, onboardingResponse{
			Suggestions: agent.FallbackSuggestions(req.City, req.Country),
			Fallback:    true,
		})
		return
	}
	writeJSON(w, http.StatusOK, onboardingResponse{Suggestions: sugs, Fallback: fallback})
}

func validateOnboarding(req *onboardingRequest) error {
	if l := len(req.City); l < 2 || l > 80 {
		return fmt.Errorf("city length must be 2-80, got %d", l)
	}
	if l := len(req.Country); l > 80 {
		return fmt.Errorf("country length must be 0-80, got %d", l)
	}
	if !validRoles[req.Role] {
		return fmt.Errorf("invalid role: %s", req.Role)
	}
	if req.Role == "other" {
		s := req.RoleOther
		if len(s) < 1 || len(s) > 60 {
			return fmt.Errorf("role_other length must be 1-60 when role=other")
		}
	}
	if len(req.Interests) < 1 {
		return fmt.Errorf("at least 1 interest required")
	}
	if len(req.Interests) > 8 {
		req.Interests = req.Interests[:8] // silent cap per spec
	}
	for _, in := range req.Interests {
		if !validInterests[in] {
			return fmt.Errorf("invalid interest: %s", in)
		}
	}
	return nil
}
