package routes

import (
	"database/sql"

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
	achievementRepo := repository.NewAchievementRepository(db, mongoDB)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg)
	userUsecase := usecase.NewUserUsecase(userRepo, studentRepo, lecturerRepo, authUsecase)
	achievementUsecase := usecase.NewAchievementUsecase(achievementRepo, studentRepo, userRepo)
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
