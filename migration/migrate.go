package migration

import (
	"log"
	"votespher/internal/models"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
    // ปล่อยให้ AutoMigrate จัดการทุกอย่าง (มันจะสร้างตารางก่อน แล้วค่อยทำ ALTER TABLE เพื่อใส่ FK ให้อัตโนมัติ)
    err := db.AutoMigrate(
        &models.Province{},
        &models.Area{},
        &models.Party{},
        &models.Voter{},
        &models.Candidate{},
        &models.OTP{},
        &models.Admin{},
        &models.SystemConfig{},
        &models.Vote{},
    )

    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    log.Println("Migration completed successfully")
}