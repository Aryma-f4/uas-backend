package usecase

import (
	"context"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/google/uuid"
)

type UserUsecase struct {
	userRepo     *repository.UserRepository
	studentRepo  *repository.StudentRepository
	lecturerRepo *repository.LecturerRepository
	authUsecase  *AuthUsecase
}

func NewUserUsecase(
	userRepo *repository.UserRepository,
	studentRepo *repository.StudentRepository,
	lecturerRepo *repository.LecturerRepository,
	authUsecase *AuthUsecase,
) *UserUsecase {
	return &UserUsecase{
		userRepo:     userRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
		authUsecase:  authUsecase,
	}
}

func (u *UserUsecase) CreateUser(ctx context.Context, req *entity.CreateUserRequest) (*entity.User, error) {
	
	hashedPassword, err := u.authUsecase.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	
	user := &entity.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	
	role, err := u.userRepo.GetRoleByName(ctx, "Mahasiswa")
	if err == nil && role.ID == req.RoleID && req.StudentID != "" {
		student := &entity.Student{
			ID:           uuid.New(),
			UserID:       user.ID,
			StudentID:    req.StudentID,
			ProgramStudy: req.ProgramStudy,
			AcademicYear: req.AcademicYear,
		}
		if err := u.studentRepo.Create(ctx, student); err != nil {
			
			u.userRepo.Delete(ctx, user.ID)
			return nil, err
		}
	}

	dosenRole, err := u.userRepo.GetRoleByName(ctx, "Dosen Wali")
	if err == nil && dosenRole.ID == req.RoleID && req.LecturerID != "" {
		lecturer := &entity.Lecturer{
			ID:         uuid.New(),
			UserID:     user.ID,
			LecturerID: req.LecturerID,
			Department: req.Department,
		}
		if err := u.lecturerRepo.Create(ctx, lecturer); err != nil {
			u.userRepo.Delete(ctx, user.ID)
			return nil, err
		}
	}

	return user, nil
}

func (u *UserUsecase) GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *UserUsecase) UpdateUser(ctx context.Context, id uuid.UUID, req *entity.UpdateUserRequest) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *UserUsecase) ListUsers(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	return u.userRepo.List(ctx, limit, offset)
}

func (u *UserUsecase) UpdateUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return u.userRepo.UpdateRole(ctx, userID, roleID)
}

func (u *UserUsecase) GetRoles(ctx context.Context) ([]*entity.Role, error) {
	return u.userRepo.GetRoles(ctx)
}

func (u *UserUsecase) CheckPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	return u.userRepo.CheckPermission(ctx, roleID, permission)
}
