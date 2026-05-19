package result

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"votespher/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.Area{}, &models.Party{}, &models.Candidate{}, &models.Vote{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func seedTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`INSERT INTO areas (area_id, area_name, province_id) VALUES (?, ?, ?)`, 1, "กรุงเทพมหานคร เขต 1", 1).Error; err != nil {
		t.Fatalf("failed to insert area: %v", err)
	}
	if err := db.Exec(`INSERT INTO parties (party_id, party_no, party_name) VALUES (?, ?, ?)`, 1, 31, "พรรคก้าวหน้า").Error; err != nil {
		t.Fatalf("failed to insert party 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO parties (party_id, party_no, party_name) VALUES (?, ?, ?)`, 2, 29, "พรรคเพื่อธรรม").Error; err != nil {
		t.Fatalf("failed to insert party 2: %v", err)
	}
	if err := db.Exec(`INSERT INTO candidates (candidate_id, area_id, party_id, candidate_no, full_name) VALUES (?, ?, ?, ?, ?)`, 1, 1, 1, 1, "นายสมชาย รักชาติ").Error; err != nil {
		t.Fatalf("failed to insert candidate 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO candidates (candidate_id, area_id, party_id, candidate_no, full_name) VALUES (?, ?, ?, ?, ?)`, 2, 1, 2, 2, "นางสาวสมหญิง มุ่งมั่น").Error; err != nil {
		t.Fatalf("failed to insert candidate 2: %v", err)
	}
	votes := []struct{ voteID, areaID, candidateID, partyID int }{
		{1, 1, 1, 1}, {2, 1, 1, 1}, {3, 1, 1, 1}, {4, 1, 2, 2}, {5, 1, 2, 2},
	}
	for _, v := range votes {
		if err := db.Exec(`INSERT INTO votes (vote_id, area_id, candidate_id, party_id, created_at) VALUES (?, ?, ?, ?, datetime('now'))`, v.voteID, v.areaID, v.candidateID, v.partyID).Error; err != nil {
			t.Fatalf("failed to insert vote %d: %v", v.voteID, err)
		}
	}
}

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := NewResultRepository(db)
	svc := NewResultService(repo)
	h := NewResultHandler(svc)
	r := gin.Default()
	r.GET("/results/areas/:area_id", h.GetAreaResult)
	return r
}

func setupProvinceRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := NewResultRepository(db)
	svc := NewResultService(repo)
	h := NewResultHandler(svc)
	r := gin.Default()
	r.GET("/results/provinces/:provinces_name/areas/:area_id", h.GetProvinceAreaResult)
	return r
}

func TestGetAreaResultHandler_InvalidAreaID(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/areas/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetAreaResultHandler_AreaNotFound(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/areas/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetAreaResultHandler_Success(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	r := setupRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/areas/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp AreaResultResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AreaID != 1 {
		t.Fatalf("expected area_id=1, got %d", resp.AreaID)
	}
	if len(resp.CandidateResults) != 2 {
		t.Fatalf("expected 2 candidate results, got %d", len(resp.CandidateResults))
	}
	if len(resp.PartyListResults) != 2 {
		t.Fatalf("expected 2 party results, got %d", len(resp.PartyListResults))
	}
	if resp.CandidateResults[0].Votes != 3 {
		t.Fatalf("expected first candidate votes=3, got %d", resp.CandidateResults[0].Votes)
	}
}

func TestGetAreaResultHandler_NoVotes(t *testing.T) {
	db := setupTestDB(t)
	if err := db.Exec(`INSERT INTO areas (area_id, area_name, province_id) VALUES (?, ?, ?)`, 2, "เชียงใหม่ เขต 1", 1).Error; err != nil {
		t.Fatalf("failed to insert empty area: %v", err)
	}
	r := setupRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/areas/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp AreaResultResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.CandidateResults) != 0 {
		t.Fatalf("expected empty candidate results")
	}
}

// setupIsolatedDB creates a private in-memory SQLite DB (not cache=shared) so province
// tests don't collide with data seeded by the GetAreaResult tests.
func setupIsolatedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open isolated db: %v", err)
	}
	if err := db.AutoMigrate(&models.Area{}, &models.Party{}, &models.Candidate{}, &models.Vote{}); err != nil {
		t.Fatalf("failed to migrate isolated db: %v", err)
	}
	return db
}

// --- GetProvinceAreaResult ---

func TestGetProvinceAreaResult_InvalidAreaID(t *testing.T) {
	db := setupIsolatedDB(t)
	r := setupProvinceRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/provinces/Bangkok/areas/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetProvinceAreaResult_AreaNotFound(t *testing.T) {
	db := setupIsolatedDB(t)
	r := setupProvinceRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/provinces/Bangkok/areas/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetProvinceAreaResult_Success(t *testing.T) {
	db := setupIsolatedDB(t)
	seedTestData(t, db)
	r := setupProvinceRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/provinces/Bangkok/areas/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestGetProvinceAreaResult_DatabaseError(t *testing.T) {
	db := setupIsolatedDB(t)
	sqlDB, _ := db.DB()
	sqlDB.Close()
	r := setupProvinceRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/provinces/Bangkok/areas/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetAreaResultHandler_DatabaseError(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	sqlDB.Close()
	r := setupRouter(db)
	req := httptest.NewRequest(http.MethodGet, "/results/areas/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
