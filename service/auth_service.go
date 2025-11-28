package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/airlangga/achievement-reporting/models"
	"github.com/airlangga/achievement-reporting/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID, user.RoleID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, token, refreshToken, nil
}

func (s *AuthService) generateToken(userID, roleID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role_id": roleID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *AuthService) generateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
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

func (s *AuthService) HashPassword(password string) (string, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
