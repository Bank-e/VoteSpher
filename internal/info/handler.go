package info

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// 🔹 Struct
type InfoHandler struct {
	service InfoService
}

// 🔹 Constructor
func NewInfoHandler(service InfoService) *InfoHandler {
	return &InfoHandler{service: service}
}

// 🔹 GET /candidates
func (h *InfoHandler) GetCandidatesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		areaIDStr := r.URL.Query().Get("area_id")
		if areaIDStr == "" {
			http.Error(w, "area_id is required", http.StatusBadRequest)
			return
		}

		areaID, err := strconv.Atoi(areaIDStr)
		if err != nil {
			http.Error(w, "invalid area_id", http.StatusBadRequest)
			return
		}

		result, err := h.service.GetCandidates(areaID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

// 🔹 GET /parties
func (h *InfoHandler) GetPartiesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		result, err := h.service.GetParties()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}
