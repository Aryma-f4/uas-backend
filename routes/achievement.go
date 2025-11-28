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

func setupAchievementRoutes(r fiber.Router, db *sql.DB, mongoClient *mongo.Client) {
	userRepo := repository.NewUserRepository(db)
	achievementRepo := repository.NewAchievementRepository(db, mongoClient)
	studentRepo := repository.NewStudentRepository(db)
	authService := service.NewAuthService(userRepo)
	achievementService := service.NewAchievementService(achievementRepo, studentRepo, userRepo)

	r.Use(middleware.AuthMiddleware(authService))

	r.Get("/", func(c *fiber.Ctx) error {
		limit, offset := helpers.GetPaginationParams(c)

		references, err := achievementRepo.ListAll(c.Context(), limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve achievements", err.Error())
		}

		return helpers.ListResponse(c, references, len(references), limit, offset)
	})

	r.Get("/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid achievement id format", err.Error())
		}

		achievement, err := achievementService.GetAchievementDetail(c.Context(), id)
		if err != nil {
			return helpers.ErrorResponse(c, 404, "achievement not found", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "achievement retrieved", achievement)
	})

	r.Post("/", func(c *fiber.Ctx) error {
		userID, err := helpers.GetUserIDFromLocals(c)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", err.Error())
		}

		student, err := studentRepo.GetByUserID(c.Context(), userID)
		if err != nil {
			return helpers.ErrorResponse(c, 403, "forbidden", "only students can create achievements")
		}

		req := &models.CreateAchievementRequest{}
		if err := c.BodyParser(req); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		if req.Title == "" || req.AchievementType == "" {
			return helpers.ErrorResponse(c, 422, "validation error", "title and achievement_type required")
		}

		achievement, err := achievementService.CreateAchievement(c.Context(), student.ID, req)
		if err != nil {
			return helpers.ErrorResponse(c, 400, "failed to create achievement", err.Error())
		}

		return helpers.SuccessResponse(c, 201, "achievement created successfully", achievement)
	})

	r.Post("/:id/submit", func(c *fiber.Ctx) error {
		userID, err := helpers.GetUserIDFromLocals(c)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", err.Error())
		}

		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid achievement id format", err.Error())
		}

		ref, err := achievementRepo.GetReferenceByID(c.Context(), id)
		if err != nil {
			return helpers.ErrorResponse(c, 404, "achievement not found", err.Error())
		}

		student, _ := studentRepo.GetByID(c.Context(), ref.StudentID)
		if student.UserID != userID {
			return helpers.ErrorResponse(c, 403, "forbidden", "not authorized to submit this achievement")
		}

		if err := achievementService.SubmitForVerification(c.Context(), id); err != nil {
			return helpers.ErrorResponse(c, 400, "failed to submit achievement", err.Error())
		}

		return helpers.SuccessResponseWithoutData(c, 200, "achievement submitted for verification")
	})

	r.Post("/:id/verify", func(c *fiber.Ctx) error {
		userID, err := helpers.GetUserIDFromLocals(c)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", err.Error())
		}

		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid achievement id format", err.Error())
		}

		if err := achievementService.VerifyAchievement(c.Context(), id, userID); err != nil {
			return helpers.ErrorResponse(c, 400, "failed to verify achievement", err.Error())
		}

		return helpers.SuccessResponseWithoutData(c, 200, "achievement verified successfully")
	})

	r.Post("/:id/reject", func(c *fiber.Ctx) error {
		userID, err := helpers.GetUserIDFromLocals(c)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", err.Error())
		}

		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid achievement id format", err.Error())
		}

		req := struct {
			Note string `json:"note"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		if req.Note == "" {
			return helpers.ErrorResponse(c, 422, "validation error", "note is required")
		}

		if err := achievementService.RejectAchievement(c.Context(), id, userID, req.Note); err != nil {
			return helpers.ErrorResponse(c, 400, "failed to reject achievement", err.Error())
		}

		return helpers.SuccessResponseWithoutData(c, 200, "achievement rejected successfully")
	})

	r.Delete("/:id", func(c *fiber.Ctx) error {
		userID, err := helpers.GetUserIDFromLocals(c)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", err.Error())
		}

		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid achievement id format", err.Error())
		}

		student, _ := studentRepo.GetByUserID(c.Context(), userID)
		if err := achievementService.DeleteAchievement(c.Context(), id, student.ID); err != nil {
			return helpers.ErrorResponse(c, 400, "failed to delete achievement", err.Error())
		}

		return helpers.SuccessResponseWithoutData(c, 200, "achievement deleted successfully")
	})
}
