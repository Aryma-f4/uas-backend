package usecase

import (
	"context"
	"errors"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementUsecase struct {
	achievementRepo *repository.AchievementRepository
	studentRepo     *repository.StudentRepository
	userRepo        *repository.UserRepository
}

func NewAchievementUsecase(
	achievementRepo *repository.AchievementRepository,
	studentRepo *repository.StudentRepository,
	userRepo *repository.UserRepository,
) *AchievementUsecase {
	return &AchievementUsecase{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
		userRepo:        userRepo,
	}
}

func (u *AchievementUsecase) Create(ctx context.Context, userID uuid.UUID, req *entity.CreateAchievementRequest) (*entity.AchievementResponse, error) {
	// Get student by user ID
	student, err := u.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Calculate points based on achievement type
	points := calculatePoints(req.AchievementType, req.Details)

	// Create achievement in MongoDB
	achievement := &entity.Achievement{
		StudentID:       student.ID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		Points:          points,
		Attachments:     []entity.Attachment{},
	}

	mongoID, err := u.achievementRepo.CreateMongo(ctx, achievement)
	if err != nil {
		return nil, err
	}

	// Create reference in PostgreSQL
	ref := &entity.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          student.ID,
		MongoAchievementID: mongoID.Hex(),
		Status:             entity.StatusDraft,
	}

	if err := u.achievementRepo.CreateReference(ctx, ref); err != nil {
		// Rollback MongoDB insert
		u.achievementRepo.DeleteMongo(ctx, mongoID)
		return nil, err
	}

	// Add status history
	history := &entity.AchievementStatusHistory{
		ID:               uuid.New(),
		AchievementRefID: ref.ID,
		NewStatus:        entity.StatusDraft,
		ChangedBy:        userID,
		Note:             "Achievement created",
	}
	u.achievementRepo.AddStatusHistory(ctx, history)

	return &entity.AchievementResponse{
		ID:              mongoID.Hex(),
		StudentID:       student.ID,
		AchievementType: achievement.AchievementType,
		Title:           achievement.Title,
		Description:     achievement.Description,
		Details:         achievement.Details,
		Attachments:     achievement.Attachments,
		Tags:            achievement.Tags,
		Points:          achievement.Points,
		Status:          entity.StatusDraft,
		CreatedAt:       achievement.CreatedAt,
		UpdatedAt:       achievement.UpdatedAt,
	}, nil
}

func (u *AchievementUsecase) GetByID(ctx context.Context, id string) (*entity.AchievementResponse, error) {
	mongoID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid achievement ID")
	}

	achievement, err := u.achievementRepo.GetMongoByID(ctx, mongoID)
	if err != nil {
		return nil, err
	}

	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get student name
	student, _ := u.studentRepo.GetByID(ctx, achievement.StudentID)
	studentName := ""
	if student != nil {
		studentName = student.FullName
	}

	return &entity.AchievementResponse{
		ID:              id,
		StudentID:       achievement.StudentID,
		StudentName:     studentName,
		AchievementType: achievement.AchievementType,
		Title:           achievement.Title,
		Description:     achievement.Description,
		Details:         achievement.Details,
		Attachments:     achievement.Attachments,
		Tags:            achievement.Tags,
		Points:          achievement.Points,
		Status:          ref.Status,
		SubmittedAt:     ref.SubmittedAt,
		VerifiedAt:      ref.VerifiedAt,
		VerifiedBy:      ref.VerifiedBy,
		RejectionNote:   ref.RejectionNote,
		CreatedAt:       achievement.CreatedAt,
		UpdatedAt:       achievement.UpdatedAt,
	}, nil
}

func (u *AchievementUsecase) Update(ctx context.Context, id string, userID uuid.UUID, req *entity.UpdateAchievementRequest) (*entity.AchievementResponse, error) {
	mongoID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid achievement ID")
	}

	// Check ownership and status
	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ref.Status != entity.StatusDraft && ref.Status != entity.StatusRejected {
		return nil, errors.New("can only update draft or rejected achievements")
	}

	student, err := u.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	if ref.StudentID != student.ID {
		return nil, errors.New("not authorized to update this achievement")
	}

	// Get existing achievement
	achievement, err := u.achievementRepo.GetMongoByID(ctx, mongoID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Title != "" {
		achievement.Title = req.Title
	}
	if req.Description != "" {
		achievement.Description = req.Description
	}
	if req.Details != nil {
		achievement.Details = req.Details
	}
	if req.Tags != nil {
		achievement.Tags = req.Tags
	}

	// Recalculate points
	achievement.Points = calculatePoints(achievement.AchievementType, achievement.Details)

	if err := u.achievementRepo.UpdateMongo(ctx, mongoID, achievement); err != nil {
		return nil, err
	}

	// If rejected, reset to draft
	if ref.Status == entity.StatusRejected {
		u.achievementRepo.UpdateReferenceStatus(ctx, id, entity.StatusDraft, nil, "")
	}

	return u.GetByID(ctx, id)
}

