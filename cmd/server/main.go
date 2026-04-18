package main

import (
	"log"
	"os"
	"votespher/config"
	"votespher/internal/auth"
	"votespher/internal/election"
	"votespher/internal/voting"
	"votespher/internal/middleware"
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

	// ==========================================
	// 🟢 Public Routes (ไม่ต้องใช้ Token)
	// ==========================================
	
	// แก้เป็น r.POST และเอา gin.WrapH ออก เพราะเป็น Gin Handler แล้ว
	r.POST("/dev/mock-token", auth.MockTokenHandler()) 
	
	// r.GET("/v1/voter/verify", auth.VerifyVoterHandler(db))
	// r.GET("/v1/candidates", info.GetCandidatesHandler(db))

	// ==========================================
	// 🟡 Protected Routes (ต้องใช้ Token - สิทธิ์ Voter หรือ Admin)
	// ==========================================
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.POST("/ballot/submit", voting.SubmitBallotHandler(db)) // เช็คชื่อฟังก์ชันให้ตรงกับที่คุณตั้งใน voting/handler.go นะครับ
		
		protected.GET("/ballot/status", voting.GetBallotStatusHandler(db)) // ฟังก์ชันนี้จะรวมสถานะระบบและสถานะผู้ใช้เข้าด้วยกัน
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