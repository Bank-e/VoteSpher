package election

import (
	"context"
	"errors"
	"testing"
	"time"

	"votespher/internal/models"
)

// ============================================================
// mockRepository — implement Repository สำหรับ unit test
// ทำเองแบบมือ ไม่พึ่งไลบรารี mock เพื่อไม่เพิ่ม dependency
// ============================================================

type mockRepository struct {
	// ค่าที่จะให้แต่ละเมธอดคืน
	adminToReturn  *models.Admin
	adminErr       error
	configToReturn *models.SystemConfig
	configErr      error
	createErr      error

	// สำหรับเช็คว่าถูกเรียกจริง + ด้วยอะไร
	createCalledWithOld *models.SystemConfig
	createCalledWithNew *models.SystemConfig
}

func (m *mockRepository) GetAdminByVoterID(_ context.Context, _ uint) (*models.Admin, error) {
	return m.adminToReturn, m.adminErr
}

func (m *mockRepository) GetActiveConfig(_ context.Context) (*models.SystemConfig, error) {
	return m.configToReturn, m.configErr
}

func (m *mockRepository) CreateConfigVersion(_ context.Context, oldCfg, newCfg *models.SystemConfig) error {
	m.createCalledWithOld = oldCfg
	m.createCalledWithNew = newCfg
	return m.createErr
}

// ============================================================
// helpers
// ============================================================

func validRequest() UpdateConfigRequest {
	return UpdateConfigRequest{
		Status:    "OPEN",
		StartTime: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 5, 1, 17, 0, 0, 0, time.UTC),
	}
}

func okMockRepo() *mockRepository {
	return &mockRepository{
		adminToReturn:  &models.Admin{ID: 99, VoterID: 1},
		configToReturn: &models.SystemConfig{ID: 1, Status: statusPrepare, IsActive: true},
	}
}

func assertAppError(t *testing.T, err error, wantStatus int, wantSentinel error) {
	t.Helper()
	appErr, ok := AsAppError(err)
	if !ok {
		t.Fatalf("expected *AppError, got %T: %v", err, err)
	}
	if appErr.HTTPStatus() != wantStatus {
		t.Fatalf("expected status %d, got %d (msg=%q)", wantStatus, appErr.HTTPStatus(), appErr.Message)
	}
	if wantSentinel != nil && !errors.Is(err, wantSentinel) {
		t.Fatalf("expected wraps sentinel %v, got %v", wantSentinel, err)
	}
}

// ============================================================
// pure-helper tests
// ============================================================

