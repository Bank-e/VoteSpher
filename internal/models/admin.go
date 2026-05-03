package models

type Admin struct {
    ID      uint  `gorm:"primaryKey;autoIncrement;column:admin_id"` // ใช้ ID แต่ผูกกับคอลัมน์ admin_id
    VoterID uint  `gorm:"not null;uniqueIndex"`
    Voter   Voter // ไม่ต้องใส่ Tag เลย GORM ผูกให้อัตโนมัติ!
}