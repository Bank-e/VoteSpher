package election

<<<<<<< Updated upstream
import "gorm.io/gorm"

// UpdateElectionConfig อัปเดตการตั้งค่าการเลือกตั้ง
func UpdateElectionConfig(db *gorm.DB, req UpdateConfigRequest) (*ConfigResponse, error) {
	// ดึง config ที่ active อยู่มาก่อน
	cfg, err := GetActiveConfig(db)
	if err != nil {
		return nil, err
	}

	// อัปเดตค่าตาม request
	if err := UpdateConfig(db, cfg, req.Status, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// คืนค่า config ที่อัปเดตแล้ว
=======
import (
	"context"
	"strings"
	"time"
	"votespher/internal/models"
)

// Service คือ business logic layer ของ election
// แยก interface จาก implementation เพื่อให้ handler สามารถ mock ได้
type Service interface {
	UpdateElectionConfig(ctx context.Context, voterID uint, req UpdateConfigRequest) (*ConfigResponse, error)
}

// service เก็บ dependencies ที่ business logic ต้องใช้
// ตัวเล็กเพื่อบังคับให้คนใช้ผ่าน NewService
type service struct {
	repo Repository
}

// NewService สร้าง Service ใหม่จาก Repository
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// statusXxx เป็น constant แทนการเขียน string ดิบ — ลดโอกาสพิมพ์ผิด
const (
	statusPrepare  = "PREPARE"
	statusOpen     = "OPEN"
	statusPaused   = "PAUSED"
	statusClosed   = "CLOSED"
	statusCounting = "COUNTING"
)

// UpdateElectionConfig อัปเดตการตั้งค่าการเลือกตั้งแบบ Versioning
//
// ขั้นตอน:
//  1. ตรวจสอบสิทธิ์ admin จาก voterID
//  2. validate request (status, time range)
//  3. ดึง config เดิมที่ active อยู่
//  4. ตรวจสอบ state machine
//  5. สร้าง config ใหม่และยกเลิกของเก่าใน transaction เดียว
func (s *service) UpdateElectionConfig(ctx context.Context, voterID uint, req UpdateConfigRequest) (*ConfigResponse, error) {
	// 1. เช็คสิทธิ์แอดมิน
	admin, err := s.repo.GetAdminByVoterID(ctx, voterID)
	if err != nil {
		return nil, forbidden(ErrNotAdmin)
	}

	// 2. normalize + validate request
	newStatus := normalizeStatus(req.Status)
	if err := validateStatus(newStatus); err != nil {
		return nil, err
	}
	if err := validateTimeRange(req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// 3. ดึง config ที่ active อยู่ปัจจุบัน
	cfg, err := s.repo.GetActiveConfig(ctx)
	if err != nil {
		return nil, internal(ErrConfigNotFound, err)
	}

	// 4. State Machine — ป้องกัน transition ที่ไม่ถูกต้อง
	if err := validateStateTransition(cfg.Status, newStatus); err != nil {
		return nil, err
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

	// 6. บันทึกแบบ Transaction
	if err := s.repo.CreateConfigVersion(ctx, cfg, &newConfig); err != nil {
		// ถ้า repo คืน *AppError มาแล้ว ส่งต่อตรงๆ
		if _, ok := AsAppError(err); ok {
			return nil, err
		}
		return nil, internal(ErrConfigUpdateFailed, err)
	}

	// 7. คืนค่า config ที่อัปเดตแล้ว
>>>>>>> Stashed changes
	return &ConfigResponse{
		ConfigID:  cfg.ConfigID,
		Status:    cfg.Status,
		StartTime: cfg.StartTime,
		EndTime:   cfg.EndTime,
		UpdatedAt: cfg.UpdatedAt,
		IsActive:  cfg.IsActive,
	}, nil
}
<<<<<<< Updated upstream
=======

// ============================================================
// Helpers — แยกออกมาเพื่อเขียน test ทีละชิ้นได้สะดวก
// ============================================================

// normalizeStatus แปลง status ให้เป็นตัวพิมพ์ใหญ่ทั้งหมด
// ป้องกันคนพิมพ์สลับ (เช่น Open, open, OPEN ให้เท่ากันหมด)
func normalizeStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

// allowedStatuses คือเซตของ status ที่ระบบรองรับ
var allowedStatuses = map[string]struct{}{
	statusPrepare:  {},
	statusOpen:     {},
	statusPaused:   {},
	statusClosed:   {},
	statusCounting: {},
}

// validateStatus ตรวจว่าค่า status (ที่ normalize แล้ว) อยู่ในชุดที่รองรับ
func validateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return badRequest(ErrInvalidStatus)
	}
	return nil
}

// validateTimeRange ตรวจสอบว่าเวลาเริ่มต้น < เวลาสิ้นสุด
func validateTimeRange(start, end time.Time) error {
	if start.After(end) || start.Equal(end) {
		return badRequest(ErrInvalidTimeRange)
	}
	return nil
}

// validateStateTransition ตรวจสอบกฎ state machine
//
// PREPARE --> OPEN
// OPEN    --> PAUSED, CLOSED
// PAUSED  --> OPEN
// CLOSED  --> COUNTING (ห้ามกลับไป OPEN เด็ดขาด!)
func validateStateTransition(oldStatus, newStatus string) error {
	if oldStatus == statusClosed && newStatus == statusOpen {
		return badRequest(ErrInvalidStateTransition)
	}
	return nil
}
>>>>>>> Stashed changes
