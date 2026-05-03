package result

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"votespher/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.AutoMigrate(&models.Area{}, &models.Party{}, &models.Vote{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	return db
}

func seedTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// area
	if err := db.Exec(
		`INSERT INTO areas (area_id, area_name) VALUES (?, ?)`,
		1, "กรุงเทพมหานคร เขต 1",
	).Error; err != nil {
		t.Fatalf("failed to insert area: %v", err)
	}

	// parties
	if err := db.Exec(
		`INSERT INTO parties (party_id, party_no, party_name) VALUES (?, ?, ?)`,
		1, 1, "พรรคประชาชน",
	).Error; err != nil {
		t.Fatalf("failed to insert party 1: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO parties (party_id, party_no, party_name) VALUES (?, ?, ?)`,
		3, 3, "พรรคภูมิใจไทย",
	).Error; err != nil {
		t.Fatalf("failed to insert party 3: %v", err)
	}

	if err := db.Exec(
		`INSERT INTO parties (party_id, party_no, party_name) VALUES (?, ?, ?)`,
		4, 4, "พรรคกล้าธรรม",
	).Error; err != nil {
		t.Fatalf("failed to insert party 4: %v", err)
	}

	// votes in area 1
	votes := []struct {
		voteID  int
		areaID  int
		partyID int
	}{
		{1, 1, 1},
		{2, 1, 1},
		{3, 1, 3},
		{4, 1, 3},
		{5, 1, 4},
	}

	for _, v := range votes {
		if err := db.Exec(
			`INSERT INTO votes (vote_id, area_id, party_id) VALUES (?, ?, ?)`,
			v.voteID, v.areaID, v.partyID,
		).Error; err != nil {
			t.Fatalf("failed to insert vote %d: %v", v.voteID, err)
		}
	}
}

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.GET("/results/provinces/:provinces_name/areas/:area_id", GetProvinceAreaResultHandler(db))
	return r
}

func TestGetProvinceAreaResultHandler_InvalidAreaID(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/results/provinces/bangkok/areas/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "invalid area_id: must be a number"
	if resp["error"] != expected {
		t.Fatalf("expected error %q, got %q", expected, resp["error"])
	}
}

func TestGetProvinceAreaResultHandler_AreaNotFound(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/results/provinces/bangkok/areas/99999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["error"] != "area not found" {
		t.Fatalf("expected error %q, got %q", "area not found", resp["error"])
	}
}

func TestGetProvinceAreaResultHandler_Success(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/results/provinces/bangkok/areas/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 3 {
		t.Fatalf("expected 3 result rows, got %d", len(resp))
	}

	results := make(map[string]float64)
	for _, row := range resp {
		partyName, ok := row["party_name"].(string)
		if !ok {
			t.Fatalf("party_name missing or invalid: %#v", row)
		}

		total, ok := row["total"].(float64)
		if !ok {
			t.Fatalf("total missing or invalid: %#v", row)
		}

		results[partyName] = total
	}

	if results["พรรคประชาชน"] != 2 {
		t.Fatalf("expected พรรคประชาชน total = 2, got %v", results["พรรคประชาชน"])
	}
	if results["พรรคภูมิใจไทย"] != 2 {
		t.Fatalf("expected พรรคภูมิใจไทย total = 2, got %v", results["พรรคภูมิใจไทย"])
	}
	if results["พรรคกล้าธรรม"] != 1 {
		t.Fatalf("expected พรรคกล้าธรรม total = 1, got %v", results["พรรคกล้าธรรม"])
	}
}
func TestGetProvinceAreaResultHandler_ProvinceMismatchButAreaExists(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/results/provinces/unknown/areas/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 3 {
		t.Fatalf("expected 3 result rows, got %d", len(resp))
	}
}

func TestGetProvinceAreaResultHandler_ProvinceMismatchAndAreaNotFound(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/results/provinces/xxxx/areas/99999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["error"] != "area not found" {
		t.Fatalf("expected error %q, got %q", "area not found", resp["error"])
	}
}

func TestGetProvinceAreaResultHandler_AreaExistsButNoVotes(t *testing.T) {
	db := setupTestDB(t)

	if err := db.Exec(
		`INSERT INTO areas (area_id, area_name) VALUES (?, ?)`,
		2, "กรุงเทพมหานคร เขต 2",
	).Error; err != nil {
		t.Fatalf("failed to insert area without votes: %v", err)
	}

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/results/provinces/bangkok/areas/2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 0 {
		t.Fatalf("expected empty result, got %d rows", len(resp))
	}
}
