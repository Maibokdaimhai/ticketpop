package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Maibokdaimhai/ticketpop/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type RegisterParams struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
}

type RegisterResult struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (s *UserService) Register(ctx context.Context, params RegisterParams) (*RegisterResult, error) {
	exists, err := s.repo.ExistsEmail(ctx, params.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email นี้ถูกใช้งานแล้ว")
	}

	exists, err = s.repo.ExistsUsername(ctx, params.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username นี้ถูกใช้งานแล้ว")
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// ดึง active status id
	statusID, err := s.repo.GetActiveStatusID(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Email:       params.Email,
		Username:    params.Username,
		DisplayName: params.DisplayName,
		StatusID:    statusID,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateCredential(ctx, userID, string(hashed)); err != nil {
		return nil, err
	}

	return &RegisterResult{
		ID:          userID,
		Email:       params.Email,
		Username:    params.Username,
		DisplayName: params.DisplayName,
	}, nil
}

type LoginParams struct {
	Email    string
	Password string
}

type LoginResult struct {
	Token string `json:"token"`
}

func (s *UserService) Login(ctx context.Context, params LoginParams) (*LoginResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, errors.New("email หรือ password ไม่ถูกต้อง")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)); err != nil {
		return nil, errors.New("email หรือ password ไม่ถูกต้อง")
	}

	// สร้าง JWT token
	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: tokenString}, nil
}
