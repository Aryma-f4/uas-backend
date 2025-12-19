package routes

import (
	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/middleware"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router, authUsecase *usecase.AuthUsecase) {
	auth := router.Group("/auth")

	// POST /api/v1/auth/login
	auth.Post("/login", func(c *fiber.Ctx) error {
		var req entity.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if req.Username == "" || req.Password == "" {
			return utils.ValidationErrorResponse(c, "Username and password are required")
		}

		response, err := authUsecase.Login(c.Context(), &req)
		if err != nil {
			return utils.UnauthorizedResponse(c, err.Error())
		}

		return utils.SuccessResponse(c, response)
	})

	// POST /api/v1/auth/refresh
	auth.Post("/refresh", func(c *fiber.Ctx) error {
		var req entity.RefreshTokenRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if req.RefreshToken == "" {
			return utils.ValidationErrorResponse(c, "Refresh token is required")
		}

		token, refreshToken, err := authUsecase.RefreshToken(c.Context(), req.RefreshToken)
		if err != nil {
			return utils.UnauthorizedResponse(c, err.Error())
		}

		return utils.SuccessResponse(c, fiber.Map{
			"token":         token,
			"refresh_token": refreshToken,
		})
	})

	// POST /api/v1/auth/logout
	auth.Post("/logout", func(c *fiber.Ctx) error {
		// In a JWT-based system, logout is handled client-side
		// Server can optionally blacklist the token
		return utils.SuccessMessageResponse(c, "Logged out successfully")
	})

	// GET /api/v1/auth/profile (protected)
	auth.Get("/profile", middleware.AuthMiddleware(authUsecase), func(c *fiber.Ctx) error {
		userID, err := utils.GetUserIDFromContext(c)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid user context")
		}

		profile, err := authUsecase.GetProfile(c.Context(), userID)
		if err != nil {
			return utils.NotFoundResponse(c, "User not found")
		}

		return utils.SuccessResponse(c, profile)
	})
}
