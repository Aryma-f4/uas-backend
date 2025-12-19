package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/Aryma-f4/uas-backend/app/repository"
	"github.com/Aryma-f4/uas-backend/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthUsecase(userRepo *repository.UserRepository, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, req *entity.LoginRequest) (*entity.LoginResponse, error) {
	
	user, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		user, err = u.userRepo.GetByEmail(ctx, req.Username)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}
	}

	
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	
	permissions, err := u.userRepo.GetPermissions(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	
	token, err := u.generateToken(user.ID, user.RoleID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &entity.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User: entity.UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Email:       user.Email,
			Role:        user.RoleName,
			Permissions: permissions,
		},
	}, nil
}

func (u *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	
	token, err := jwt.ParseWithClaims(refreshToken, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.config.JWTSecret), nil
	})
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		return "", "", errors.New("invalid token type")
	}

	
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", "", errors.New("invalid user id")
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	
	newToken, err := u.generateToken(user.ID, user.RoleID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := u.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return newToken, newRefreshToken, nil
}

func (u *AuthUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.UserInfo, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissions, err := u.userRepo.GetPermissions(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &entity.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Email:       user.Email,
		Role:        user.RoleName,
		Permissions: permissions,
	}, nil
}

func (u *AuthUsecase) ValidateToken(tokenString string) (uuid.UUID, uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.config.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	roleID, err := uuid.Parse(claims["role_id"].(string))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return userID, roleID, nil
}

func (u *AuthUsecase) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *AuthUsecase) generateToken(userID, roleID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role_id": roleID.String(),
		"exp":     time.Now().Add(time.Duration(u.config.JWTExpireHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.config.JWTSecret))
}

func (u *AuthUsecase) generateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Duration(u.config.JWTRefreshExpHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.config.JWTSecret))
}
