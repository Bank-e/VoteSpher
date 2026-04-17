package election

import (
	"errors"
	"strings"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// UpdateElectionConfig อัปเดตการตั้งค่าการเลือกตั้งแบบ Versioning
func UpdateElectionConfig(db *gorm.DB, voterID uint, req UpdateConfigRequest) (*ConfigResponse, error) {
	// เช็คสิทธิ์แอดมิน (Business Logic)
	admin, err := GetAdminByVoterID(db, voterID)
	if err != nil {
		return nil, errors.New("403: คุณไม่มีสิทธิ์ผู้ดูแลระบบ")
	}

	// 1. แปลง Status เป็นตัวพิมพ์ใหญ่ทั้งหมด ป้องกันคนพิมพ์สลับ (เช่น Open, open, OPEN)
	newStatus := strings.ToUpper(req.Status)

	// 2. ตรวจสอบความสมเหตุสมผลของเวลา (กันแอดมินตั้งเวลาผิด)
	if req.StartTime.After(req.EndTime) {
		return nil, errors.New("400: เวลาเปิดหีบต้องมาก่อนเวลาปิดหีบเสมอ")
	}

	// 3. ดึง config ที่ active อยู่ปัจจุบัน
	cfg, err := GetActiveConfig(db)
	if err != nil {
		return nil, errors.New("500: ไม่พบการตั้งค่าระบบที่กำลังใช้งานอยู่")
	}

	// 4. State Machine (ป้องกันเปิดหีบใหม่ถ้าระบบปิดไปแล้ว)
	
	// PREPARE (เตรียมการ) --> เปลี่ยนเป็น OPEN (เปิดหีบ) ได้
    // OPEN (เปิดหีบ) --> เปลี่ยนเป็น PAUSED (พักเบรก/ไฟดับ) หรือ CLOSED (ปิดหีบ) ได้
	// PAUSED (พักเบรก) --> เปลี่ยนกลับเป็น OPEN (เปิดหีบต่อ) ได้
	// CLOSED (ปิดหีบ) --> ต้องเปลี่ยนเป็น COUNTING (นับคะแนน) เท่านั้น! ห้ามกลับไป OPEN เด็ดขาด
	if cfg.Status == "CLOSED" && newStatus == "OPEN" {
		return nil, errors.New("400: ไม่อนุญาตให้เปิดหีบใหม่ เนื่องจากระบบถูกปิด (CLOSED) ไปแล้ว")
	}

	// 5. เตรียมข้อมูล Config ชุดใหม่ (ประทับตรา adminID ลงไป)
	newConfig := models.SystemConfig{
		AdminID:   admin.ID,
		Status:    newStatus,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		IsActive:  true,
		// ฟิลด์ ID และ UpdatedAt จะถูก GORM สร้างให้อัตโนมัติ
	}

	// 6. เรียกใช้ Repository เพื่อบันทึกแบบ Transaction
	if err := CreateConfigVersion(db, cfg, &newConfig); err != nil {
		return nil, errors.New("500: " + err.Error())
	}

	// 7. คืนค่า config ที่อัปเดตแล้ว
	return &ConfigResponse{
		ConfigID:  newConfig.ID,
		Status:    newConfig.Status,
		StartTime: newConfig.StartTime,
		EndTime:   newConfig.EndTime,
		UpdatedAt: newConfig.UpdatedAt,
		IsActive:  newConfig.IsActive,
	}, nil
}