package routes

import (
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/app/usecase"
	"github.com/Aryma-f4/uas-backend/middleware"
	"github.com/Aryma-f4/uas-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupReportRoutes(router fiber.Router, achievementUsecase *usecase.AchievementUsecase, studentUsecase *usecase.StudentUsecase, userRepo *repository.UserRepository, authUsecase *usecase.AuthUsecase) {
	reports := router.Group("/reports")
	reports.Use(middleware.AuthMiddleware(authUsecase))

	// GET /api/v1/reports/statistics - Get overall statistics
	reports.Get("/statistics", middleware.RequireAnyPermission(userRepo, "report:read", "report:all"), func(c *fiber.Ctx) error {
		roleName := utils.GetRoleNameFromContext(c)

		var stats interface{}
		var err error

		switch roleName {
		case "Admin":
			// Admin sees all statistics
			stats, err = achievementUsecase.GetStatistics(c.Context(), nil)
		case "Dosen Wali":
			// Dosen Wali sees statistics of their advisees
			userID, _ := utils.GetUserIDFromContext(c)
			lecturer, lerr := userRepo.GetLecturerByUserID(c.Context(), userID)
			if lerr != nil {
				return utils.ForbiddenResponse(c, "Lecturer profile not found")
			}
			
			// Get advisee IDs and calculate stats
			advisees, _, _ := studentUsecase.GetAdvisees(c.Context(), lecturer.ID, 1000, 0)
			if len(advisees) > 0 {
				// Calculate combined stats for all advisees
				stats, err = achievementUsecase.GetStatistics(c.Context(), &advisees[0].ID)
			} else {
				stats = map[string]interface{}{
					"total_achievements": 0,
					"total_verified":     0,
					"total_pending":      0,
					"total_rejected":     0,
				}
			}
		case "Mahasiswa":
			// Mahasiswa sees their own statistics
			userID, _ := utils.GetUserIDFromContext(c)
			student, serr := userRepo.GetStudentByUserID(c.Context(), userID)
			if serr != nil {
				return utils.ForbiddenResponse(c, "Student profile not found")
			}
			stats, err = achievementUsecase.GetStatistics(c.Context(), &student.ID)
		default:
			return utils.ForbiddenResponse(c, "Insufficient permissions")
		}

		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch statistics")
		}

		return utils.SuccessResponse(c, stats)
	})

	// GET /api/v1/reports/student/:id - Get student-specific report
	reports.Get("/student/:id", middleware.RequireAnyPermission(userRepo, "report:read", "report:all"), func(c *fiber.Ctx) error {
		studentID, err := utils.ParseUUID(c.Params("id"))
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid student ID")
		}

		// Get student info
		student, err := studentUsecase.GetByID(c.Context(), studentID)
		if err != nil {
			return utils.NotFoundResponse(c, "Student not found")
		}

		// Get statistics
		stats, err := achievementUsecase.GetStatistics(c.Context(), &studentID)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch statistics")
		}

		// Get achievements
		achievements, _, err := achievementUsecase.ListByStudentID(c.Context(), studentID, 100, 0)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to fetch achievements")
		}

		report := map[string]interface{}{
			"student_info": map[string]interface{}{
				"id":            student.ID,
				"student_id":    student.StudentID,
				"full_name":     student.FullName,
				"email":         student.Email,
				"program_study": student.ProgramStudy,
				"academic_year": student.AcademicYear,
				"advisor_name":  student.AdvisorName,
			},
			"statistics":   stats,
			"achievements": achievements,
		}

		return utils.SuccessResponse(c, report)
	})
}
