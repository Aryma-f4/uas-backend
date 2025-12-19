package usecase

import (
	"context"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/google/uuid"
)

type StudentUsecase struct {
	studentRepo  *repository.StudentRepository
	lecturerRepo *repository.LecturerRepository
}

func NewStudentUsecase(studentRepo *repository.StudentRepository, lecturerRepo *repository.LecturerRepository) *StudentUsecase {
	return &StudentUsecase{
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

func (u *StudentUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Student, error) {
	return u.studentRepo.GetByID(ctx, id)
}

func (u *StudentUsecase) List(ctx context.Context, limit, offset int) ([]*entity.Student, int, error) {
	return u.studentRepo.List(ctx, limit, offset)
}

func (u *StudentUsecase) UpdateAdvisor(ctx context.Context, studentID, advisorID uuid.UUID) error {
	// Verify lecturer exists
	_, err := u.lecturerRepo.GetByID(ctx, advisorID)
	if err != nil {
		return err
	}

	return u.studentRepo.UpdateAdvisor(ctx, studentID, advisorID)
}

func (u *StudentUsecase) GetAdvisees(ctx context.Context, lecturerID uuid.UUID, limit, offset int) ([]*entity.Student, int, error) {
	return u.studentRepo.GetByAdvisorID(ctx, lecturerID, limit, offset)
}
