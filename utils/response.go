package utils

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Meta   Meta        `json:"meta"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func SuccessResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Response{
		Status: "success",
		Data:   data,
	})
}

func SuccessMessageResponse(c *fiber.Ctx, message string) error {
	return c.JSON(Response{
		Status:  "success",
		Message: message,
	})
}

func SuccessWithMessageResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func PaginatedSuccessResponse(c *fiber.Ctx, data interface{}, page, limit, total int) error {
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return c.JSON(PaginatedResponse{
		Status: "success",
		Data:   data,
		Meta: Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Status:  "error",
		Message: message,
	})
}

func BadRequestResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusBadRequest, message)
}

func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusUnauthorized, message)
}

func ForbiddenResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusForbidden, message)
}

func NotFoundResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusNotFound, message)
}

func ConflictResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusConflict, message)
}

func ValidationErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusUnprocessableEntity, message)
}

func InternalServerErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, message)
}
