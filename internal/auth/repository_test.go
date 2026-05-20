package auth

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"votespher/internal/models"
)

func setupRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.AutoMigrate(&models.Province{}, &models.Area{}, &models.Voter{}, &models.OTP{}, &models.Admin{})
	return db
}

func insertBaseData(db *gorm.DB) {
	db.Exec("INSERT INTO provinces (province_id, province_name) VALUES (1, 'Bangkok')")
	db.Exec("INSERT INTO areas (area_id, province_id, area_name) VALUES (1, 1, 'Area 1')")
	db.Exec("INSERT INTO voters (voter_id, citizen_id_hash, area_id, phone_number) VALUES (1, 'myhash', 1, '0812345678')")
}

func TestRepo_FindOTPByRefCode_Found(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	exp := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id,voter_id,otp_code,ref_code,expires_at,is_used,attempts) VALUES (1,1,'123456','ref001',?,0,0)", exp)

	repo := NewAuthRepository(db)
	otp, err := repo.FindOTPByRefCode("ref001")
	if err != nil || otp.OTPCode != "123456" {
		t.Fatalf("expected OTP found, got err=%v", err)
	}
}

func TestRepo_FindOTPByRefCode_NotFound(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewAuthRepository(db)
	if _, err := repo.FindOTPByRefCode("noref"); err == nil {
		t.Error("expected error")
	}
}

func TestRepo_FindOTPByRefCode_Expired(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	exp := time.Now().Add(-5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id,voter_id,otp_code,ref_code,expires_at,is_used,attempts) VALUES (1,1,'123456','ref002',?,0,0)", exp)

	repo := NewAuthRepository(db)
	if _, err := repo.FindOTPByRefCode("ref002"); err == nil {
		t.Error("expected error for expired OTP")
	}
}

func TestRepo_FindOTPByRefCode_AlreadyUsed(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	exp := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id,voter_id,otp_code,ref_code,expires_at,is_used,attempts) VALUES (1,1,'123456','ref003',?,1,0)", exp)

	repo := NewAuthRepository(db)
	if _, err := repo.FindOTPByRefCode("ref003"); err == nil {
		t.Error("expected error for used OTP")
	}
}

func TestRepo_MarkOTPAsUsed(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	exp := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id,voter_id,otp_code,ref_code,expires_at,is_used,attempts) VALUES (1,1,'123456','ref004',?,0,0)", exp)

	repo := NewAuthRepository(db)
	if err := repo.MarkOTPAsUsed(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var otp models.OTP
	db.First(&otp, 1)
	if !otp.IsUsed {
		t.Error("expected IsUsed=true")
	}
}

func TestRepo_UpdateOTPAttempts(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	exp := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id,voter_id,otp_code,ref_code,expires_at,is_used,attempts) VALUES (1,1,'123456','ref005',?,0,0)", exp)

	repo := NewAuthRepository(db)
	if err := repo.UpdateOTPAttempts(1, 3, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRepo_UpdateOTPAttempts_MarkUsed(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	exp := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO otps (otp_id,voter_id,otp_code,ref_code,expires_at,is_used,attempts) VALUES (1,1,'123456','ref006',?,0,0)", exp)

	repo := NewAuthRepository(db)
	if err := repo.UpdateOTPAttempts(1, 5, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRepo_FindVoterByID_Found(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	repo := NewAuthRepository(db)
	voter, err := repo.FindVoterByID(1)
	if err != nil || voter.ID != 1 {
		t.Fatalf("expected voter, got err=%v", err)
	}
}

func TestRepo_FindVoterByID_NotFound(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewAuthRepository(db)
	if _, err := repo.FindVoterByID(999); err == nil {
		t.Error("expected error")
	}
}

func TestRepo_FindVoterByCitizenIDHash_Found(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	repo := NewAuthRepository(db)
	voter, err := repo.FindVoterByCitizenIDHash("myhash")
	if err != nil || voter.ID != 1 {
		t.Fatalf("expected voter, got err=%v", err)
	}
}

func TestRepo_FindVoterByCitizenIDHash_NotFound(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewAuthRepository(db)
	if _, err := repo.FindVoterByCitizenIDHash("badhash"); err == nil {
		t.Error("expected error")
	}
}

func TestRepo_CreateOTP(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	repo := NewAuthRepository(db)
	otp := &models.OTP{
		VoterID:   1,
		OTPCode:   "654321",
		RefCode:   "ref007",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := repo.CreateOTP(otp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if otp.ID == 0 {
		t.Error("expected OTP ID to be set")
	}
}

func TestRepo_CheckIsAdmin_True(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	db.Exec("INSERT INTO admins (admin_id, voter_id) VALUES (1, 1)")
	repo := NewAuthRepository(db)
	if !repo.CheckIsAdmin(1) {
		t.Error("expected voter 1 to be admin")
	}
}

func TestRepo_CheckIsAdmin_False(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewAuthRepository(db)
	if repo.CheckIsAdmin(999) {
		t.Error("expected voter 999 NOT to be admin")
	}
}

func TestRepo_FindVoterWithArea_NotFound(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewAuthRepository(db)
	if _, err := repo.FindVoterWithArea(999); err == nil {
		t.Error("expected error for not found")
	}
}

func TestRepo_FindVoterWithArea(t *testing.T) {
	db := setupRepoTestDB(t)
	insertBaseData(db)
	repo := NewAuthRepository(db)
	voter, err := repo.FindVoterWithArea(1)
	if err != nil || voter.ID != 1 {
		t.Fatalf("unexpected error: %v", err)
	}
	if voter.Area.AreaName != "Area 1" {
		t.Error("expected area to be preloaded")
	}
}
