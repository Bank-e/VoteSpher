package result

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupRepoResultDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.Exec("CREATE TABLE areas (area_id INTEGER PRIMARY KEY, area_name TEXT, province_id INTEGER)")
	db.Exec("CREATE TABLE parties (party_id INTEGER PRIMARY KEY, party_no INTEGER, party_name TEXT)")
	db.Exec("CREATE TABLE candidates (candidate_id INTEGER PRIMARY KEY, area_id INTEGER, party_id INTEGER, candidate_no INTEGER, full_name TEXT)")
	db.Exec("CREATE TABLE votes (vote_id INTEGER PRIMARY KEY, area_id INTEGER, candidate_id INTEGER, party_id INTEGER, created_at DATETIME)")
	return db
}

func TestRepo_GetVoteResultByArea_AreaNotFound(t *testing.T) {
	db := setupRepoResultDB(t)
	repo := NewResultRepository(db)
	_, err := repo.GetVoteResultByArea(999)
	if err == nil {
		t.Error("expected error for non-existent area")
	}
}

func TestRepo_GetVoteResultByArea_CandidateQueryError(t *testing.T) {
	db := setupRepoResultDB(t)
	db.Exec("INSERT INTO areas (area_id, area_name, province_id) VALUES (1, 'Test', 1)")
	db.Exec("DROP TABLE IF EXISTS votes")

	repo := NewResultRepository(db)
	_, err := repo.GetVoteResultByArea(1)
	if err == nil {
		t.Error("expected error when votes table is missing")
	}
}

func TestRepo_GetVoteResultByArea_PartyQueryError(t *testing.T) {
	// votes WITHOUT party_id column → candidate query succeeds, party query fails
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.Exec("CREATE TABLE areas (area_id INTEGER PRIMARY KEY, area_name TEXT, province_id INTEGER)")
	db.Exec("CREATE TABLE parties (party_id INTEGER PRIMARY KEY, party_no INTEGER, party_name TEXT)")
	db.Exec("CREATE TABLE candidates (candidate_id INTEGER PRIMARY KEY, area_id INTEGER, party_id INTEGER, candidate_no INTEGER, full_name TEXT)")
	db.Exec("CREATE TABLE votes (vote_id INTEGER PRIMARY KEY, area_id INTEGER, candidate_id INTEGER, created_at DATETIME)") // NO party_id
	db.Exec("INSERT INTO areas VALUES (1, 'Test Area', 1)")

	repo := NewResultRepository(db)
	_, err := repo.GetVoteResultByArea(1)
	if err == nil {
		t.Error("expected error from party query (votes.party_id missing)")
	}
}
