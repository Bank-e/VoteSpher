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

	// ==========================================
	// Dependency Injection (DI) Setup
	// ==========================================
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	electionRepo := election.NewRepository(db)
	electionSvc := election.NewService(electionRepo)
	electionHandler := election.NewHandler(electionSvc)

	infoRepo := info.NewInfoRepository(db)
	infoService := info.NewInfoService(infoRepo)
	infoHandler := info.NewInfoHandler(infoService)

	resultRepo := result.NewResultRepository(db)
	resultService := result.NewResultService(resultRepo)
	resultHandler := result.NewResultHandler(resultService)

	voteRepo := voting.NewVotingRepository(db)
	voteService := voting.NewVotingService(voteRepo)
	voteHandler := voting.NewVotingHandler(voteService)
	
	// ==========================================
	// Public Routes
	// ==========================================
	r.POST("/voter/verify", authHandler.VerifyVoter)
	r.POST("/voter/otp-request", authHandler.OTPRequest)
	r.POST("/voter/otp-confirm", authHandler.OTPConfirm)

	r.GET("/candidates", gin.WrapH(infoHandler.GetCandidatesHandler()))
	r.GET("/parties", gin.WrapH(infoHandler.GetPartiesHandler()))

	r.GET("/results/area/:id", resultHandler.GetAreaResult)
	r.GET("/results/areas", realtime.GetAllAreasVotesHandler(db))
	r.GET("/results/areas/:area_id", realtime.GetVoteResultByAreaHandler(db))

	// ==========================================
	// Protected Routes (Require Login)
	// ==========================================
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.POST("/ballot/submit", voteHandler.SubmitBallotHandler())
		protected.GET("/ballot/status", voteHandler.GetBallotStatusHandler())
	}

	// ==========================================
	// Admin Routes
	// ==========================================
	admin := r.Group("/")
	admin.Use(middleware.RequireAuth(), middleware.RequireRole("admin"))
	{
		admin.PATCH("/election/config", electionHandler.UpdateConfig)
	}

	log.Println("Server is running on port 8080...")
	r.Run(":8080")
}
