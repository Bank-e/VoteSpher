package result

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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.Exec(`CREATE TABLE areas  (area_id INTEGER PRIMARY KEY, area_name TEXT, province_id INTEGER)`)
	db.Exec(`CREATE TABLE parties (party_id INTEGER PRIMARY KEY, party_no INTEGER, party_name TEXT, logo_url TEXT)`)
	db.Exec(`CREATE TABLE votes  (vote_id INTEGER PRIMARY KEY, area_id INTEGER, party_id INTEGER, candidate_id INTEGER, created_at DATETIME)`)
	return db
}

func seedData(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec(`INSERT INTO areas  VALUES (1,'กรุงเทพฯ เขต 1',1)`)
	db.Exec(`INSERT INTO parties VALUES (1,1,'พรรค A','')`)
	db.Exec(`INSERT INTO parties VALUES (2,2,'พรรค B','')`)
	db.Exec(`INSERT INTO votes  VALUES (1,1,1,NULL,datetime('now'))`)
	db.Exec(`INSERT INTO votes  VALUES (2,1,1,NULL,datetime('now'))`)
	db.Exec(`INSERT INTO votes  VALUES (3,1,2,NULL,datetime('now'))`)
}

func newRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/results/provinces/:provinces_name/areas/:area_id", GetProvinceAreaResultHandler(db))
	return r
}

func TestGetProvinceAreaResult_InvalidAreaID(t *testing.T) {
	db := setupTestDB(t)
	r := newRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/results/provinces/BKK/areas/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetProvinceAreaResult_AreaNotFound(t *testing.T) {
	db := setupTestDB(t)
	r := newRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/results/provinces/BKK/areas/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetProvinceAreaResult_Success(t *testing.T) {
	db := setupTestDB(t)
	seedData(t, db)
	r := newRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/results/provinces/BKK/areas/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected at least 1 party in result")
	}
}

func TestGetProvinceAreaResult_NoVotes(t *testing.T) {
	db := setupTestDB(t)
	db.Exec(`INSERT INTO areas VALUES (2,'เขต 2',1)`)
	r := newRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/results/provinces/BKK/areas/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (empty result), got %d", w.Code)
	}
	var result []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected empty result for area with no votes, got %d", len(result))
	}
}
