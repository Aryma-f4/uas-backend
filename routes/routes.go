package routes

import (
	"database/sql"
	"log"

	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/config"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(app *fiber.App, db *sql.DB, mongoDB *mongo.Database, cfg *config.Config) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	
	// PERBAIKAN: Guard mongoDB != nil sebelum membuat achievementRepo
	var achievementRepo *repository.AchievementRepository
	if mongoDB != nil {
		achievementRepo = repository.NewAchievementRepository(db, mongoDB)
	} else {
		log.Println("[WARN] mongoDB is nil — running in test/stub mode")
		// Tetap set achievementRepo ke nil; usecase dan handler harus handle nil case
		achievementRepo = nil
	}

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg)
	userUsecase := usecase.NewUserUsecase(userRepo, studentRepo, lecturerRepo, authUsecase)
	
	// PERBAIKAN: Pass nil achievementRepo jika mongoDB nil
	achievementUsecase := usecase. NewAchievementUsecase(achievementRepo, studentRepo, userRepo)
	studentUsecase := usecase.NewStudentUsecase(studentRepo, lecturerRepo)
	lecturerUsecase := usecase.NewLecturerUsecase(lecturerRepo, studentRepo)

	// API v1 group
	api := app.Group("/api/v1")

	// Setup route groups
	SetupAuthRoutes(api, authUsecase)
	SetupUserRoutes(api, userUsecase, userRepo, authUsecase)
	SetupAchievementRoutes(api, achievementUsecase, userRepo, authUsecase)
	SetupStudentRoutes(api, studentUsecase, achievementUsecase, userRepo, authUsecase)
	SetupLecturerRoutes(api, lecturerUsecase, studentUsecase, userRepo, authUsecase)
	SetupReportRoutes(api, achievementUsecase, studentUsecase, userRepo, authUsecase)
}
