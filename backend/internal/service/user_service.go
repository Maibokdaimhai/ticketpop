package service

import (
	"context"
	"errors"

	"github.com/Maibokdaimhai/ticketpop/internal/repository"
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
