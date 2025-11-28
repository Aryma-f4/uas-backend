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

func setupAuthRoutes(r fiber.Router, db *sql.DB) {
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)

	r.Post("/login", func(c *fiber.Ctx) error {
		req := &models.LoginRequest{}
		if err := c.BodyParser(req); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			return helpers.ErrorResponse(c, 422, "validation error", "username and password required")
		}

		user, token, refreshToken, err := authService.Login(c.Context(), req.Username, req.Password)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "login failed", err.Error())
		}

		permissions, _ := userRepo.GetPermissions(c.Context(), user.RoleID)
		role, _ := userRepo.GetRole(c.Context(), user.RoleID)
		roleName := ""
		if role != nil {
			roleName = role.Name
		}

		return helpers.SuccessResponse(c, 200, "login successful", fiber.Map{
			"token":         token,
			"refresh_token": refreshToken,
			"user": fiber.Map{
				"id":          user.ID,
				"username":    user.Username,
				"full_name":   user.FullName,
				"email":       user.Email,
				"role":        roleName,
				"permissions": permissions,
			},
		})
	})

	r.Post("/refresh", func(c *fiber.Ctx) error {
		req := struct {
			RefreshToken string `json:"refresh_token"`
		}{}
		if err := c.BodyParser(&req); err != nil {
			return helpers.ErrorResponse(c, 400, "invalid request format", err.Error())
		}

		if req.RefreshToken == "" {
			return helpers.ErrorResponse(c, 422, "validation error", "refresh_token required")
		}

		userID, roleID, err := authService.ValidateToken(req.RefreshToken)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "invalid refresh token", err.Error())
		}

		token, err := authService.(*service.AuthService).generateToken(userID, roleID)
		if err != nil {
			return helpers.ErrorResponse(c, 500, "token generation failed", err.Error())
		}

		return helpers.SuccessResponse(c, 200, "token refreshed", fiber.Map{
			"token": token,
		})
	})

	r.Get("/profile", middleware.AuthMiddleware(authService), func(c *fiber.Ctx) error {
		userID, err := helpers.GetUserIDFromLocals(c)
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", "user not found")
		}

		user, err := userRepo.GetByID(c.Context(), userID)
		if err != nil {
			return helpers.ErrorResponse(c, 404, "user not found", err.Error())
		}

		permissions, _ := userRepo.GetPermissions(c.Context(), user.RoleID)
		role, _ := userRepo.GetRole(c.Context(), user.RoleID)
		roleName := ""
		if role != nil {
			roleName = role.Name
		}

		return helpers.SuccessResponse(c, 200, "profile retrieved", fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"full_name":   user.FullName,
			"email":       user.Email,
			"role":        roleName,
			"permissions": permissions,
		})
	})
}
