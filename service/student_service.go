package service

import (
	"context"

	"github.com/airlangga/achievement-reporting/models"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/google/uuid"
)

type StudentService struct {
	studentRepo *repository.StudentRepository
}

func NewStudentService(studentRepo *repository.StudentRepository) *StudentService {
	return &StudentService{studentRepo: studentRepo}
}

func (s *StudentService) CreateStudent(ctx context.Context, student *models.Student) error {
	return s.studentRepo.Create(ctx, student)
}

func (s *StudentService) GetStudent(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	return s.studentRepo.GetByID(ctx, id)
}

func (s *StudentService) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*models.Student, error) {
	return s.studentRepo.GetByUserID(ctx, userID)
}

func (s *StudentService) SetAdvisor(ctx context.Context, studentID, advisorID uuid.UUID) error {
	return s.studentRepo.SetAdvisor(ctx, studentID, advisorID)
}

func (s *StudentService) GetAdvisees(ctx context.Context, advisorID uuid.UUID) ([]*models.Student, error) {
	return s.studentRepo.GetAdvisees(ctx, advisorID)
}

func (s *StudentService) ListStudents(ctx context.Context, limit, offset int) ([]*models.Student, error) {
	return s.studentRepo.List(ctx, limit, offset)
}
