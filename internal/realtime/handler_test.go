package realtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ===== Mock DB =====
func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// create tables
	db.Exec(`CREATE TABLE areas (area_id INTEGER, area_name TEXT);`)
	db.Exec(`CREATE TABLE votes (vote_id INTEGER, area_id INTEGER);`)

	// seed data
	db.Exec(`INSERT INTO areas (area_id, area_name) VALUES (1, 'A'), (2, 'B');`)
	db.Exec(`INSERT INTO votes (vote_id, area_id) VALUES 
	(1,1),(2,1),(3,2);`)

	return db
}

// ===== Test: BuildResponse =====
func TestBuildResponseV2(t *testing.T) {

	rows := []AreaVoteRow{
		{AreaID: 1, AreaName: "A", TotalVotes: 100},
		{AreaID: 2, AreaName: "B", TotalVotes: 200},
	}

	resp := BuildResponse(rows)

	if resp.TotalVotes != 300 {
		t.Errorf("expected total votes 300, got %d", resp.TotalVotes)
	}

	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}

	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}

// ===== Test: Handler basic =====
func TestGetAllAreasVotesHandler(t *testing.T) {

	db := setupTestDB()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/votes", GetAllAreasVotesHandler(db))

	req, _ := http.NewRequest(http.MethodGet, "/votes", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// ===== Test: Handler JSON format =====
func TestHandlerResponseFormat(t *testing.T) {
	db := setupTestDB()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/votes", GetAllAreasVotesHandler(db))

	req, _ := http.NewRequest(http.MethodGet, "/votes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if resp.TotalVotes == 0 {
		t.Error("expected total votes > 0")
	}

	if len(resp.Areas) == 0 {
		t.Error("expected areas but got empty")
	}

	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}
