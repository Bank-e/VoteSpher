package migration

import (
	"log"
	"votespher/internal/models"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) {
	// ปิด FK check ชั่วคราวระหว่าง migrate
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")

	err := db.AutoMigrate(
		// ลำดับที่ 1 — ไม่มี FK เลย
		&models.Area{},
		&models.Party{},

		// ลำดับที่ 2 — FK ไปหา Area
		&models.Voter{},
		&models.Candidate{},

		// ลำดับที่ 3 — FK ไปหา Voter
		&models.OTP{},
		&models.Admin{},

		// ลำดับที่ 4 — FK ไปหาหลายตาราง
		&models.Vote{},

		// ลำดับที่ 5 — FK ไปหา Admin
		&models.SystemConfig{},
	)

	// เปิด FK check กลับมาหลัง migrate เสร็จ
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration completed successfully")
}