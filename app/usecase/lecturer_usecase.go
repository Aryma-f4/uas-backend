package usecase

import (
	"context"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/google/uuid"
)

type LecturerUsecase struct {
	lecturerRepo *repository.LecturerRepository
	studentRepo  *repository.StudentRepository
}

func NewLecturerUsecase(lecturerRepo *repository.LecturerRepository, studentRepo *repository.StudentRepository) *LecturerUsecase {
	return &LecturerUsecase{
		lecturerRepo: lecturerRepo,
		studentRepo:  studentRepo,
	}
}

func (u *LecturerUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Lecturer, error) {
	return u.lecturerRepo.GetByID(ctx, id)
}

func (u *LecturerUsecase) List(ctx context.Context, limit, offset int) ([]*entity.Lecturer, int, error) {
	return u.lecturerRepo.List(ctx, limit, offset)
}

func (u *LecturerUsecase) GetAdvisees(ctx context.Context, lecturerID uuid.UUID, limit, offset int) ([]*entity.Student, int, error) {
	return u.studentRepo.GetByAdvisorID(ctx, lecturerID, limit, offset)
}
