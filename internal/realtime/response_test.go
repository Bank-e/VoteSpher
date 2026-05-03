package realtime

import (
	"testing"
)

func TestBuildResponse(t *testing.T) {

	rows := []AreaVoteRow{
		{AreaID: 1, AreaName: "A", TotalVotes: 100},
		{AreaID: 2, AreaName: "B", TotalVotes: 200},
	}

	resp := BuildResponse(rows)

	// ✅ check total votes
	if resp.TotalVotes != 300 {
		t.Errorf("expected total votes 300, got %d", resp.TotalVotes)
	}

	// ✅ check areas length
	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}

	// ✅ check last_updated not empty
	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}
