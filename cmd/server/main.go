package main

import (
	"log"
	"os"
	"votespher/config"
	"votespher/internal/auth"
	"votespher/internal/election"
	"votespher/internal/info"
	"votespher/internal/middleware"
	"votespher/internal/realtime"
	"votespher/internal/result"
	"votespher/internal/voting"
	"votespher/migration"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. โหลด Environment Variables และเชื่อมต่อ Database
	config.LoadEnv()
	db := config.ConnectDB()

	// 2. ตรวจ flag ก่อน run migration
	if os.Getenv("RUN_MIGRATION") == "true" {
		migration.Run(db)
		return
	}

	// 3. รัน Data Seeding (ใส่ข้อมูลจำลอง 20 รายการ)
	if os.Getenv("RUN_SEED") == "true" {
		migration.SeedData(db)
		return
	}

	// 4. สร้าง HTTP Router ด้วย Gin
	r := gin.Default()

	// Refactor Layered Architecture
	voteRepo := voting.NewVotingRepository(db)
    voteService := voting.NewVotingService(voteRepo)
    voteHandler := voting.NewVotingHandler(voteService)

	// ==========================================
	// 🟢 Public Routes (ไม่ต้องใช้ Token)
	// ==========================================

	// แก้เป็น r.POST และเอา gin.WrapH ออก เพราะเป็น Gin Handler แล้ว
	r.POST("/dev/mock-token", auth.MockTokenHandler())

	// --- API สำหรับระบบยืนยันตัวตนผู้มีสิทธิ์เลือกตั้ง ---
	// ตรวจสอบเลขบัตรประชาชน 13 หลัก ว่ามีสิทธิ์โหวตหรือไม่
	r.POST("/voter/verify", auth.VerifyVoterHandler(db))

	// ขอรับรหัส OTP 6 หลัก เพื่อนำไปใช้ยืนยันการเข้าระบบ
	r.POST("/voter/otp-request", auth.OTPRequestHandler(db))

	r.GET("/candidates", gin.WrapH(info.GetCandidatesHandler(db)))

	r.GET("/parties", gin.WrapH(info.GetPartiesHandler(db)))

	r.GET("/results/provinces/:provinces_name/areas/:area_id", result.GetProvinceAreaResultHandler(db))
	// เพิ่ม API สำหรับผลโหวตแบบเรียลไทม์
	r.GET("/results/areas", realtime.GetAllAreasVotesHandler(db))
	//เพิ่ทม API สำหรับผลโหวตแบบเรียลไทม์แยกตามเขต
	r.GET("/results/areas/:area_id", realtime.GetVoteResultByAreaHandler(db))

	// ==========================================
	// 🟡 Protected Routes (ต้องใช้ Token - สิทธิ์ Voter หรือ Admin)
	// ==========================================
	protected := r.Group("/")
    protected.Use(middleware.RequireAuth())
    {
        protected.POST("/ballot/submit", voteHandler.SubmitBallotHandler())
        protected.GET("/ballot/status", voteHandler.GetBallotStatusHandler())
    }

	// ==========================================
	// 🔴 Admin Routes (ต้องใช้ Token และต้องเป็น Role "admin")
	// ==========================================
	admin := r.Group("/")
	admin.Use(middleware.RequireAuth(), middleware.RequireRole("admin"))
	{
		admin.PATCH("/election/config", election.UpdateConfigHandler(db)) // ใช้ PATCH หรือ PUT ตามที่คุณออกแบบไว้
	}

	// Start Server
	log.Println("Server is running on port 8080...")
	r.Run(":8080")
}
