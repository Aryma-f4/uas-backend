package routes

import (
	"database/sql"

	"github.com/airlangga/achievement-reporting/helpers"
	"github.com/airlangga/achievement-reporting/middleware"
	"github.com/airlangga/achievement-reporting/models"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/airlangga/achievement-reporting/service"
	"github.com/gofiber/fiber/v2"
)

func setupUserRoutes(r fiber.Router, db *sql.DB) {
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo)

	r.Use(middleware.AuthMiddleware(authService))

	r.Get("/", func(c *fiber.Ctx) error {
		limit, offset := helpers.GetPaginationParams(c)

		users, err := userService.ListUsers(c.Context(), limit, offset)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "failed to retrieve users", err.Error())
		}

		return helpers.ListResponse(c, users, len(users), limit, offset)
	})

	r.Get("/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid user id format", err.Error())
		}

		user, err := userService.GetUser(c.Context(), id)
		if err != nil {
			return helpers.ErrorResponse(c, 404, "user not found", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "user retrieved", user)
	})

	r.Post("/", func(c *fiber.Ctx) error {
		req := &models.CreateUserRequest{}
		if err := c.BodyParser(req); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		// Validate input
		if req.Username == "" || req.Email == "" || req.Password == "" {
			return helpers.ErrorResponse(c, 422, "validation error", "username, email, and password required")
		}

		if !helpers.ValidateEmail(req.Email) {
			return helpers.ErrorResponse(c, 422, "validation error", "invalid email format")
		}

		if !helpers.ValidatePassword(req.Password) {
			return helpers.ErrorResponse(c, 422, "validation error", "password must be at least 8 chars with uppercase, lowercase, digit, and special char")
		}

		user, err := userService.CreateUser(c.Context(), req, authService)
		if err != nil {
			return helpers.ErrorResponse(c, 400, "failed to create user", err.Error())
		}

		return helpers.SuccessResponse(c, 201, "user created successfully", user)
	})

	r.Put("/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid user id format", err.Error())
		}

		updates := &models.User{}
		if err := c.BodyParser(updates); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		user, err := userService.UpdateUser(c.Context(), id, updates)
		if err != nil {
			return helpers.ErrorResponse(c, 400, "failed to update user", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "user updated successfully", user)
	})

	r.Delete("/:id", func(c *fiber.Ctx) error {
		id, err := helpers.GetUUIDFromParams(c, "id")
		if err != nil {
			return helpers.ErrorResponse(c, 400, "invalid user id format", err.Error())
		}

		if err := userService.DeleteUser(c.Context(), id); err != nil {
			return helpers.ErrorResponse(c, 400, "failed to delete user", err.Error())
		}

		return helpers.SuccessResponseWithoutData(c, 200, "user deleted successfully")
	})
}
