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
	"votespher/pkg"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	db := config.ConnectDB()

	pkg.StartEmailWorker(3)

	if os.Getenv("RUN_MIGRATION") == "true" {
		migration.Run(db)
		return
	}
	if os.Getenv("RUN_SEED") == "true" {
		migration.SeedData(db)
		return
	}
	if os.Getenv("RUN_ADD_TEST_ACCOUNTS") == "true" {
		migration.SeedTestAccounts(db)
		return
	}

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(middleware.RateLimit())

	// ==========================================
	// Dependency Injection
	// ==========================================
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService, db)

	electionRepo := election.NewRepository(db)
	electionSvc := election.NewService(electionRepo)
	electionHandler := election.NewHandler(electionSvc)

	infoRepo := info.NewInfoRepository(db)
	infoService := info.NewInfoService(infoRepo)
	infoHandler := info.NewInfoHandler(infoService)

	resultRepo := result.NewResultRepository(db)
	resultSvc := result.NewResultService(resultRepo)
	resultHandler := result.NewResultHandler(resultSvc)

	realtimeRepo := realtime.NewRealtimeRepository(db)
	realtimeSvc := realtime.NewRealtimeService(realtimeRepo)
	realtimeHandler := realtime.NewRealtimeHandler(realtimeSvc)

	voteRepo := voting.NewVotingRepository(db)
	voteService := voting.NewVotingService(voteRepo)
	voteHandler := voting.NewVotingHandler(voteService)

	// ==========================================
	// Dev endpoints (ENABLE_DEV_ENDPOINTS=true only)
	// ==========================================
	if os.Getenv("ENABLE_DEV_ENDPOINTS") == "true" {
		r.POST("/dev/mock-token", auth.MockTokenHandler())
		log.Println("⚠️  Dev endpoints enabled — DO NOT use in production")
	}

	// ==========================================
	// Public Routes
	// ==========================================
	r.POST("/voter/verify", authHandler.VerifyVoter)
	r.POST("/voter/otp-request", authHandler.OTPRequest)
	r.POST("/voter/otp-confirm", authHandler.OTPConfirm)

	r.GET("/candidates", gin.WrapH(infoHandler.GetCandidatesHandler()))
	r.GET("/parties", gin.WrapH(infoHandler.GetPartiesHandler()))

	r.GET("/election/config", electionHandler.GetConfig)

	r.GET("/results/areas", realtimeHandler.GetAllAreasVotes)
	r.GET("/results/areas/:area_id", resultHandler.GetAreaResult)
	r.GET("/results/provinces/:provinces_name/areas/:area_id", resultHandler.GetProvinceAreaResult)

	// ==========================================
	// Protected Routes (JWT required)
	// ==========================================
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/voter/me", auth.VoterMeHandler(db))
		protected.POST("/ballot/submit",
			middleware.AuditLog(db, "SUBMIT_VOTE"),
			voteHandler.SubmitBallotHandler())
		protected.GET("/ballot/status", voteHandler.GetBallotStatusHandler())
	}

	// ==========================================
	// Admin Routes (JWT + role=admin)
	// ==========================================
	admin := r.Group("/")
	admin.Use(middleware.RequireAuth(), middleware.RequireRole("admin"))
	{
		admin.PATCH("/election/config", electionHandler.UpdateConfig)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s...", port)
	r.Run(":" + port)
}
