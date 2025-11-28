package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CustomLogger middleware for request logging
func CustomLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate elapsed time
		elapsed := time.Since(start)

		// Log request details
		fmt.Printf("[%s] %s %s - %d - %v\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			elapsed,
		)

		return err
	}
}
