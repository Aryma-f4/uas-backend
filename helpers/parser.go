package helpers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParseUUID parses UUID from string
func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// GetPaginationParams extracts limit and offset from query
func GetPaginationParams(c *fiber.Ctx) (limit int, offset int) {
	limit = 10
	offset = 0
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 100 {
				parsed = 100
			}
			limit = parsed
		}
	}
	
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	return limit, offset
}

// GetUUIDFromParams gets UUID from URL parameters
func GetUUIDFromParams(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	return ParseUUID(c.Params(paramName))
}

// GetUserIDFromLocals retrieves user ID from context
func GetUserIDFromLocals(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fiber.NewError(401, "unauthorized")
	}
	return userID, nil
}
