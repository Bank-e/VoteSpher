package realtime

import "gorm.io/gorm"

type RealtimeRepository interface {
	GetAllAreasVotes() ([]AreaVoteRow, error)
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
