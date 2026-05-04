package models

type Province struct {
	ID           uint   `gorm:"primaryKey;column:province_id"`
	ProvinceName string `gorm:"column:province_name;unique"`
}
