package routes

import (
	"database/sql"

	"github.com/airlangga/achievement-reporting/helpers"
	"github.com/airlangga/achievement-reporting/middleware"
	"github.com/airlangga/achievement-reporting/models"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/airlangga/achievement-reporting/service"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupStudentRoutes(r fiber.Router, db *sql.DB, mongoClient *mongo.Client) {
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	achievementRepo := repository.NewAchievementRepository(db, mongoClient)
	authService := service.NewAuthService(userRepo)
	studentService := service.NewStudentService(studentRepo)

	r.Use(middleware.AuthMiddleware(authService))

	r.Get("/", func(c *fiber.Ctx) error {
		limit, offset := helpers.GetPaginationParams(c)

		students, err := studentService.ListStudents(c.Context(), limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve students", err.Error())
		}

		return helpers.ListResponse(c, students, len(students), limit, offset)
	})

	r.Get("/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid student id format", err.Error())
		}

		student, err := studentService.GetStudent(c.Context(), id)
		if err != nil {
			return helpers.ErrorResponse(c, 404, "student not found", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "student retrieved", student)
	})

	r.Get("/:id/achievements", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid student id format", err.Error())
		}

		limit, offset := helpers.GetPaginationParams(c)

		achievements, err := achievementRepo.GetStudentAchievements(c.Context(), id, limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve achievements", err.Error())
		}

		return helpers.ListResponse(c, achievements, len(achievements), limit, offset)
	})

	r.Put("/:id/advisor", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid student id format", err.Error())
		}

		req := &models.SetAdvisorRequest{}
		if err := c.BodyParser(req); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		if req.AdvisorID == "" {
			return helpers.ErrorResponse(c, 422, "validation error", "advisor_id required")
		}

		advisorID, err := helpers.ParseUUID(req.AdvisorID)
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid advisor id format", err.Error())
		}

		if err := studentService.SetAdvisor(c.Context(), id, advisorID); err != nil {
			return helpers.ErrorResponse(c, 400, "failed to set advisor", err.Error())
		}

		return helpers.SuccessResponseWithoutData(c, 200, "advisor set successfully")
	})
}
