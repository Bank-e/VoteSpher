package info

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

// GET /candidates?area_id=1
func GetCandidatesHandler(db *gorm.DB) http.HandlerFunc {
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

		result, err := GetCandidatesService(db, areaID)
		if err != nil {
			log.Printf("GetCandidates error: %v", err)
			http.Error(w, "เกิดข้อผิดพลาดภายในระบบ", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

// GET /parties
func GetPartiesHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		result, err := GetPartiesService(db)
		if err != nil {
			log.Printf("GetParties error: %v", err)
			http.Error(w, "เกิดข้อผิดพลาดภายในระบบ", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}