func TestNormalizeStatus(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"open", "OPEN"},
		{"Open", "OPEN"},
		{" closed ", "CLOSED"},
		{"PAUSED", "PAUSED"},
	}
	for _, tc := range cases {
		got := normalizeStatus(tc.in)
		if got != tc.want {
			t.Errorf("normalizeStatus(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestValidateTimeRange(t *testing.T) {
	now := time.Now()

	if err := validateTimeRange(now, now.Add(time.Hour)); err != nil {
		t.Errorf("expected nil error for valid range, got %v", err)
	}

	if err := validateTimeRange(now, now.Add(-time.Hour)); err == nil {
		t.Error("expected error for end-before-start, got nil")
	}

	if err := validateTimeRange(now, now); err == nil {
		t.Error("expected error for end-equals-start, got nil")
	}
}

func TestValidateStatus(t *testing.T) {
	for _, ok := range []string{"PREPARE", "OPEN", "PAUSED", "CLOSED", "COUNTING"} {
		if err := validateStatus(ok); err != nil {
			t.Errorf("expected %q to be valid, got %v", ok, err)
		}
	}

	for _, bad := range []string{"", "FOO", "open", "Open"} {
		if err := validateStatus(bad); err == nil {
			t.Errorf("expected %q to be invalid, got nil", bad)
		}
	}
}

func TestValidateStateTransition(t *testing.T) {
	if err := validateStateTransition(statusClosed, statusOpen); err == nil {
		t.Error("expected error for CLOSED -> OPEN, got nil")
	}

	allowed := [][2]string{
		{statusPrepare, statusOpen},
		{statusOpen, statusPaused},
		{statusOpen, statusClosed},
		{statusPaused, statusOpen},
		{statusClosed, statusCounting},
	}
	for _, p := range allowed {
		if err := validateStateTransition(p[0], p[1]); err != nil {
			t.Errorf("expected no error for %s -> %s, got %v", p[0], p[1], err)
		}
	}
}

// ============================================================
// service tests
// ============================================================

func TestUpdateElectionConfig_Success(t *testing.T) {
	repo := okMockRepo()
	svc := NewService(repo)

	resp, err := svc.UpdateElectionConfig(context.Background(), 1, validRequest())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Status != "OPEN" {
		t.Errorf("expected status OPEN, got %q", resp.Status)
	}
	if !resp.IsActive {
		t.Error("expected IsActive=true")
	}

	// ตรวจว่า repo ถูกเรียกด้วย admin id ที่ถูกต้อง
	if repo.createCalledWithNew == nil {
		t.Fatal("expected CreateConfigVersion to be called")
	}
	if repo.createCalledWithNew.AdminID != 99 {
		t.Errorf("expected AdminID=99, got %d", repo.createCalledWithNew.AdminID)
	}
}

func TestUpdateElectionConfig_NotAdmin(t *testing.T) {
	repo := okMockRepo()
	repo.adminToReturn = nil
	repo.adminErr = errors.New("record not found")
	svc := NewService(repo)

	_, err := svc.UpdateElectionConfig(context.Background(), 1, validRequest())
	assertAppError(t, err, 403, ErrNotAdmin)
}

func TestUpdateElectionConfig_InvalidStatus(t *testing.T) {
	repo := okMockRepo()
	svc := NewService(repo)

	req := validRequest()
	req.Status = "NOT_A_STATUS"

	_, err := svc.UpdateElectionConfig(context.Background(), 1, req)
	assertAppError(t, err, 400, ErrInvalidStatus)
}

func TestUpdateElectionConfig_InvalidTimeRange(t *testing.T) {
	repo := okMockRepo()
	svc := NewService(repo)

	req := validRequest()
	req.StartTime, req.EndTime = req.EndTime, req.StartTime // สลับให้ start > end

	_, err := svc.UpdateElectionConfig(context.Background(), 1, req)
	assertAppError(t, err, 400, ErrInvalidTimeRange)
}

func TestUpdateElectionConfig_NoActiveConfig(t *testing.T) {
	repo := okMockRepo()
	repo.configToReturn = nil
	repo.configErr = errors.New("record not found")
	svc := NewService(repo)

	_, err := svc.UpdateElectionConfig(context.Background(), 1, validRequest())
	assertAppError(t, err, 500, ErrConfigNotFound)
}

func TestUpdateElectionConfig_ClosedToOpenForbidden(t *testing.T) {
	repo := okMockRepo()
	repo.configToReturn = &models.SystemConfig{ID: 1, Status: statusClosed, IsActive: true}
	svc := NewService(repo)

	_, err := svc.UpdateElectionConfig(context.Background(), 1, validRequest())
	assertAppError(t, err, 400, ErrInvalidStateTransition)
}

func TestUpdateElectionConfig_RepositoryFailurePropagatesAppError(t *testing.T) {
	repo := okMockRepo()
	repo.createErr = internal(ErrConfigUpdateFailed, errors.New("db down"))
	svc := NewService(repo)

	_, err := svc.UpdateElectionConfig(context.Background(), 1, validRequest())
	assertAppError(t, err, 500, ErrConfigUpdateFailed)
}

func TestUpdateElectionConfig_RepositoryRawErrorWrapped(t *testing.T) {
	repo := okMockRepo()
	repo.createErr = errors.New("some raw error")
	svc := NewService(repo)

	_, err := svc.UpdateElectionConfig(context.Background(), 1, validRequest())
	assertAppError(t, err, 500, ErrConfigUpdateFailed)
}

func TestUpdateElectionConfig_NormalizesLowercaseStatus(t *testing.T) {
	repo := okMockRepo()
	svc := NewService(repo)

	req := validRequest()
	req.Status = "open"

	resp, err := svc.UpdateElectionConfig(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "OPEN" {
		t.Errorf("expected normalized status OPEN, got %q", resp.Status)
	}
}
