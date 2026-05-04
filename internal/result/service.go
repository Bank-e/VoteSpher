package result

import "gorm.io/gorm"

func GetAreaResultService(db *gorm.DB, areaID uint) (interface{}, error) {
	// เรียก repository
	return GetVoteResultByArea(db, areaID)
}
