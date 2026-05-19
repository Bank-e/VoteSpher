package realtime

import "gorm.io/gorm"

type RealtimeRepository interface {
	GetAllAreasVotes() ([]AreaVoteRow, error)
	GetTopCandidatesByArea(limit int) ([]AreaCandidateRow, error)
	GetPartyVotes() ([]PartyVoteRow, error)
}

type realtimeRepository struct {
	db *gorm.DB
}

func NewRealtimeRepository(db *gorm.DB) RealtimeRepository {
	return &realtimeRepository{db: db}
}

func (r *realtimeRepository) GetAllAreasVotes() ([]AreaVoteRow, error) {
	var results []AreaVoteRow
	err := r.db.Table("votes v").
		Select("a.area_id, a.area_name, COUNT(v.vote_id) as total_votes").
		Joins("JOIN areas a ON v.area_id = a.area_id").
		Group("a.area_id, a.area_name").
		Scan(&results).Error
	return results, err
}

// GetTopCandidatesByArea returns top N candidates per area using ROW_NUMBER window function
func (r *realtimeRepository) GetTopCandidatesByArea(limit int) ([]AreaCandidateRow, error) {
	var results []AreaCandidateRow
	subquery := `
		SELECT
			v.area_id,
			c.candidate_no,
			c.full_name AS candidate_name,
			p.party_name,
			COUNT(v.vote_id) AS votes,
			ROW_NUMBER() OVER (PARTITION BY v.area_id ORDER BY COUNT(v.vote_id) DESC) AS rn
		FROM votes v
		JOIN candidates c ON v.candidate_id = c.candidate_id
		JOIN parties p ON c.party_id = p.party_id
		WHERE v.candidate_id IS NOT NULL
		GROUP BY v.area_id, c.candidate_no, c.full_name, p.party_name
	`
	err := r.db.Raw("SELECT area_id, candidate_no, candidate_name, party_name, votes FROM ("+subquery+") ranked WHERE rn <= ?", limit).
		Scan(&results).Error
	return results, err
}

func (r *realtimeRepository) GetPartyVotes() ([]PartyVoteRow, error) {
	var results []PartyVoteRow
	err := r.db.Table("votes v").
		Select("p.party_no, p.party_name, COUNT(v.vote_id) AS votes").
		Joins("JOIN parties p ON v.party_id = p.party_id").
		Where("v.party_id IS NOT NULL").
		Group("p.party_no, p.party_name").
		Order("votes DESC").
		Scan(&results).Error
	return results, err
}
