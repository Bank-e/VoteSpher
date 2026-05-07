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

	t.Cleanup(func() {
		sqlDB.Close()
	})

	// create tables
	if err := db.Exec(`
	CREATE TABLE parties (
		party_id INTEGER PRIMARY KEY AUTOINCREMENT,
		party_no INTEGER,
		party_name TEXT,
		logo_url TEXT
	);
	`).Error; err != nil {
		t.Fatalf("create parties table error: %v", err)
	}

	if err := db.Exec(`
	CREATE TABLE candidates (
		candidate_no INTEGER,
		full_name TEXT,
		party_id INTEGER,
		area_id INTEGER,
		biography TEXT
	);
	`).Error; err != nil {
		t.Fatalf("create candidates table error: %v", err)
	}

	return db
}

func TestGetParties(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	if err := db.Exec(`
	INSERT INTO parties (party_no, party_name, logo_url)
	VALUES (1, 'Test Party', 'logo.png');
	`).Error; err != nil {
		t.Fatalf("insert error: %v", err)
	}

	result, err := repo.GetParties()

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestGetCandidates(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInfoRepository(db)

	// insert party
	if err := db.Exec(`
	INSERT INTO parties (party_no, party_name, logo_url)
	VALUES (1, 'Test Party', 'logo.png');
	`).Error; err != nil {
		t.Fatalf("insert party error: %v", err)
	}

	// 🔥 ดึง party_id จริงจาก DB
	var partyID int
	if err := db.Raw(`SELECT party_id FROM parties LIMIT 1`).Scan(&partyID).Error; err != nil {
		t.Fatalf("get party_id error: %v", err)
	}

	// insert candidate using real party_id
	if err := db.Exec(`
	INSERT INTO candidates (candidate_no, full_name, party_id, area_id, biography)
	VALUES (?, ?, ?, ?, ?);
	`, 1, "John Doe", partyID, 1, "test bio").Error; err != nil {
		t.Fatalf("insert candidate error: %v", err)
	}

	result, err := repo.GetCandidates(1)

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
