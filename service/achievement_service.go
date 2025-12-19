package service

import (
	"context"
	"errors"
	"time"

	"github.com/Aryma-f4/uas-backend/models"
	"github.com/Aryma-f4/uas-backend/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService struct {
	achievementRepo *repository.AchievementRepository
	studentRepo     *repository.StudentRepository
	userRepo        *repository.UserRepository
}

func NewAchievementService(
	achievementRepo *repository.AchievementRepository,
	studentRepo *repository.StudentRepository,
	userRepo *repository.UserRepository,
) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
		userRepo:        userRepo,
	}
}

func (s *AchievementService) CreateAchievement(ctx context.Context, studentID uuid.UUID, req *models.CreateAchievementRequest) (*models.Achievement, error) {
	achievement := &models.Achievement{
		ID:              primitive.NewObjectID(),
		StudentID:       studentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.achievementRepo.CreateAchievement(ctx, achievement); err != nil {
		return nil, err
	}

	// Create reference in PostgreSQL
	ref := &models.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.achievementRepo.CreateReference(ctx, ref); err != nil {
		return nil, err
	}

	return achievement, nil
}

func (s *AchievementService) GetAchievementDetail(ctx context.Context, referenceID uuid.UUID) (*models.Achievement, error) {
	ref, err := s.achievementRepo.GetReferenceByID(ctx, referenceID)
	if err != nil {
		return nil, err
	}

	objectID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return nil, err
	}

	return s.achievementRepo.GetAchievementByID(ctx, objectID)
}

func (s *AchievementService) SubmitForVerification(ctx context.Context, referenceID uuid.UUID) error {
	return s.achievementRepo.SubmitForVerification(ctx, referenceID)
}

func (s *AchievementService) VerifyAchievement(ctx context.Context, referenceID uuid.UUID, verifiedBy uuid.UUID) error {
	return s.achievementRepo.VerifyAchievement(ctx, referenceID, verifiedBy)
}

func (s *AchievementService) RejectAchievement(ctx context.Context, referenceID uuid.UUID, verifiedBy uuid.UUID, note string) error {
	return s.achievementRepo.RejectAchievement(ctx, referenceID, verifiedBy, note)
}

func (s *AchievementService) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*models.AchievementReference, error) {
	return s.achievementRepo.GetStudentAchievements(ctx, studentID, limit, offset)
}

func (s *AchievementService) DeleteAchievement(ctx context.Context, referenceID uuid.UUID, studentID uuid.UUID) error {
	ref, err := s.achievementRepo.GetReferenceByID(ctx, referenceID)
	if err != nil {
		return err
	}

	if ref.StudentID != studentID {
		return errors.New("not authorized to delete this achievement")
	}

	if ref.Status != "draft" {
		return errors.New("can only delete draft achievements")
	}

	objectID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return err
	}

	return s.achievementRepo.DeleteAchievement(ctx, objectID, referenceID)
}
