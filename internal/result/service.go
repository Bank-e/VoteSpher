package result

import "gorm.io/gorm"

func GetProvinceAreaResultService(db *gorm.DB, provinceName string, areaID string) (interface{}, error) {
	return GetVoteResultByArea(db, areaID)
}
