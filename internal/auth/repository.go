package auth

import (
	"time"
	"votespher/internal/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	FindOTPByRefCode(refCode string) (*models.OTP, error)
	MarkOTPAsUsed(otpID uint) error
	UpdateOTPAttempts(otpID uint, attempts int, markUsed bool) error
	FindVoterByID(voterID uint) (*models.Voter, error)
	FindVoterByCitizenIDHash(citizenIDHash string) (*models.Voter, error)
	CreateOTP(otp *models.OTP) error
	CheckIsAdmin(voterID uint) bool
	FindVoterWithArea(voterID uint) (*models.Voter, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) FindOTPByRefCode(refCode string) (*models.OTP, error) {
	var otp models.OTP
	err := r.db.Where("ref_code = ? AND is_used = false AND expires_at > ?", refCode, time.Now()).
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *authRepository) MarkOTPAsUsed(otpID uint) error {
	return r.db.Model(&models.OTP{}).
		Where("otp_id = ?", otpID).
		Update("is_used", true).Error
}

func (r *authRepository) UpdateOTPAttempts(otpID uint, attempts int, markUsed bool) error {
	updates := map[string]interface{}{"attempts": attempts}
	if markUsed {
		updates["is_used"] = true
	}
	return r.db.Model(&models.OTP{}).Where("otp_id = ?", otpID).Updates(updates).Error
}

func (r *authRepository) FindVoterByID(voterID uint) (*models.Voter, error) {
	var voter models.Voter
	err := r.db.First(&voter, voterID).Error
	if err != nil {
		return nil, err
	}
	return &voter, nil
}

func (r *authRepository) FindVoterByCitizenIDHash(citizenIDHash string) (*models.Voter, error) {
	var voter models.Voter
	err := r.db.Preload("Area.Province").Where("citizen_id_hash = ?", citizenIDHash).First(&voter).Error
	if err != nil {
		return nil, err
	}
	return &voter, nil
}

func (r *authRepository) CreateOTP(otp *models.OTP) error {
	return r.db.Create(otp).Error
}

func (r *authRepository) CheckIsAdmin(voterID uint) bool {
	var count int64
	r.db.Model(&models.Admin{}).Where("voter_id = ?", voterID).Count(&count)
	return count > 0
}

func (r *authRepository) FindVoterWithArea(voterID uint) (*models.Voter, error) {
	var voter models.Voter
	err := r.db.Preload("Area.Province").First(&voter, voterID).Error
	if err != nil {
		return nil, err
	}
	return &voter, nil
}
