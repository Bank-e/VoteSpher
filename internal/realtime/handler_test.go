package realtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.Exec(`CREATE TABLE areas (area_id INTEGER PRIMARY KEY, area_name TEXT)`)
	db.Exec(`CREATE TABLE votes  (vote_id INTEGER PRIMARY KEY, area_id INTEGER)`)
	db.Exec(`INSERT INTO areas VALUES (1,'เขต A'),(2,'เขต B')`)
	db.Exec(`INSERT INTO votes VALUES (1,1),(2,1),(3,2)`)
	return db
}

func TestBuildResponse(t *testing.T) {
	rows := []AreaVoteRow{
		{AreaID: 1, AreaName: "เขต A", TotalVotes: 100},
		{AreaID: 2, AreaName: "เขต B", TotalVotes: 200},
	}
	resp := BuildResponse(rows)
	if resp.TotalVotes != 300 {
		t.Errorf("expected total=300, got %d", resp.TotalVotes)
	}
	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}
	if resp.LastUpdated == "" {
		t.Error("last_updated must not be empty")
	}
}

func TestBuildResponse_Empty(t *testing.T) {
	resp := BuildResponse([]AreaVoteRow{})
	if resp.TotalVotes != 0 {
		t.Errorf("expected 0, got %d", resp.TotalVotes)
	}
	if len(resp.Areas) != 0 {
		t.Errorf("expected empty areas, got %d", len(resp.Areas))
	}
}

func TestGetAllAreasVotesHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	r := gin.New()
	r.GET("/results/areas", GetAllAreasVotesHandler(db))

	req, _ := http.NewRequest(http.MethodGet, "/results/areas", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.TotalVotes != 3 {
		t.Errorf("expected 3 total votes, got %d", resp.TotalVotes)
	}
	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}
}

func TestGetVoteResultByAreaHandler_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	r := gin.New()
	r.GET("/results/areas/:area_id", GetVoteResultByAreaHandler(db))

	req, _ := http.NewRequest(http.MethodGet, "/results/areas/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetVoteResultByAreaHandler_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	r := gin.New()
	r.GET("/results/areas/:area_id", GetVoteResultByAreaHandler(db))

	req, _ := http.NewRequest(http.MethodGet, "/results/areas/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-numeric area_id, got %d", w.Code)
	}
}
