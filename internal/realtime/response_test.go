package realtime

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ใช้ driver เดียวกับโปรเจค (glebarez/sqlite ไม่ใช่ gorm.io/driver/sqlite)
func setupRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	db.Exec(`CREATE TABLE areas (area_id INTEGER PRIMARY KEY, area_name TEXT)`)
	db.Exec(`CREATE TABLE parties (party_id INTEGER PRIMARY KEY, party_no INTEGER, party_name TEXT, logo_url TEXT)`)
	db.Exec(`CREATE TABLE candidates (candidate_id INTEGER PRIMARY KEY, area_id INTEGER, party_id INTEGER, candidate_no INTEGER, full_name TEXT, biography TEXT)`)
	db.Exec(`CREATE TABLE votes (vote_id INTEGER PRIMARY KEY, area_id INTEGER, candidate_id INTEGER, party_id INTEGER, created_at DATETIME)`)

	return db
}

func seedBasicData(db *gorm.DB) {
	db.Exec(`INSERT INTO areas VALUES (1, 'Area A'), (2, 'Area B')`)
	db.Exec(`INSERT INTO parties VALUES (1, 1, 'Party X', ''), (2, 2, 'Party Y', '')`)
	db.Exec(`INSERT INTO candidates VALUES
		(1, 1, 1, 1, 'Alice',   ''),
		(2, 1, 2, 2, 'Bob',     ''),
		(3, 2, 1, 1, 'Charlie', '')`)
	// Area A → Alice×3, Bob×2 | Area B → Charlie×5
	db.Exec(`INSERT INTO votes VALUES
		(1,  1, 1, 1, NULL),(2,  1, 1, 1, NULL),(3,  1, 1, 1, NULL),
		(4,  1, 2, 2, NULL),(5,  1, 2, 2, NULL),
		(6,  2, 3, 1, NULL),(7,  2, 3, 1, NULL),(8,  2, 3, 1, NULL),
		(9,  2, 3, 1, NULL),(10, 2, 3, 1, NULL)`)
}

// ----- buildResponse (pure function) -----

func TestBuildResponse(t *testing.T) {
	areaRows := []AreaVoteRow{
		{AreaID: 1, AreaName: "A", TotalVotes: 100},
		{AreaID: 2, AreaName: "B", TotalVotes: 200},
	}
	candidateRows := []AreaCandidateRow{
		{AreaID: 1, CandidateNo: 1, CandidateName: "Alice", PartyName: "Party X", Votes: 60},
		{AreaID: 1, CandidateNo: 2, CandidateName: "Bob", PartyName: "Party Y", Votes: 40},
		{AreaID: 2, CandidateNo: 3, CandidateName: "Charlie", PartyName: "Party X", Votes: 120},
	}
	partyRows := []PartyVoteRow{
		{PartyNo: 1, PartyName: "Party X", Votes: 180},
		{PartyNo: 2, PartyName: "Party Y", Votes: 120},
	}

	resp := buildResponse(areaRows, candidateRows, partyRows)

	if resp.TotalVotes != 300 {
		t.Errorf("expected total votes 300, got %d", resp.TotalVotes)
	}
	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}
	if len(resp.Areas[0].Candidates) != 2 {
		t.Errorf("expected 2 candidates in area 1, got %d", len(resp.Areas[0].Candidates))
	}
	if len(resp.Areas[1].Candidates) != 1 {
		t.Errorf("expected 1 candidate in area 2, got %d", len(resp.Areas[1].Candidates))
	}
	if len(resp.Party) != 2 {
		t.Errorf("expected 2 parties, got %d", len(resp.Party))
	}
	if resp.Party[0].PartyName != "Party X" {
		t.Errorf("expected first party 'Party X', got '%s'", resp.Party[0].PartyName)
	}
	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}

// ----- GetAllAreasVotes -----

func TestGetAllAreasVotes_ReturnsCorrectTotals(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetAllAreasVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 areas, got %d", len(rows))
	}

	totals := map[int]int{}
	for _, r := range rows {
		totals[r.AreaID] = r.TotalVotes
	}
	if totals[1] != 5 {
		t.Errorf("Area A: expected 5, got %d", totals[1])
	}
	if totals[2] != 5 {
		t.Errorf("Area B: expected 5, got %d", totals[2])
	}
}

func TestGetAllAreasVotes_EmptyVotes(t *testing.T) {
	db := setupRepoTestDB(t)
	db.Exec(`INSERT INTO areas VALUES (1, 'Area A')`)

	rows, err := NewRealtimeRepository(db).GetAllAreasVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

func TestGetAllAreasVotes_AreaNameIsPopulated(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetAllAreasVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := map[int]string{}
	for _, r := range rows {
		names[r.AreaID] = r.AreaName
	}
	if names[1] != "Area A" {
		t.Errorf("expected 'Area A', got '%s'", names[1])
	}
	if names[2] != "Area B" {
		t.Errorf("expected 'Area B', got '%s'", names[2])
	}
}

// ----- GetTopCandidatesByArea -----

func TestGetTopCandidatesByArea_LimitOne(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetTopCandidatesByArea(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (1 per area), got %d", len(rows))
	}

	winners := map[int]string{}
	for _, r := range rows {
		winners[r.AreaID] = r.CandidateName
	}
	if winners[1] != "Alice" {
		t.Errorf("Area A: expected Alice, got %s", winners[1])
	}
	if winners[2] != "Charlie" {
		t.Errorf("Area B: expected Charlie, got %s", winners[2])
	}
}

func TestGetTopCandidatesByArea_LimitTwo(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetTopCandidatesByArea(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Area A: 2 candidates, Area B: 1 → total 3
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}

func TestGetTopCandidatesByArea_VoteCountIsCorrect(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetTopCandidatesByArea(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	votesMap := map[string]int{}
	for _, r := range rows {
		votesMap[r.CandidateName] = r.Votes
	}
	if votesMap["Alice"] != 3 {
		t.Errorf("Alice: expected 3, got %d", votesMap["Alice"])
	}
	if votesMap["Bob"] != 2 {
		t.Errorf("Bob: expected 2, got %d", votesMap["Bob"])
	}
	if votesMap["Charlie"] != 5 {
		t.Errorf("Charlie: expected 5, got %d", votesMap["Charlie"])
	}
}

func TestGetTopCandidatesByArea_NoVotes(t *testing.T) {
	db := setupRepoTestDB(t)

	rows, err := NewRealtimeRepository(db).GetTopCandidatesByArea(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// ----- GetPartyVotes -----

func TestGetPartyVotes_ReturnsCorrectTotals(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetPartyVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 parties, got %d", len(rows))
	}

	totals := map[string]int{}
	for _, r := range rows {
		totals[r.PartyName] = r.Votes
	}
	if totals["Party X"] != 8 { // Alice(3)+Charlie(5)
		t.Errorf("Party X: expected 8, got %d", totals["Party X"])
	}
	if totals["Party Y"] != 2 { // Bob(2)
		t.Errorf("Party Y: expected 2, got %d", totals["Party Y"])
	}
}

func TestGetPartyVotes_OrderedByVotesDesc(t *testing.T) {
	db := setupRepoTestDB(t)
	seedBasicData(db)

	rows, err := NewRealtimeRepository(db).GetPartyVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(rows); i++ {
		if rows[i-1].Votes < rows[i].Votes {
			t.Errorf("not sorted DESC at index %d", i)
		}
	}
}

func TestGetPartyVotes_NoVotes(t *testing.T) {
	db := setupRepoTestDB(t)

	rows, err := NewRealtimeRepository(db).GetPartyVotes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}
