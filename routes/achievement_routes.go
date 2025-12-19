package routes

import (
	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/middleware"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupAchievementRoutes(router fiber.Router, achievementUsecase *usecase.AchievementUsecase, userRepo *repository.UserRepository, authUsecase *usecase.AuthUsecase) {
	achievements := router.Group("/achievements")
	achievements.Use(middleware.AuthMiddleware(authUsecase))

	// GET /api/v1/achievements - List achievements (filtered by role)
	achievements.Get("/", middleware.RequireAnyPermission(userRepo, "achievement:read", "achievement:verify"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		roleName := utils.GetRoleNameFromContext(c)

		filter := &entity.AchievementFilter{
			Status:          c.Query("status"),
			AchievementType: c.Query("type"),
		}
		filter.Page, filter.Limit, _ = utils.ParsePagination(c)

		achievementList, total, err := achievementUsecase.List(c.Context(), userID, roleName, filter)
		if err != nil {
			return utils.InternalServerErrorResponse(c, err.Error())
		}

		return utils.PaginatedSuccessResponse(c, achievementList, filter.Page, filter.Limit, total)
	})

	// GET /api/v1/achievements/:id - Get achievement detail
	achievements.Get("/:id", middleware.RequireAnyPermission(userRepo, "achievement:read", "achievement:verify"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		achievement, err := achievementUsecase.GetByID(c.Context(), id)
		if err != nil {
			return utils.NotFoundResponse(c, "Achievement not found")
		}

		return utils.SuccessResponse(c, achievement)
	})

	// POST /api/v1/achievements - Create achievement (Mahasiswa only)
	achievements.Post("/", middleware.RequirePermission(userRepo, "achievement:create"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		var req entity.CreateAchievementRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if req.Title == "" {
			return utils.ValidationErrorResponse(c, "Title is required")
		}
		if req.AchievementType == "" {
			return utils.ValidationErrorResponse(c, "Achievement type is required")
		}

		achievement, err := achievementUsecase.Create(c.Context(), userID, &req)
		if err != nil {
			return utils.InternalServerErrorResponse(c, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(utils.Response{
			Status: "success",
			Data:   achievement,
		})
	})

	// PUT /api/v1/achievements/:id - Update achievement (Mahasiswa only, draft/rejected status)
	achievements.Put("/:id", middleware.RequirePermission(userRepo, "achievement:update"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		id := c.Params("id")

		var req entity.UpdateAchievementRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		achievement, err := achievementUsecase.Update(c.Context(), id, userID, &req)
		if err != nil {
			return utils.BadRequestResponse(c, err.Error())
		}

		return utils.SuccessResponse(c, achievement)
	})

	// DELETE /api/v1/achievements/:id - Delete achievement (Mahasiswa only, draft status)
	achievements.Delete("/:id", middleware.RequirePermission(userRepo, "achievement:delete"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		id := c.Params("id")

		if err := achievementUsecase.Delete(c.Context(), id, userID); err != nil {
			return utils.BadRequestResponse(c, err.Error())
		}

		return utils.SuccessMessageResponse(c, "Achievement deleted successfully")
	})

	// POST /api/v1/achievements/:id/submit - Submit for verification (Mahasiswa only)
	achievements.Post("/:id/submit", middleware.RequirePermission(userRepo, "achievement:create"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		id := c.Params("id")

		if err := achievementUsecase.Submit(c.Context(), id, userID); err != nil {
			return utils.BadRequestResponse(c, err.Error())
		}

		return utils.SuccessMessageResponse(c, "Achievement submitted for verification")
	})

	// POST /api/v1/achievements/:id/verify - Verify achievement (Dosen Wali only)
	achievements.Post("/:id/verify", middleware.RequirePermission(userRepo, "achievement:verify"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		id := c.Params("id")

		if err := achievementUsecase.Verify(c.Context(), id, userID); err != nil {
			return utils.BadRequestResponse(c, err.Error())
		}

		return utils.SuccessMessageResponse(c, "Achievement verified successfully")
	})

	// POST /api/v1/achievements/:id/reject - Reject achievement (Dosen Wali only)
	achievements.Post("/:id/reject", middleware.RequirePermission(userRepo, "achievement:reject"), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		id := c.Params("id")

		var req entity.RejectAchievementRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if req.RejectionNote == "" {
			return utils.ValidationErrorResponse(c, "Rejection note is required")
		}

		if err := achievementUsecase.Reject(c.Context(), id, userID, req.RejectionNote); err != nil {
			return utils.BadRequestResponse(c, err.Error())
		}

		return utils.SuccessMessageResponse(c, "Achievement rejected")
	})

	// GET /api/v1/achievements/:id/history - Get status history
	achievements.Get("/:id/history", middleware.RequireAnyPermission(userRepo, "achievement:read", "achievement:verify"), func(c *fiber.Ctx) error {
		id := c.Params("id")

		history, err := achievementUsecase.GetHistory(c.Context(), id)
		if err != nil {
			return utils.NotFoundResponse(c, "Achievement not found")
		}

		return utils.SuccessResponse(c, history)
	})
}
