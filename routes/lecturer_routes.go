package routes

import (
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/middleware"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupLecturerRoutes(router fiber.Router, lecturerUsecase *usecase.LecturerUsecase, studentUsecase *usecase.StudentUsecase, userRepo *repository.UserRepository, authUsecase *usecase.AuthUsecase) {
	lecturers := router.Group("/lecturers")
	lecturers.Use(middleware.AuthMiddleware(authUsecase))

	// GET /api/v1/lecturers - List all lecturers
	lecturers.Get("/", middleware.RequireAnyPermission(userRepo, "lecturer:read", "lecturer:manage"), func(c *fiber.Ctx) error {
		page, limit, offset := utils.ParsePagination(c)

		lecturerList, total, err := lecturerUsecase.List(c.Context(), limit, offset)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch lecturers")
		}

		return utils.PaginatedSuccessResponse(c, lecturerList, page, limit, total)
	})

	// GET /api/v1/lecturers/:id - Get lecturer by ID
	lecturers.Get("/:id", middleware.RequireAnyPermission(userRepo, "lecturer:read", "lecturer:manage"), func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid lecturer ID")
		}

		lecturer, err := lecturerUsecase.GetByID(c.Context(), id)
		if err != nil {
			return utils.NotFoundResponse(c, "Lecturer not found")
		}

		return utils.SuccessResponse(c, lecturer)
	})

	// GET /api/v1/lecturers/:id/advisees - Get lecturer's advisees
	lecturers.Get("/:id/advisees", middleware.RequireAnyPermission(userRepo, "lecturer:read", "student:read"), func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid lecturer ID")
		}

		page, limit, offset := utils.ParsePagination(c)

		advisees, total, err := lecturerUsecase.GetAdvisees(c.Context(), id, limit, offset)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch advisees")
		}

		return utils.PaginatedSuccessResponse(c, advisees, page, limit, total)
	})
}
