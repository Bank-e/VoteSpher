package info

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	// สร้าง table ให้ตรงกับ query จริง
	db.Exec(`
	CREATE TABLE parties (
		party_id INTEGER PRIMARY KEY AUTOINCREMENT,
		party_no INTEGER,
		party_name TEXT,
		logo_url TEXT
	);
	`)

	db.Exec(`
	CREATE TABLE candidates (
		candidate_no INTEGER,
		full_name TEXT,
		party_id INTEGER,
		area_id INTEGER,
		biography TEXT
	);
	`)

	return db
}

func TestGetParties(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`
	INSERT INTO parties (party_no, party_name, logo_url)
	VALUES (1, 'Test Party', 'logo.png');
	`)

	result, err := GetParties(db)

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestGetCandidates(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`
	INSERT INTO parties (party_id, party_no, party_name, logo_url)
	VALUES (1, 1, 'Test Party', 'logo.png');
	`)

	db.Exec(`
	INSERT INTO candidates (candidate_no, full_name, party_id, area_id, biography)
	VALUES (1, 'John Doe', 1, 1, 'test bio');
	`)

	result, err := GetCandidates(db, 1) // ✅ แก้ตรงนี้

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}

	if result[0].Name != "John Doe" {
		t.Fatalf("expected name John Doe, got %s", result[0].Name)
	}
}