package election

import (
	"context"
	"errors"
	"testing"
	"time"

	"votespher/internal/models"

	"github.com/glebarez/sqlite" // pure-Go sqlite driver — ไม่ต้องการ CGO/gcc
	"gorm.io/gorm"
)

// setupRepoDB สร้าง sqlite in-memory + migrate schema
//
// ใช้ ":memory:" (ไม่ใช่ shared cache) เพื่อให้แต่ละ test มี DB แยกกัน
// ป้องกัน test pollution ระหว่างกัน
//
// driver: github.com/glebarez/sqlite (pure-Go, no CGO required)
func setupRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.Voter{}, &models.Admin{}, &models.SystemConfig{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	return db
}

func TestRepository_GetAdminByVoterID_Found(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	admin := models.Admin{ID: 1, VoterID: 100}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	got, err := repo.GetAdminByVoterID(context.Background(), 100)
	if err != nil {
		t.Fatalf("expected found, got error: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("expected admin id=1, got %+v", got)
	}
}

func TestRepository_GetAdminByVoterID_NotFound(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	_, err := repo.GetAdminByVoterID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for unknown voterID, got nil")
	}
}

func TestRepository_GetActiveConfig_NotFound(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	if _, err := repo.GetActiveConfig(context.Background()); err == nil {
		t.Fatal("expected error when no active config exists")
	}
}

func TestRepository_GetActiveConfig_Found(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	cfg := models.SystemConfig{
		AdminID:   1,
		Status:    statusPrepare,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  true,
	}
	if err := db.Create(&cfg).Error; err != nil {
		t.Fatalf("seed config: %v", err)
	}

	got, err := repo.GetActiveConfig(context.Background())
	if err != nil {
		t.Fatalf("expected found, got %v", err)
	}
	if !got.IsActive {
		t.Errorf("expected IsActive=true, got %v", got.IsActive)
	}
	if got.Status != statusPrepare {
		t.Errorf("expected status=%s, got %s", statusPrepare, got.Status)
	}
}

func TestRepository_CreateConfigVersion_Success(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	// seed old active config
	oldCfg := models.SystemConfig{
		AdminID:   1,
		Status:    statusPrepare,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  true,
	}
	if err := db.Create(&oldCfg).Error; err != nil {
		t.Fatalf("seed old config: %v", err)
	}

	newCfg := &models.SystemConfig{
		AdminID:   1,
		Status:    statusOpen,
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now().Add(2 * time.Hour),
		IsActive:  true,
	}

	if err := repo.CreateConfigVersion(context.Background(), &oldCfg, newCfg); err != nil {
		t.Fatalf("CreateConfigVersion: %v", err)
	}

	// 1. old config ต้องถูก deactivate
	var refreshedOld models.SystemConfig
	if err := db.First(&refreshedOld, oldCfg.ID).Error; err != nil {
		t.Fatalf("reload old: %v", err)
	}
	if refreshedOld.IsActive {
		t.Error("expected old config to be deactivated")
	}

	// 2. new config ต้องถูกสร้างและ active
	var newRefreshed models.SystemConfig
	if err := db.First(&newRefreshed, newCfg.ID).Error; err != nil {
		t.Fatalf("reload new: %v", err)
	}
	if !newRefreshed.IsActive {
		t.Error("expected new config to be active")
	}
	if newRefreshed.Status != statusOpen {
		t.Errorf("expected new status=OPEN, got %s", newRefreshed.Status)
	}

	// 3. ต้องมี config 2 แถว (เก็บ version ครบ)
	var count int64
	if err := db.Model(&models.SystemConfig{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 config rows, got %d", count)
	}
}

// TestRepository_CreateConfigVersion_DeactivateFails ทดสอบ branch ที่ Update step
// (ขั้นยกเลิก config เก่า) ล้มเหลว
//
// trigger ด้วยการ DROP table หลังจาก seed — ทำให้ Update SQL ล้มเหลวเพราะ
// table หาย ส่งผลให้ repo คืน error ที่ wrap ErrConfigDeactivate
func TestRepository_CreateConfigVersion_DeactivateFails(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	oldCfg := models.SystemConfig{
		ID:        1,
		AdminID:   1,
		Status:    statusPrepare,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  true,
	}
	if err := db.Create(&oldCfg).Error; err != nil {
		t.Fatalf("seed old: %v", err)
	}

	// ทำลาย schema เพื่อบังคับให้ Update ล้มเหลว
	if err := db.Exec("DROP TABLE system_configs").Error; err != nil {
		t.Fatalf("drop table: %v", err)
	}

	newCfg := &models.SystemConfig{
		AdminID:   1,
		Status:    statusOpen,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  true,
	}

	err := repo.CreateConfigVersion(context.Background(), &oldCfg, newCfg)
	if err == nil {
		t.Fatal("expected error after dropping table, got nil")
	}
	if !errors.Is(err, ErrConfigDeactivate) {
		t.Errorf("expected wraps ErrConfigDeactivate, got %v", err)
	}
}

// TestRepository_CreateConfigVersion_RollbackOnCreateFail ทดสอบว่า transaction
// rollback ถูกต้องเมื่อขั้นตอนที่ 2 (Create newConfig) ล้มเหลว
//
// trigger ความล้มเหลวด้วยการใส่ AdminID เป็นค่าที่ไม่มี (ไม่กระทบใน sqlite ที่ FK disabled)
// — ในที่นี้เราใส่ค่า ID ซ้ำกันเพื่อบังคับให้ INSERT ล้มเหลว
func TestRepository_CreateConfigVersion_RollbackOnCreateFail(t *testing.T) {
	db := setupRepoDB(t)
	repo := NewRepository(db)

	// seed config สอง rows: id=1 active, id=2 ที่จะถูก Create ซ้ำให้ INSERT ล้มเหลว
	oldCfg := models.SystemConfig{
		ID:        1,
		AdminID:   1,
		Status:    statusPrepare,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  true,
	}
	if err := db.Create(&oldCfg).Error; err != nil {
		t.Fatalf("seed old: %v", err)
	}

	occupied := models.SystemConfig{
		ID:        2,
		AdminID:   1,
		Status:    statusOpen,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  false,
	}
	if err := db.Create(&occupied).Error; err != nil {
		t.Fatalf("seed occupied: %v", err)
	}

	// new config ใช้ ID=2 ซ้ำ — INSERT จะล้มเหลวด้วย UNIQUE constraint
	newCfg := &models.SystemConfig{
		ID:        2,
		AdminID:   1,
		Status:    statusClosed,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		IsActive:  true,
	}

	err := repo.CreateConfigVersion(context.Background(), &oldCfg, newCfg)
	if err == nil {
		t.Fatal("expected error from duplicate primary key, got nil")
	}

	// ต้อง rollback: oldCfg ต้องยังคง IsActive=true
	var reloaded models.SystemConfig
	if err := db.First(&reloaded, oldCfg.ID).Error; err != nil {
		t.Fatalf("reload old: %v", err)
	}
	if !reloaded.IsActive {
		t.Error("expected rollback: oldCfg should still be active")
	}
}
