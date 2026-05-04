package info

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.Exec(`CREATE TABLE parties (
		party_id INTEGER PRIMARY KEY AUTOINCREMENT,
		party_no INTEGER, party_name TEXT, logo_url TEXT
	)`).Error; err != nil {
		t.Fatalf("create parties: %v", err)
	}
	if err := db.Exec(`CREATE TABLE candidates (
		candidate_no INTEGER, full_name TEXT,
		party_id INTEGER, area_id INTEGER, biography TEXT
	)`).Error; err != nil {
		t.Fatalf("create candidates: %v", err)
	}
	return db
}

func TestGetParties(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	db.Exec(`INSERT INTO parties (party_no, party_name, logo_url) VALUES (1,'Test Party','logo.png')`)

	result, err := repo.GetParties()
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 party, got %d", len(result))
	}
	if result[0].PartyName != "Test Party" {
		t.Errorf("expected 'Test Party', got '%s'", result[0].PartyName)
	}
}

func TestGetParties_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	result, err := repo.GetParties()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 parties, got %d", len(result))
	}
}

func TestGetCandidates(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	db.Exec(`INSERT INTO parties (party_no, party_name, logo_url) VALUES (1,'Test Party','logo.png')`)
	var partyID int
	db.Raw(`SELECT party_id FROM parties LIMIT 1`).Scan(&partyID)
	db.Exec(`INSERT INTO candidates (candidate_no, full_name, party_id, area_id, biography)
		VALUES (?, ?, ?, ?, ?)`, 1, "John Doe", partyID, 1, "test bio")

	result, err := repo.GetCandidates(1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(result))
	}
	if result[0].Name != "John Doe" {
		t.Errorf("expected 'John Doe', got '%s'", result[0].Name)
	}
}

func TestGetCandidates_WrongArea(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	db.Exec(`INSERT INTO parties (party_id, party_no, party_name, logo_url) VALUES (1,1,'P','')`)
	db.Exec(`INSERT INTO candidates (candidate_no, full_name, party_id, area_id, biography) VALUES (1,'Alice',1,1,'bio')`)

	result, err := repo.GetCandidates(99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 candidates for wrong area, got %d", len(result))
	}
}

func TestGetCandidates_MultipleAreas(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	db.Exec(`INSERT INTO parties (party_id, party_no, party_name, logo_url) VALUES (1,1,'P1',''),(2,2,'P2','')`)
	db.Exec(`INSERT INTO candidates (candidate_no, full_name, party_id, area_id, biography) VALUES
		(1,'Alice',1,1,'bio'),(2,'Bob',2,2,'bio'),(3,'Carol',1,1,'bio')`)

	r1, _ := repo.GetCandidates(1)
	r2, _ := repo.GetCandidates(2)
	if len(r1) != 2 {
		t.Errorf("area 1: expected 2, got %d", len(r1))
	}
	if len(r2) != 1 {
		t.Errorf("area 2: expected 1, got %d", len(r2))
	}
}
