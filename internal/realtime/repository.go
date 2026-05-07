package realtime

import "gorm.io/gorm"

func GetAreaVotes(db *gorm.DB, areaID string) ([]AreaVoteRow, error) {

	var results []AreaVoteRow

	err := db.Table("votes v").
		Select("a.area_id, a.area_name, COUNT(v.vote_id) as total_votes").
		Joins("JOIN areas a ON v.area_id = a.area_id").
		Where("v.area_id = ?", areaID).
		Group("a.area_id, a.area_name").
		Scan(&results).Error

	return results, err
}
func GetAllAreasVotes(db *gorm.DB) ([]AreaVoteRow, error) {

	var results []AreaVoteRow

	err := db.Table("votes v").
		Select("a.area_id, a.area_name, COUNT(v.vote_id) as total_votes").
		Joins("JOIN areas a ON v.area_id = a.area_id").
		Group("a.area_id, a.area_name").
		Scan(&results).Error

	return results, err
}
