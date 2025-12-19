package middleware

import (
	"strings"

	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func AuthMiddleware(authUsecase *usecase.AuthUsecase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.UnauthorizedResponse(c, "Invalid authorization format")
		}

		userID, roleID, err := authUsecase.ValidateToken(parts[1])
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid or expired token")
		}

		c.Locals("user_id", userID.String())
		c.Locals("role_id", roleID.String())

		return c.Next()
	}
}

func RequireRole(userRepo *repository.UserRepository, allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleIDStr := c.Locals("role_id")
		if roleIDStr == nil {
			return utils.UnauthorizedResponse(c, "Role not found in context")
		}

		roleID, err := uuid.Parse(roleIDStr.(string))
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid role ID")
		}

		// Get user's role name
		roles, err := userRepo.GetRoles(c.Context())
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch roles")
		}

		var userRoleName string
		for _, role := range roles {
			if role.ID == roleID {
				userRoleName = role.Name
				break
			}
		}

		if userRoleName == "" {
			return utils.ForbiddenResponse(c, "User role not found")
		}

		// Check if user's role is in allowed roles
		allowed := false
		for _, allowedRole := range allowedRoles {
			if userRoleName == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			return utils.ForbiddenResponse(c, "Insufficient role permissions")
		}

		c.Locals("role_name", userRoleName)
		return c.Next()
	}
}

func RequirePermission(userRepo *repository.UserRepository, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleIDStr := c.Locals("role_id")
		if roleIDStr == nil {
			return utils.UnauthorizedResponse(c, "Role not found in context")
		}

		roleID, err := uuid.Parse(roleIDStr.(string))
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid role ID")
		}

		hasPermission, err := userRepo.CheckPermission(c.Context(), roleID, permission)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to check permissions")
		}

		if !hasPermission {
			return utils.ForbiddenResponse(c, "Insufficient permissions")
		}

		// Also get and set role name
		roles, _ := userRepo.GetRoles(c.Context())
		for _, role := range roles {
			if role.ID == roleID {
				c.Locals("role_name", role.Name)
				break
			}
		}

		return c.Next()
	}
}

func RequireAnyPermission(userRepo *repository.UserRepository, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleIDStr := c.Locals("role_id")
		if roleIDStr == nil {
			return utils.UnauthorizedResponse(c, "Role not found in context")
		}

		roleID, err := uuid.Parse(roleIDStr.(string))
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid role ID")
		}

		for _, permission := range permissions {
			hasPermission, err := userRepo.CheckPermission(c.Context(), roleID, permission)
			if err == nil && hasPermission {
				// Also get and set role name
				roles, _ := userRepo.GetRoles(c.Context())
				for _, role := range roles {
					if role.ID == roleID {
						c.Locals("role_name", role.Name)
						break
					}
				}
				return c.Next()
			}
		}

		return utils.ForbiddenResponse(c, "Insufficient permissions")
	}
}
