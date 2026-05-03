package election

import (
	"context"
	"strings"
	"time"
	"votespher/internal/models"
)

// Service คือ business logic layer ของ election
type Service interface {
	UpdateElectionConfig(ctx context.Context, voterID uint, req UpdateConfigRequest) (*ConfigResponse, error)
}

// service เก็บ dependencies ที่ business logic ต้องใช้
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
	}

	// 6. บันทึกแบบ Transaction
	if err := s.repo.CreateConfigVersion(ctx, cfg, &newConfig); err != nil {
		if _, ok := AsAppError(err); ok {
			return nil, err
		}
		return nil, internal(ErrConfigUpdateFailed, err)
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

// ============================================================
// Helpers
// ============================================================

func normalizeStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

var allowedStatuses = map[string]struct{}{
	statusPrepare:  {},
	statusOpen:     {},
	statusPaused:   {},
	statusClosed:   {},
	statusCounting: {},
}

func validateStatus(status string) error {
	if _, ok := allowedStatuses[status]; !ok {
		return badRequest(ErrInvalidStatus)
	}
	return nil
}

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