func (u *AchievementUsecase) Delete(ctx context.Context, id string, userID uuid.UUID) error {
	mongoID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid achievement ID")
	}

	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return err
	}

	if ref.Status != entity.StatusDraft {
		return errors.New("can only delete draft achievements")
	}

	student, err := u.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.New("student profile not found")
	}

	if ref.StudentID != student.ID {
		return errors.New("not authorized to delete this achievement")
	}

	// Delete from MongoDB
	if err := u.achievementRepo.DeleteMongo(ctx, mongoID); err != nil {
		return err
	}

	// Delete reference from PostgreSQL
	return u.achievementRepo.DeleteReference(ctx, id)
}

func (u *AchievementUsecase) Submit(ctx context.Context, id string, userID uuid.UUID) error {
	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return err
	}

	if ref.Status != entity.StatusDraft && ref.Status != entity.StatusRejected {
		return errors.New("can only submit draft or rejected achievements")
	}

	student, err := u.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.New("student profile not found")
	}

	if ref.StudentID != student.ID {
		return errors.New("not authorized to submit this achievement")
	}

	// Update status
	if err := u.achievementRepo.UpdateReferenceStatus(ctx, id, entity.StatusSubmitted, nil, ""); err != nil {
		return err
	}

	// Add history
	history := &entity.AchievementStatusHistory{
		ID:               uuid.New(),
		AchievementRefID: ref.ID,
		OldStatus:        ref.Status,
		NewStatus:        entity.StatusSubmitted,
		ChangedBy:        userID,
		Note:             "Submitted for verification",
	}
	u.achievementRepo.AddStatusHistory(ctx, history)

	return nil
}

func (u *AchievementUsecase) Verify(ctx context.Context, id string, verifierID uuid.UUID) error {
	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return err
	}

	if ref.Status != entity.StatusSubmitted {
		return errors.New("can only verify submitted achievements")
	}

	// Verify that the verifier is the student's advisor
	lecturer, err := u.userRepo.GetLecturerByUserID(ctx, verifierID)
	if err != nil {
		return errors.New("verifier is not a lecturer")
	}

	student, err := u.studentRepo.GetByID(ctx, ref.StudentID)
	if err != nil {
		return err
	}

	if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
		return errors.New("not authorized to verify this achievement")
	}

	// Update status
	if err := u.achievementRepo.UpdateReferenceStatus(ctx, id, entity.StatusVerified, &verifierID, ""); err != nil {
		return err
	}

	// Add history
	history := &entity.AchievementStatusHistory{
		ID:               uuid.New(),
		AchievementRefID: ref.ID,
		OldStatus:        ref.Status,
		NewStatus:        entity.StatusVerified,
		ChangedBy:        verifierID,
		Note:             "Achievement verified",
	}
	u.achievementRepo.AddStatusHistory(ctx, history)

	return nil
}

func (u *AchievementUsecase) Reject(ctx context.Context, id string, verifierID uuid.UUID, note string) error {
	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return err
	}

	if ref.Status != entity.StatusSubmitted {
		return errors.New("can only reject submitted achievements")
	}

	// Verify that the verifier is the student's advisor
	lecturer, err := u.userRepo.GetLecturerByUserID(ctx, verifierID)
	if err != nil {
		return errors.New("verifier is not a lecturer")
	}

	student, err := u.studentRepo.GetByID(ctx, ref.StudentID)
	if err != nil {
		return err
	}

	if student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
		return errors.New("not authorized to reject this achievement")
	}

	// Update status
	if err := u.achievementRepo.UpdateReferenceStatus(ctx, id, entity.StatusRejected, nil, note); err != nil {
		return err
	}

	// Add history
	history := &entity.AchievementStatusHistory{
		ID:               uuid.New(),
		AchievementRefID: ref.ID,
		OldStatus:        ref.Status,
		NewStatus:        entity.StatusRejected,
		ChangedBy:        verifierID,
		Note:             note,
	}
	u.achievementRepo.AddStatusHistory(ctx, history)

	return nil
}

func (u *AchievementUsecase) GetHistory(ctx context.Context, id string) ([]*entity.AchievementStatusHistory, error) {
	ref, err := u.achievementRepo.GetReferenceByMongoID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u.achievementRepo.GetStatusHistory(ctx, ref.ID)
}

