package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Initialize database
	db, err := InitPostgres()
	if err != nil {
		log.Fatalf("Failed to connect PostgreSQL: %v", err)
	}
	defer db.Close()

	mongoClient, err := InitMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(nil); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	// Run migrations
	if err := RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup routes
	setupRoutes(app, db, mongoClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func setupRoutes(app *fiber.App, db interface{}, mongoClient interface{}) {
	v1 := app.Group("/api/v1")

	// Auth routes
	authRoutes := v1.Group("/auth")
	setupAuthRoutes(authRoutes, db)

	// User routes (Admin only)
	userRoutes := v1.Group("/users")
	setupUserRoutes(userRoutes, db)

	// Achievement routes
	achievementRoutes := v1.Group("/achievements")
	setupAchievementRoutes(achievementRoutes, db, mongoClient)

	// Student routes
	studentRoutes := v1.Group("/students")
	setupStudentRoutes(studentRoutes, db, mongoClient)

	// Lecturer routes
	lecturerRoutes := v1.Group("/lecturers")
	setupLecturerRoutes(lecturerRoutes, db)

	// Report routes
	reportRoutes := v1.Group("/reports")
	setupReportRoutes(reportRoutes, db, mongoClient)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
