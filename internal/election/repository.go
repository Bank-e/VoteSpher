package election

import (
	"context"
	"votespher/internal/models"

	"gorm.io/gorm"
)

// Repository คือสัญญา (contract) ของ Data Access Layer ของ election
type Repository interface {
	GetAdminByVoterID(ctx context.Context, voterID uint) (*models.Admin, error)
	GetActiveConfig(ctx context.Context) (*models.SystemConfig, error)
	CreateConfigVersion(ctx context.Context, oldConfig, newConfig *models.SystemConfig) error
}

// repository คือ implementation จริงของ Repository (พึ่งพา *gorm.DB)
type repository struct {
	db *gorm.DB
}

// NewRepository สร้าง Repository ใหม่จาก *gorm.DB
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// GetAdminByVoterID ค้นหาข้อมูลแอดมินจากรหัสผู้โหวต
func (r *repository) GetAdminByVoterID(ctx context.Context, voterID uint) (*models.Admin, error) {
	var admin models.Admin
	if err := r.db.WithContext(ctx).Where("voter_id = ?", voterID).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetActiveConfig หา config ที่ active อยู่ตัวล่าสุด
func (r *repository) GetActiveConfig(ctx context.Context) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// CreateConfigVersion ยกเลิกของเก่าและสร้างของใหม่แบบ Transaction
//
// ขั้นตอน:
//  1. set is_active = false ของ config เก่า
//  2. insert config ใหม่
//
// ถ้าขั้นใดขั้นหนึ่งล้มเหลว จะ rollback ทั้งหมด
func (r *repository) CreateConfigVersion(ctx context.Context, oldConfig, newConfig *models.SystemConfig) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. ยกเลิกของเก่า (set is_active = false)
		if err := tx.Model(oldConfig).Update("is_active", false).Error; err != nil {
			return internal(ErrConfigDeactivate, err)
		}

		// 2. บันทึกของใหม่ (Insert row ใหม่)
		if err := tx.Create(newConfig).Error; err != nil {
			return internal(ErrConfigUpdateFailed, err)
		}

		return nil
	})
}
