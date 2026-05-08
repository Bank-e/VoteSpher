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
	config.LoadEnv()
	db := config.ConnectDB()

	if os.Getenv("RUN_MIGRATION") == "true" {
		migration.Run(db)
		return
	}

	if os.Getenv("RUN_SEED") == "true" {
		migration.SeedData(db)
		return
	}

	r := gin.Default()

<<<<<<< HEAD
	// ==========================================
	// 📌 Dependency Injection (DI) Setup
	// ==========================================

	// 👉 DI ของ Election (เพื่อนทำเสร็จแล้ว ใช้ได้ปกติ)
	electionRepo := election.NewRepository(db)
	electionSvc := election.NewService(electionRepo)
	electionHandler := election.NewHandler(electionSvc)

	// 👉 DI ของ Auth ที่ฟูจิเพิ่ง Refactor เสร็จ (สมบูรณ์ 100%)
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	// ❌ เอา DI ของ voting, info, result ออกไป เพราะไฟล์ของเพื่อนยังไม่รองรับระบบนี้

	// ==========================================
	// 📌 Public Routes
	// ==========================================

	// ✅ Auth Routes (ใช้ระบบใหม่ที่ฟูจิทำ ไม่มีเส้น MockToken แล้ว)
	r.POST("/voter/verify", authHandler.VerifyVoter)
	r.POST("/voter/otp-request", authHandler.OTPRequest)
	r.POST("/voter/otp-confirm", authHandler.OTPConfirm)

	// ⚠️ Info & Result Routes (กลับไปใช้โค้ดแบบเก่าของเพื่อนไปก่อน จะได้ไม่ Error)
	r.GET("/candidates", gin.WrapH(info.GetCandidatesHandler()))
	r.GET("/parties", gin.WrapH(info.GetPartiesHandler()))

	r.GET("/results/area/:id", result.GetAreaResultHandler(db))
	r.GET("/results/areas", realtime.GetAllAreasVotesHandler(db))
	r.GET("/results/areas/:area_id", realtime.GetVoteResultByAreaHandler(db))

	// ==========================================
	// 📌 Protected Routes (Require Login)
	// ==========================================
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		// ⚠️ Voting Routes (กลับไปใช้โค้ดแบบเก่าของเพื่อน)
		protected.POST("/ballot/submit", voting.SubmitBallotHandler(db))
		protected.GET("/ballot/status", voting.GetBallotStatusHandler(db))
	}

	// ==========================================
	// 📌 Admin Routes
	// ==========================================
=======
	voteRepo := voting.NewVotingRepository(db)
	voteService := voting.NewVotingService(voteRepo)
	voteHandler := voting.NewVotingHandler(voteService)

	infoRepo := info.NewInfoRepository(db)
	infoService := info.NewInfoService(infoRepo)
	infoHandler := info.NewInfoHandler(infoService)

	resultRepo := result.NewResultRepository(db)
	resultService := result.NewResultService(resultRepo)
	resultHandler := result.NewResultHandler(resultService)

	electionRepo := election.NewRepository(db)
	electionSvc := election.NewService(electionRepo)
	electionHandler := election.NewHandler(electionSvc)

	r.POST("/dev/mock-token", auth.MockTokenHandler())

	r.POST("/voter/verify", auth.VerifyVoterHandler(db))
	r.POST("/voter/otp-request", auth.OTPRequestHandler(db))

	r.GET("/candidates", gin.WrapH(infoHandler.GetCandidatesHandler()))
	r.GET("/parties", gin.WrapH(infoHandler.GetPartiesHandler()))

	r.GET("/results/area/:id", resultHandler.GetAreaResult)
	r.GET("/results/areas", realtime.GetAllAreasVotesHandler(db))
	r.GET("/results/areas/:area_id", realtime.GetVoteResultByAreaHandler(db))

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.POST("/ballot/submit", voteHandler.SubmitBallotHandler())
		protected.GET("/ballot/status", voteHandler.GetBallotStatusHandler())
	}

>>>>>>> e4a160189aca1a94396bb88f87428eb0df414f09
	admin := r.Group("/")
	admin.Use(middleware.RequireAuth(), middleware.RequireRole("admin"))
	{
		admin.PATCH("/election/config", electionHandler.UpdateConfig)
	}

	log.Println("Server is running on port 8080...")
	r.Run(":8080")
}
