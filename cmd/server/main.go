package main

import (
	"log"
	"net/http"
	"votespher/config"
	"votespher/internal/election"
	"votespher/internal/middleware"
	"votespher/migration"
)

func main() {
	// 1. โหลด Environment Variables และเชื่อมต่อ Database
	config.LoadEnv()
	db := config.ConnectDB()

	// 2. รัน migration ทุกครั้งที่ start server
	migration.Run(db)

	// 3. สร้าง HTTP Router (ServeMux)
	mux := http.NewServeMux()

	// ==========================================
	// 🟢 Public Routes (ไม่ต้องใช้ Token)
	// ==========================================
	// ตัวอย่าง (ถ้าคุณมี package auth/info):
	// mux.HandleFunc("/v1/voter/verify", auth.VerifyVoterHandler(db))
	// mux.HandleFunc("/v1/candidates", info.GetCandidatesHandler(db))

	// ==========================================
	// 🟡 Protected Routes (ต้องใช้ Token - สิทธิ์ Voter หรือ Admin)
	// ==========================================
	// ตัวอย่างการครอบเฉพาะ RequireAuth:
	// ballotSubmitHandler := middleware.RequireAuth(voting.SubmitBallotHandler(db))
	// mux.HandleFunc("/v1/ballot/submit", ballotSubmitHandler)

	// ==========================================
	// 🔴 Admin Routes (ต้องใช้ Token และต้องเป็น Role "admin")
	// ==========================================
	// นำ Handler หลักมาครอบด้วย RequireRole และ RequireAuth ตามลำดับ 
	configHandler := election.UpdateConfigHandler(db)
	protectedAdminHandler := middleware.RequireAuth(
		middleware.RequireRole("admin", configHandler),
	)
	
	// ลงทะเบียน Route สำหรับแก้ไข Config
	mux.HandleFunc("/election/config", protectedAdminHandler)

	// ==========================================
	// 4. Start Server
	// ==========================================
	log.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}