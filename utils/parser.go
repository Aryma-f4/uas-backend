package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func ParsePagination(c *fiber.Ctx) (page, limit, offset int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset = (page - 1) * limit
	return page, limit, offset
}

func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}
	return uuid.Parse(userIDStr.(string))
}

func GetRoleIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	roleIDStr := c.Locals("role_id")
	if roleIDStr == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "Role not found")
	}
	return uuid.Parse(roleIDStr.(string))
}

func GetRoleNameFromContext(c *fiber.Ctx) string {
	roleName := c.Locals("role_name")
	if roleName == nil {
		return ""
	}
	return roleName.(string)
}
