package realtime

import "time"

func BuildResponse(rows []AreaVoteRow) Response {

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
