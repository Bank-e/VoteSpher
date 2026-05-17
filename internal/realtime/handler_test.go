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
	db.Exec(`CREATE TABLE areas (area_id INTEGER PRIMARY KEY, area_name TEXT);`)
	db.Exec(`CREATE TABLE parties (party_id INTEGER PRIMARY KEY, party_no INTEGER, party_name TEXT, logo_url TEXT);`)
	db.Exec(`CREATE TABLE candidates (candidate_id INTEGER PRIMARY KEY, area_id INTEGER, party_id INTEGER, candidate_no INTEGER, full_name TEXT, biography TEXT);`)
	db.Exec(`CREATE TABLE votes (vote_id INTEGER PRIMARY KEY, area_id INTEGER, candidate_id INTEGER, party_id INTEGER, created_at DATETIME);`)

	// seed data
	db.Exec(`INSERT INTO areas (area_id, area_name) VALUES (1, 'Area A'), (2, 'Area B');`)
	db.Exec(`INSERT INTO parties (party_id, party_no, party_name) VALUES (1, 1, 'Party X'), (2, 2, 'Party Y');`)
	db.Exec(`INSERT INTO candidates (candidate_id, area_id, party_id, candidate_no, full_name) VALUES 
		(1, 1, 1, 1, 'Alice'),
		(2, 1, 2, 2, 'Bob'),
		(3, 2, 1, 1, 'Charlie');`)
	db.Exec(`INSERT INTO votes (vote_id, area_id, candidate_id, party_id) VALUES 
		(1, 1, 1, 1),
		(2, 1, 1, 1),
		(3, 1, 2, 2),
		(4, 2, 3, 1),
		(5, 2, 3, 1);`)

	return db
}

// ===== Test: BuildResponse =====
func TestBuildResponseV2(t *testing.T) {

	areaRows := []AreaVoteRow{
		{AreaID: 1, AreaName: "A", TotalVotes: 100},
		{AreaID: 2, AreaName: "B", TotalVotes: 200},
	}

	candidateRows := []AreaCandidateRow{
		{AreaID: 1, CandidateNo: 1, CandidateName: "Alice", PartyName: "Party X", Votes: 60},
	}

	partyRows := []PartyVoteRow{
		{PartyNo: 1, PartyName: "Party X", Votes: 180},
	}

	resp := buildResponse(areaRows, candidateRows, partyRows)

	if resp.TotalVotes != 300 {
		t.Errorf("expected total votes 300, got %d", resp.TotalVotes)
	}

	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}

	if len(resp.Party) != 1 {
		t.Errorf("expected 1 party, got %d", len(resp.Party))
	}

	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}

// ===== Test: Handler basic =====
func TestGetAllAreasVotesHandler(t *testing.T) {

	db := setupTestDB()

	repo := NewRealtimeRepository(db)
	svc := NewRealtimeService(repo)
	handler := NewRealtimeHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/votes", handler.GetAllAreasVotes)

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

	repo := NewRealtimeRepository(db)
	svc := NewRealtimeService(repo)
	handler := NewRealtimeHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/votes", handler.GetAllAreasVotes)

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

	// ✅ check that candidates are populated
	hasCandidates := false
	for _, area := range resp.Areas {
		if len(area.Candidates) > 0 {
			hasCandidates = true
			break
		}
	}
	if !hasCandidates {
		t.Error("expected at least one area with candidates")
	}

	// ✅ check that party is populated
	if len(resp.Party) == 0 {
		t.Error("expected party results but got empty")
	}

	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}
