package models

type Province struct {
	ID           uint   `gorm:"primaryKey;autoIncrement;column:province_id"`
	ProvinceName string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Areas        []Area // GORM จะรู้ว่านี่คือ 1-to-Many
}