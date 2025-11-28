package helpers

import (
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse sends a successful response with data
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "success",
		"message": message,
		"data": data,
	})
}

// SuccessResponseWithoutData sends success response without data
func SuccessResponseWithoutData(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "success",
		"message": message,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, details interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "error",
		"message": message,
		"details": details,
	})
}

// ListResponse sends a paginated list response
func ListResponse(c *fiber.Ctx, data interface{}, total int, limit int, offset int) error {
	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"data": data,
		"pagination": fiber.Map{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
	})
}
