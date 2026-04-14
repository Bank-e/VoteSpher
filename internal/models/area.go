package models

type Area struct {
    ID       uint   `gorm:"primaryKey;autoIncrement;column:area_id"`
    AreaName string `gorm:"type:varchar(255);not null"`
}