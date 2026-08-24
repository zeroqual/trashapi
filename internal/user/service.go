package user

import (
	"context"
	"errors"
	"trash/api/pkg/helpers"

	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrHashPassword      = errors.New("failed to hash password")
	ErrInternalServer    = errors.New("internal server error")
)

type ResponseError struct {
	OriginalError error
	ReturnError   error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email, password, name, surname string) (*User, *ResponseError) {
	//hash password
	passwordHash := helpers.HashPassword(password)
	if passwordHash == "" {
		return nil, &ResponseError{OriginalError: ErrInternalServer, ReturnError: ErrInternalServer}
	}

	// add to db
	usr, err := s.repo.Create(ctx, email, passwordHash, name, surname)
	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) {
			if pqError.Code == "23505" {
				return nil, &ResponseError{OriginalError: pqError, ReturnError: ErrUserAlreadyExists}
			}
		}
		return nil, &ResponseError{OriginalError: err, ReturnError: ErrInternalServer}
	}

	return usr, nil
}
