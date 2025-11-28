package middleware

import (
	"strings"

	"github.com/airlangga/achievement-reporting/helpers"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/airlangga/achievement-reporting/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthContext struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helpers.ErrorResponse(c, 401, "unauthorized", "missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return helpers.ErrorResponse(c, 401, "unauthorized", "invalid authorization format")
		}

		userID, roleID, err := authService.ValidateToken(parts[1])
		if err != nil {
			return helpers.ErrorResponse(c, 401, "unauthorized", "invalid token")
		}

		c.Locals("userID", userID)
		c.Locals("roleID", roleID)

		return c.Next()
	}
}

// RequireRole middleware checks if user has specific role
func RequireRole(userRepo *repository.UserRepository, requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID, ok := c.Locals("roleID").(uuid.UUID)
		if !ok {
			return helpers.ErrorResponse(c, 401, "unauthorized", "invalid role context")
		}

		role, err := userRepo.GetRole(c.Context(), roleID)
		if err != nil || role == nil {
			return helpers.ErrorResponse(c, 403, "forbidden", "user role not found")
		}

		hasRole := false
		for _, required := range requiredRoles {
			if role.Name == required {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return helpers.ErrorResponse(c, 403, "forbidden", "insufficient role permissions")
		}

		c.Locals("role", role.Name)
		return c.Next()
	}
}

// RequirePermission middleware checks if user has specific permission
func RequirePermission(userRepo *repository.UserRepository, requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return helpers.ErrorResponse(c, 401, "unauthorized", "invalid user context")
		}

		permissions, err := userRepo.GetPermissions(c.Context(), userID)
		if err != nil {
			return helpers.ErrorResponse(c, 403, "forbidden", "unable to fetch permissions")
		}

		hasPermission := false
		for _, perm := range permissions {
			if perm == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return helpers.ErrorResponse(c, 403, "forbidden", "insufficient permissions for this action")
		}

		c.Locals("permissions", permissions)
		return c.Next()
	}
}

// RequireAnyPermission checks if user has any of the required permissions
func RequireAnyPermission(userRepo *repository.UserRepository, requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return helpers.ErrorResponse(c, 401, "unauthorized", "invalid user context")
		}

		permissions, err := userRepo.GetPermissions(c.Context(), userID)
		if err != nil {
			return helpers.ErrorResponse(c, 403, "forbidden", "unable to fetch permissions")
		}

		hasPermission := false
		for _, required := range requiredPermissions {
			for _, perm := range permissions {
				if perm == required {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			return helpers.ErrorResponse(c, 403, "forbidden", "insufficient permissions for this action")
		}

		c.Locals("permissions", permissions)
		return c.Next()
	}
}

// RequireAllPermissions checks if user has all required permissions
func RequireAllPermissions(userRepo *repository.UserRepository, requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uuid.UUID)
		if !ok {
			return helpers.ErrorResponse(c, 401, "unauthorized", "invalid user context")
		}

		permissions, err := userRepo.GetPermissions(c.Context(), userID)
		if err != nil {
			return helpers.ErrorResponse(c, 403, "forbidden", "unable to fetch permissions")
		}

		for _, required := range requiredPermissions {
			found := false
			for _, perm := range permissions {
				if perm == required {
					found = true
					break
				}
			}
			if !found {
				return helpers.ErrorResponse(c, 403, "forbidden", "insufficient permissions for this action")
			}
		}

		c.Locals("permissions", permissions)
		return c.Next()
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				helpers.ErrorResponse(c, 500, "internal server error", "unexpected error occurred")
			}
		}()
		return c.Next()
	}
}