func (u *AchievementUsecase) List(ctx context.Context, userID uuid.UUID, roleName string, filter *entity.AchievementFilter) ([]*entity.AchievementResponse, int, error) {
	limit := 10
	offset := 0
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	var refs []*entity.AchievementReference
	var total int
	var err error

	switch roleName {
	case "Mahasiswa":
		// Get student's own achievements
		student, err := u.studentRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, 0, errors.New("student profile not found")
		}
		refs, total, err = u.achievementRepo.ListReferences(ctx, &student.ID, filter.Status, limit, offset)
		if err != nil {
			return nil, 0, err
		}

	case "Dosen Wali":
		// Get achievements of advisees
		lecturer, err := u.userRepo.GetLecturerByUserID(ctx, userID)
		if err != nil {
			return nil, 0, errors.New("lecturer profile not found")
		}
		
		students, _, err := u.studentRepo.GetByAdvisorID(ctx, lecturer.ID, 1000, 0)
		if err != nil {
			return nil, 0, err
		}

		studentIDs := make([]uuid.UUID, len(students))
		for i, s := range students {
			studentIDs[i] = s.ID
		}

		refs, total, err = u.achievementRepo.ListReferencesByStudentIDs(ctx, studentIDs, filter.Status, limit, offset)
		if err != nil {
			return nil, 0, err
		}

	case "Admin":
		// Get all achievements
		refs, total, err = u.achievementRepo.ListReferences(ctx, nil, filter.Status, limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	// Fetch achievement details from MongoDB
	var achievements []*entity.AchievementResponse
	for _, ref := range refs {
		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue
		}

		achievement, err := u.achievementRepo.GetMongoByID(ctx, mongoID)
		if err != nil {
			continue
		}

		// Apply type filter
		if filter.AchievementType != "" && string(achievement.AchievementType) != filter.AchievementType {
			continue
		}

		student, _ := u.studentRepo.GetByID(ctx, achievement.StudentID)
		studentName := ""
		if student != nil {
			studentName = student.FullName
		}

		achievements = append(achievements, &entity.AchievementResponse{
			ID:              ref.MongoAchievementID,
			StudentID:       achievement.StudentID,
			StudentName:     studentName,
			AchievementType: achievement.AchievementType,
			Title:           achievement.Title,
			Description:     achievement.Description,
			Details:         achievement.Details,
			Attachments:     achievement.Attachments,
			Tags:            achievement.Tags,
			Points:          achievement.Points,
			Status:          ref.Status,
			SubmittedAt:     ref.SubmittedAt,
			VerifiedAt:      ref.VerifiedAt,
			VerifiedBy:      ref.VerifiedBy,
			RejectionNote:   ref.RejectionNote,
			CreatedAt:       achievement.CreatedAt,
			UpdatedAt:       achievement.UpdatedAt,
		})
	}

	return achievements, total, nil
}

func (u *AchievementUsecase) ListByStudentID(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*entity.AchievementResponse, int, error) {
	refs, total, err := u.achievementRepo.ListReferences(ctx, &studentID, "", limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var achievements []*entity.AchievementResponse
	for _, ref := range refs {
		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue
		}

		achievement, err := u.achievementRepo.GetMongoByID(ctx, mongoID)
		if err != nil {
			continue
		}

		student, _ := u.studentRepo.GetByID(ctx, achievement.StudentID)
		studentName := ""
		if student != nil {
			studentName = student.FullName
		}

		achievements = append(achievements, &entity.AchievementResponse{
			ID:              ref.MongoAchievementID,
			StudentID:       achievement.StudentID,
			StudentName:     studentName,
			AchievementType: achievement.AchievementType,
			Title:           achievement.Title,
			Description:     achievement.Description,
			Details:         achievement.Details,
			Attachments:     achievement.Attachments,
			Tags:            achievement.Tags,
			Points:          achievement.Points,
			Status:          ref.Status,
			SubmittedAt:     ref.SubmittedAt,
			VerifiedAt:      ref.VerifiedAt,
			VerifiedBy:      ref.VerifiedBy,
			RejectionNote:   ref.RejectionNote,
			CreatedAt:       achievement.CreatedAt,
			UpdatedAt:       achievement.UpdatedAt,
		})
	}

	return achievements, total, nil
}

func (u *AchievementUsecase) GetStatistics(ctx context.Context, studentID *uuid.UUID) (*entity.StatisticsResponse, error) {
	stats, err := u.achievementRepo.GetStatistics(ctx, studentID)
	if err != nil {
		return nil, err
	}

	// Get type statistics from MongoDB
	var studentIDs []uuid.UUID
	if studentID != nil {
		studentIDs = []uuid.UUID{*studentID}
	} else {
		// Get all student IDs
		students, _, _ := u.studentRepo.List(ctx, 10000, 0)
		for _, s := range students {
			studentIDs = append(studentIDs, s.ID)
		}
	}

	if len(studentIDs) > 0 {
		typeStats, err := u.achievementRepo.GetStatsMongo(ctx, studentIDs)
		if err == nil {
			stats.ByType = typeStats
		}
	}

	return stats, nil
}

// Helper function to calculate points
func calculatePoints(achievementType entity.AchievementType, details map[string]interface{}) int {
	basePoints := map[entity.AchievementType]int{
		entity.TypeCompetition:   100,
		entity.TypePublication:   80,
		entity.TypeCertification: 60,
		entity.TypeOrganization:  50,
		entity.TypeAcademic:      70,
		entity.TypeOther:         30,
	}

	points := basePoints[achievementType]

	// Add bonus based on level for competitions
	if achievementType == entity.TypeCompetition {
		if level, ok := details["competitionLevel"].(string); ok {
			switch level {
			case "international":
				points += 100
			case "national":
				points += 50
			case "regional":
				points += 25
			}
		}
		if rank, ok := details["rank"].(float64); ok {
			if rank == 1 {
				points += 50
			} else if rank == 2 {
				points += 30
			} else if rank == 3 {
				points += 20
			}
		}
	}

	return points
}
