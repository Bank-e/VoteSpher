package voting

import (
	"errors"
	"regexp"
	"testing"
	"time"
	"votespher/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ==========================================
// 🟢 1. Helper Function สำหรับสร้าง Mock DB
// ==========================================

// setupMockDB เป็นฟังก์ชันช่วยสร้าง Database จำลองที่ทำงานร่วมกับ GORM ได้
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	// 1. สร้าง mock database connection
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// 2. ตั้งค่า GORM ให้ใช้ connection จำลองนี้ (สมมติว่าใช้ MySQL)
	dialector := mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true, // ปิดการเช็คเวอร์ชัน DB ตอนเริ่มต้น
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

// ==========================================
// 🟢 2. Test Cases สำหรับ Repository
// ==========================================

func TestGetActiveConfig_Success(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"config_id", "status", "start_time", "end_time", "is_active"}).
		AddRow(1, "OPEN", now.Add(-1*time.Hour), now.Add(1*time.Hour), true)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `system_configs` WHERE is_active = ? ORDER BY `system_configs`.`config_id` LIMIT ?")).
		WithArgs(true, 1).
		WillReturnRows(rows)

	// 3. รันเทส
	config, err := repo.GetActiveConfig()

	// 4. ตรวจสอบผล
	assert.NoError(t, err)
	if config != nil { // ป้องกัน Panic ถ้ารอบหน้าหาไม่เจออีก
		assert.Equal(t, "OPEN", config.Status)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckUserHasVoted_Voted(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	// จำลองข้อมูลว่าโหวตแล้ว (true)
	rows := sqlmock.NewRows([]string{"is_voted"}).AddRow(true)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `is_voted` FROM `voters` WHERE `voters`.`voter_id` = ? ORDER BY `voters`.`voter_id` LIMIT ?")).
		WithArgs(123, 1).
		WillReturnRows(rows)

	isVoted, err := repo.CheckUserHasVoted(123)

	assert.NoError(t, err)
	assert.True(t, isVoted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteVoteTransaction_Success(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	voterID := uint(123)
	cID := uint(1)
	pID := uint(2)
	voteRecord := models.Vote{
		AreaID:      10,
		CandidateID: &cID,
		PartyID:     &pID,
		CreatedAt:   time.Now(),
	}

	// ==========================================
	// 🔴 จำลองขั้นตอนของ Transaction ทั้งหมด
	// ==========================================

	// 1. เริ่ม Transaction (BEGIN)
	mock.ExpectBegin()

	// 2. Select Voter พร้อมล็อค Row (FOR UPDATE)
	// แก้ไขตรงนี้: เปลี่ยน "id" เป็น "voter_id" ทั้งใน NewRows และใน Query
	voterRows := sqlmock.NewRows([]string{"voter_id", "is_voted"}).AddRow(voterID, false)
	
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `voters` WHERE `voters`.`voter_id` = ? ORDER BY `voters`.`voter_id` LIMIT ? FOR UPDATE")).
		WithArgs(voterID, 1).
		WillReturnRows(voterRows)

	// 3. บันทึกผลโหวต (INSERT)
	mock.ExpectExec("INSERT INTO `votes`").
		WillReturnResult(sqlmock.NewResult(1, 1)) // สมมติว่า Insert สำเร็จ ได้ ID = 1

	// 4. อัปเดตสถานะผู้ใช้ (UPDATE)
	mock.ExpectExec("UPDATE `voters` SET").
		WillReturnResult(sqlmock.NewResult(1, 1)) // อัปเดตสำเร็จ 1 แถว

	// 5. ยืนยัน Transaction (COMMIT)
	mock.ExpectCommit()

	// รันเทส
	err := repo.ExecuteVoteTransaction(voterID, voteRecord)

	// ตรวจสอบ
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExecuteVoteTransaction_Fail_AlreadyVoted(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	voterID := uint(123)

	// เริ่ม Transaction
	mock.ExpectBegin()

	// รอบนี้จำลองว่า Select ขึ้นมาแล้วพบว่า is_voted = true
	voterRows := sqlmock.NewRows([]string{"id", "is_voted"}).AddRow(voterID, true)
	mock.ExpectQuery("SELECT \\* FROM `voters`.*FOR UPDATE").
		WillReturnRows(voterRows)

	// ถ้าโหวตไปแล้ว ต้องโดนเตะออก และมีการ Rollback
	mock.ExpectRollback()

	err := repo.ExecuteVoteTransaction(voterID, models.Vote{})

	// ตรวจสอบว่าเป็น AppError 403 แบบที่เราตั้งไว้ไหม
	assert.Error(t, err)
	appErr, ok := err.(*AppError)
	assert.True(t, ok)
	assert.Equal(t, 403, appErr.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ************************************************************
func TestGetActiveConfig_Fail(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	// สั่งให้ Query นี้พ่น Error ออกมา
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `system_configs`")).
		WillReturnError(errors.New("database is down"))

	config, err := repo.GetActiveConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestExecuteVoteTransaction_Fail_VoterNotFound(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	mock.ExpectBegin()
	// สั่งให้ตอนล็อก Row หาคนไม่เจอ (ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `voters`")).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := repo.ExecuteVoteTransaction(123, models.Vote{})
	
	assert.Error(t, err)
	appErr, _ := err.(*AppError)
	assert.Equal(t, 404, appErr.Code)
}

func TestExecuteVoteTransaction_Fail_InsertVote(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	mock.ExpectBegin()
	voterRows := sqlmock.NewRows([]string{"voter_id", "is_voted"}).AddRow(123, false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `voters`")).WillReturnRows(voterRows)
	
	// แกล้งให้จังหวะ Insert พัง
	mock.ExpectExec("INSERT INTO `votes`").WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := repo.ExecuteVoteTransaction(123, models.Vote{})
	assert.Error(t, err)
	appErr, _ := err.(*AppError)
	assert.Equal(t, 500, appErr.Code)
}

func TestCheckUserHasVoted_Fail_DBError(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	// แกล้งให้จังหวะเช็คข้อมูลโหวตพัง
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `is_voted` FROM `voters`")).
		WithArgs(123, 1).
		WillReturnError(errors.New("db connection lost"))

	isVoted, err := repo.CheckUserHasVoted(123)

	assert.Error(t, err)
	assert.False(t, isVoted)
}

func TestExecuteVoteTransaction_Fail_UpdateVoter(t *testing.T) {
	gormDB, mock := setupMockDB(t)
	repo := NewVotingRepository(gormDB)

	mock.ExpectBegin()
	// ผ่านด่านที่ 1: Select ได้ปกติ
	voterRows := sqlmock.NewRows([]string{"voter_id", "is_voted"}).AddRow(123, false)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `voters`")).WillReturnRows(voterRows)

	// ผ่านด่านที่ 2: Insert คะแนนได้ปกติ
	mock.ExpectExec("INSERT INTO `votes`").WillReturnResult(sqlmock.NewResult(1, 1))

	// พังด่านที่ 3: จังหวะอัปเดตผู้ใช้ ให้พ่น Error ออกมา
	mock.ExpectExec("UPDATE `voters` SET").WillReturnError(errors.New("update voter failed"))
	
	// ต้องมีการ Rollback ยกเลิกทั้งหมด
	mock.ExpectRollback()

	err := repo.ExecuteVoteTransaction(123, models.Vote{})
	assert.Error(t, err)
	appErr, _ := err.(*AppError)
	assert.Equal(t, 500, appErr.Code)
}