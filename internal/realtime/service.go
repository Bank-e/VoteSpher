package realtime

func BuildResponse(rows []AreaVoteRow) Response {

	areas := []AreaResponse{}

	for _, r := range rows {
		areas = append(areas, AreaResponse{
			AreaID:     r.AreaID,
			AreaName:   r.AreaName,
			TotalVotes: r.TotalVotes,
		})
	}

	return Response{
		Areas: areas,
	}
}
