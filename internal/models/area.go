package models

type Area struct {
	ID       uint   `gorm:"primaryKey;autoIncrement;column:area_id"`
	AreaName string `gorm:"type:varchar(255);not null"`

	ProvinceID uint     `gorm:"column:province_id"`
	Province   Province `gorm:"foreignKey:ProvinceID"`
}
