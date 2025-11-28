package routes

import (
	"database/sql"

	"github.com/airlangga/achievement-reporting/helpers"
	"github.com/airlangga/achievement-reporting/middleware"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/airlangga/achievement-reporting/service"
	"github.com/gofiber/fiber/v2"
)

func setupLecturerRoutes(r fiber.Router, db *sql.DB) {
	userRepo := repository.NewUserRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	authService := service.NewAuthService(userRepo)

	r.Use(middleware.AuthMiddleware(authService))

	r.Get("/", func(c *fiber.Ctx) error {
		limit, offset := helpers.GetPaginationParams(c)

		lecturers, err := lecturerRepo.List(c.Context(), limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve lecturers", err.Error())
		}

		return helpers.ListResponse(c, lecturers, len(lecturers), limit, offset)
	})

	r.Get("/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid lecturer id format", err.Error())
		}

		lecturer, err := lecturerRepo.GetByID(c.Context(), id)
		if err != nil {
			return helpers.ErrorResponse(c, 404, "lecturer not found", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "lecturer retrieved", lecturer)
	})

	r.Get("/:id/advisees", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid lecturer id format", err.Error())
		}

		advisees, err := studentRepo.GetAdvisees(c.Context(), id)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve advisees", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "advisees retrieved", advisees)
	})
}
