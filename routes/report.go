package routes

import (
	"database/sql"

	"github.com/airlangga/achievement-reporting/helpers"
	"github.com/airlangga/achievement-reporting/middleware"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/airlangga/achievement-reporting/service"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupReportRoutes(r fiber.Router, db *sql.DB, mongoClient *mongo.Client) {
	userRepo := repository.NewUserRepository(db)
	achievementRepo := repository.NewAchievementRepository(db, mongoClient)
	authService := service.NewAuthService(userRepo)

	r.Use(middleware.AuthMiddleware(authService))

	r.Get("/statistics", func(c *fiber.Ctx) error {
		limit := 100
		offset := 0

		references, err := achievementRepo.ListAll(c.Context(), limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve statistics", err.Error())
		}

		typeCount := make(map[string]int)
		statusCount := make(map[string]int)

		for _, ref := range references {
			statusCount[ref.Status]++

			achievement, err := achievementRepo.GetAchievementByID(c.Context(), ref.MongoAchievementID)
			if err == nil && achievement != nil {
				typeCount[achievement.AchievementType]++
			}
		}

		return helpers.SuccessResponse(c, 200, "statistics retrieved", fiber.Map{
			"total_achievements": len(references),
			"by_type":           typeCount,
			"by_status":         statusCount,
		})
	})

	r.Get("/student/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid student id format", err.Error())
		}

		limit, offset := helpers.GetPaginationParams(c)

		achievements, err := achievementRepo.GetStudentAchievements(c.Context(), id, limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve student achievements", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "student report retrieved", fiber.Map{
			"total":        len(achievements),
			"achievements": achievements,
		})
	})
}
