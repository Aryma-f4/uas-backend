package routes

import (
	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/middleware"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupStudentRoutes(router fiber.Router, studentUsecase *usecase.StudentUsecase, achievementUsecase *usecase.AchievementUsecase, userRepo *repository.UserRepository, authUsecase *usecase.AuthUsecase) {
	students := router.Group("/students")
	students.Use(middleware.AuthMiddleware(authUsecase))

	// GET /api/v1/students - List all students
	students.Get("/", middleware.RequireAnyPermission(userRepo, "student:read", "student:manage"), func(c *fiber.Ctx) error {
		page, limit, offset := utils.ParsePagination(c)

		studentList, total, err := studentUsecase.List(c.Context(), limit, offset)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch students")
		}

		return utils.PaginatedSuccessResponse(c, studentList, page, limit, total)
	})

	// GET /api/v1/students/:id - Get student by ID
	students.Get("/:id", middleware.RequireAnyPermission(userRepo, "student:read", "student:manage"), func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid student ID")
		}

		student, err := studentUsecase.GetByID(c.Context(), id)
		if err != nil {
			return utils.NotFoundResponse(c, "Student not found")
		}

		return utils.SuccessResponse(c, student)
	})

	// GET /api/v1/students/:id/achievements - Get student's achievements
	students.Get("/:id/achievements", middleware.RequireAnyPermission(userRepo, "student:read", "achievement:read"), func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid student ID")
		}

		page, limit, offset := utils.ParsePagination(c)

		achievementList, total, err := achievementUsecase.ListByStudentID(c.Context(), id, limit, offset)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch achievements")
		}

		return utils.PaginatedSuccessResponse(c, achievementList, page, limit, total)
	})

	// PUT /api/v1/students/:id/advisor - Update student's advisor (Admin only)
	students.Put("/:id/advisor", middleware.RequirePermission(userRepo, "student:manage"), func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid student ID")
		}

		var req entity.UpdateAdvisorRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if err := studentUsecase.UpdateAdvisor(c.Context(), id, req.AdvisorID); err != nil {
			return utils.InternalServerErrorResponse(c, err.Error())
		}

		return utils.SuccessMessageResponse(c, "Advisor updated successfully")
	})
}
