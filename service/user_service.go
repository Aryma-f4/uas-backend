package service

import (
	"context"
	"errors"

	"github.com/airlangga/achievement-reporting/models"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest, authService *AuthService) (*models.User, error) {
	hashedPassword, err := authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, updates *models.User) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updates.Username != "" {
		user.Username = updates.Username
	}
	if updates.Email != "" {
		user.Email = updates.Email
	}
	if updates.FullName != "" {
		user.FullName = updates.FullName
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.userRepo.List(ctx, limit, offset)
}

func (s *UserService) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	return s.userRepo.GetPermissions(ctx, roleID)
}
