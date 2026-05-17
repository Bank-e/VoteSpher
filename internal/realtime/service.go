package realtime

import "time"

type RealtimeService interface {
	GetAllAreasResult() (Response, error)
}

type realtimeService struct {
	repo RealtimeRepository
}

func NewRealtimeService(repo RealtimeRepository) RealtimeService {
	return &realtimeService{repo: repo}
}

func (s *realtimeService) GetAllAreasResult() (Response, error) {
	rows, err := s.repo.GetAllAreasVotes()
	if err != nil {
		return Response{}, err
	}
	return buildResponse(rows), nil
}

func buildResponse(rows []AreaVoteRow) Response {

	areas := []AreaResponse{}
	totalVotes := 0

	for _, r := range rows {
		areas = append(areas, AreaResponse{
			AreaID:     r.AreaID,
			AreaName:   r.AreaName,
			TotalVotes: r.TotalVotes,
		})

		totalVotes += r.TotalVotes
	}

	return Response{
		TotalVotes:  totalVotes,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Areas:       areas,
	}
}
