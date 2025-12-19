package routes

import (
	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/middleware"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(router fiber.Router, userUsecase *usecase.UserUsecase, userRepo *repository.UserRepository, authUsecase *usecase.AuthUsecase) {
	users := router.Group("/users")
	
	// All user routes require authentication and Admin role
	users.Use(middleware.AuthMiddleware(authUsecase))
	users.Use(middleware.RequireRole(userRepo, "Admin"))

	// GET /api/v1/users
	users.Get("/", func(c *fiber.Ctx) error {
		page, limit, offset := utils.ParsePagination(c)

		userList, total, err := userUsecase.ListUsers(c.Context(), limit, offset)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch users")
		}

		return utils.PaginatedSuccessResponse(c, userList, page, limit, total)
	})

	// GET /api/v1/users/roles
	users.Get("/roles", func(c *fiber.Ctx) error {
		roles, err := userUsecase.GetRoles(c.Context())
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch roles")
		}

		return utils.SuccessResponse(c, roles)
	})

	// GET /api/v1/users/:id
	users.Get("/:id", func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid user ID")
		}

		user, err := userUsecase.GetUser(c.Context(), id)
		if err != nil {
			return utils.NotFoundResponse(c, "User not found")
		}

		return utils.SuccessResponse(c, user)
	})

	// POST /api/v1/users
	users.Post("/", func(c *fiber.Ctx) error {
		var req entity.CreateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		// Validate required fields
		if req.Username == "" {
			return utils.ValidationErrorResponse(c, "Username is required")
		}
		if req.Email == "" {
			return utils.ValidationErrorResponse(c, "Email is required")
		}
		if req.Password == "" {
			return utils.ValidationErrorResponse(c, "Password is required")
		}
		if req.FullName == "" {
			return utils.ValidationErrorResponse(c, "Full name is required")
		}

		if !utils.ValidateEmail(req.Email) {
			return utils.ValidationErrorResponse(c, "Invalid email format")
		}

		if valid, msg := utils.ValidatePassword(req.Password); !valid {
			return utils.ValidationErrorResponse(c, msg)
		}

		user, err := userUsecase.CreateUser(c.Context(), &req)
		if err != nil {
			return utils.ConflictResponse(c, err.Error())
		}

		return c.Status(fiber.StatusCreated).JSON(utils.Response{
			Status: "success",
			Data:   user,
		})
	})

	// PUT /api/v1/users/:id
	users.Put("/:id", func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid user ID")
		}

		var req entity.UpdateUserRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if req.Email != "" && !utils.ValidateEmail(req.Email) {
			return utils.ValidationErrorResponse(c, "Invalid email format")
		}

		user, err := userUsecase.UpdateUser(c.Context(), id, &req)
		if err != nil {
			return utils.NotFoundResponse(c, "User not found")
		}

		return utils.SuccessResponse(c, user)
	})

	// DELETE /api/v1/users/:id
	users.Delete("/:id", func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid user ID")
		}

		if err := userUsecase.DeleteUser(c.Context(), id); err != nil {
			return utils.NotFoundResponse(c, "User not found")
		}

		return utils.SuccessMessageResponse(c, "User deleted successfully")
	})

	// PUT /api/v1/users/:id/role
	users.Put("/:id/role", func(c *fiber.Ctx) error {
		id, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid user ID")
		}

		var req entity.UpdateRoleRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body")
		}

		if err := userUsecase.UpdateUserRole(c.Context(), id, req.RoleID); err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to update user role")
		}

		return utils.SuccessMessageResponse(c, "User role updated successfully")
	})
}
