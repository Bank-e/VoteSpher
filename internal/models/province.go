package models

type Province struct {
	ID           uint   `gorm:"primaryKey;autoIncrement;column:province_id"`
	ProvinceName string `gorm:"type:varchar(255);not null;uniqueIndex"`
}
