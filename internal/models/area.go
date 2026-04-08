package models

type Area struct {
	AreaID   uint   `gorm:"primaryKey;autoIncrement"`
	AreaName string `gorm:"type:varchar(255);not null"`
}