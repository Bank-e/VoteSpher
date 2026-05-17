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

	// 1. ดึงจำนวนโหวตรวมของแต่ละเขต
	areaRows, err := s.repo.GetAllAreasVotes()
	if err != nil {
		return Response{}, err
	}

	// 2. ดึง Top 3 ผู้สมัครของแต่ละเขต
	candidateRows, err := s.repo.GetTopCandidatesByArea(3)
	if err != nil {
		return Response{}, err
	}

	// 3. ดึงจำนวนโหวตรวมของแต่ละพรรค
	partyRows, err := s.repo.GetPartyVotes()
	if err != nil {
		return Response{}, err
	}

	return buildResponse(areaRows, candidateRows, partyRows), nil
}

func buildResponse(areaRows []AreaVoteRow, candidateRows []AreaCandidateRow, partyRows []PartyVoteRow) Response {

	// จัดกลุ่ม candidates ตาม area_id
	candidateMap := make(map[int][]CandidateResponse)
	for _, c := range candidateRows {
		candidateMap[c.AreaID] = append(candidateMap[c.AreaID], CandidateResponse{
			CandidateNo:   c.CandidateNo,
			CandidateName: c.CandidateName,
			PartyName:     c.PartyName,
			Votes:         c.Votes,
		})
	}

	// สร้าง areas response
	areas := []AreaResponse{}
	totalVotes := 0

	for _, r := range areaRows {
		candidates := candidateMap[r.AreaID]
		if candidates == nil {
			candidates = []CandidateResponse{}
		}

		areas = append(areas, AreaResponse{
			AreaID:     r.AreaID,
			AreaName:   r.AreaName,
			TotalVotes: r.TotalVotes,
			Candidates: candidates,
		})

		totalVotes += r.TotalVotes
	}

	// สร้าง party response
	party := []PartyResponse{}
	for _, p := range partyRows {
		party = append(party, PartyResponse{
			PartyNo:   p.PartyNo,
			PartyName: p.PartyName,
			Votes:     p.Votes,
		})
	}

	return Response{
		TotalVotes:  totalVotes,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Areas:       areas,
		Party:       party,
	}
}
