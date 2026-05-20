package realtime

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupRealtimeRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec(`CREATE TABLE areas (area_id INTEGER PRIMARY KEY, area_name TEXT)`)
	db.Exec(`CREATE TABLE parties (party_id INTEGER PRIMARY KEY, party_no INTEGER, party_name TEXT)`)
	db.Exec(`CREATE TABLE candidates (candidate_id INTEGER PRIMARY KEY, area_id INTEGER, party_id INTEGER, candidate_no INTEGER, full_name TEXT)`)
	db.Exec(`CREATE TABLE votes (vote_id INTEGER PRIMARY KEY, area_id INTEGER, candidate_id INTEGER, party_id INTEGER, created_at DATETIME)`)
	return db
}

func TestRepoGetAllAreasVotes_Success(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	db.Exec(`INSERT INTO areas VALUES (1, 'Area A'), (2, 'Area B')`)
	db.Exec(`INSERT INTO votes (vote_id, area_id, candidate_id, party_id) VALUES (1, 1, 1, 1), (2, 1, 1, 1), (3, 2, 2, 2)`)

	repo := NewRealtimeRepository(db)
	rows, err := repo.GetAllAreasVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 area rows, got %d", len(rows))
	}
}

func TestRepoGetAllAreasVotes_Empty(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	repo := NewRealtimeRepository(db)
	rows, err := repo.GetAllAreasVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected empty, got %d", len(rows))
	}
}

func TestRepoGetAllAreasVotes_DBError(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	repo := NewRealtimeRepository(db)
	if _, err := repo.GetAllAreasVotes(); err == nil {
		t.Error("expected error for closed DB")
	}
}

func TestRepoGetTopCandidatesByArea_Success(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	db.Exec(`INSERT INTO areas VALUES (1, 'Area A')`)
	db.Exec(`INSERT INTO parties VALUES (1, 1, 'Party X')`)
	db.Exec(`INSERT INTO candidates VALUES (1, 1, 1, 1, 'Alice')`)
	db.Exec(`INSERT INTO votes (vote_id, area_id, candidate_id, party_id) VALUES (1, 1, 1, 1), (2, 1, 1, 1)`)

	repo := NewRealtimeRepository(db)
	rows, err := repo.GetTopCandidatesByArea(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) == 0 {
		t.Error("expected at least 1 row")
	}
}

func TestRepoGetTopCandidatesByArea_Empty(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	repo := NewRealtimeRepository(db)
	rows, err := repo.GetTopCandidatesByArea(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected empty, got %d", len(rows))
	}
}

func TestRepoGetTopCandidatesByArea_DBError(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	repo := NewRealtimeRepository(db)
	if _, err := repo.GetTopCandidatesByArea(3); err == nil {
		t.Error("expected error for closed DB")
	}
}

func TestRepoGetPartyVotes_Success(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	db.Exec(`INSERT INTO parties VALUES (1, 1, 'Party X'), (2, 2, 'Party Y')`)
	db.Exec(`INSERT INTO votes (vote_id, area_id, candidate_id, party_id) VALUES (1, 1, NULL, 1), (2, 1, NULL, 1), (3, 1, NULL, 2)`)

	repo := NewRealtimeRepository(db)
	rows, err := repo.GetPartyVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 party rows, got %d", len(rows))
	}
}

func TestRepoGetPartyVotes_Empty(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	repo := NewRealtimeRepository(db)
	rows, err := repo.GetPartyVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected empty, got %d", len(rows))
	}
}

func TestRepoGetPartyVotes_DBError(t *testing.T) {
	db := setupRealtimeRepoDB(t)
	sqlDB, _ := db.DB()
	sqlDB.Close()

	repo := NewRealtimeRepository(db)
	if _, err := repo.GetPartyVotes(); err == nil {
		t.Error("expected error for closed DB")
	}
}
