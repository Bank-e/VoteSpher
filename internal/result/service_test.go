package result

import (
	"errors"
	"testing"
)

type mockResultRepository struct {
	result AreaResultResponse
	err    error
}

func (m *mockResultRepository) GetVoteResultByArea(areaID uint) (AreaResultResponse, error) {
	if m.err != nil {
		return AreaResultResponse{}, m.err
	}
	return m.result, nil
}

func TestResultService_GetAreaResult_Success(t *testing.T) {
	mockRepo := &mockResultRepository{
		result: AreaResultResponse{
			AreaID:   1,
			AreaName: "กรุงเทพมหานคร เขต 1",
			CandidateResults: []CandidateResult{
				{
					CandidateNo: 1,
					Name:        "นายสมชาย รักชาติ",
					Votes:       3,
				},
			},
			PartyListResults: []PartyResult{
				{
					PartyNo:   31,
					PartyName: "พรรคก้าวหน้า",
					Votes:     3,
				},
			},
		},
	}

	service := NewResultService(mockRepo)

	result, err := service.GetAreaResult(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.AreaID != 1 {
		t.Fatalf("expected area_id = 1, got %d", result.AreaID)
	}

	if result.AreaName != "กรุงเทพมหานคร เขต 1" {
		t.Fatalf("expected area name กรุงเทพมหานคร เขต 1, got %s", result.AreaName)
	}

	if len(result.CandidateResults) != 1 {
		t.Fatalf("expected 1 candidate result, got %d", len(result.CandidateResults))
	}

	if len(result.PartyListResults) != 1 {
		t.Fatalf("expected 1 party result, got %d", len(result.PartyListResults))
	}
}

func TestResultService_GetAreaResult_Error(t *testing.T) {
	mockRepo := &mockResultRepository{
		err: errors.New("database error"),
	}

	service := NewResultService(mockRepo)

	_, err := service.GetAreaResult(1)

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if err.Error() != "database error" {
		t.Fatalf("expected database error, got %v", err)
	}
}
