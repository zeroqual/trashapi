package user

import (
	"context"
	"database/sql"
	"errors"
	"trash/api/pkg/helpers"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrHashPassword         = errors.New("failed to hash password")
	ErrInternalServer       = errors.New("internal server error")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentionals = errors.New("invalid credentionals")
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

func (s *UserService) GetUser(ctx context.Context, email, password string) (*User, *ResponseError) {
	//get user
	usr, err := s.repo.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &ResponseError{
				OriginalError: err,
				ReturnError:   ErrUserNotFound,
			}
		}
		return nil, &ResponseError{
			OriginalError: err,
			ReturnError:   ErrInternalServer,
		}
	}

	//compare psswd
	err = helpers.ComparePasswords(password, usr.PasswordHash)
	if err != nil {
		return nil, &ResponseError{
			OriginalError: err,
			ReturnError:   ErrInvalidCredentionals,
		}
	}

	return usr, nil
}

// заглушка(?)
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.ByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, input UserUpdate, userKeys helpers.UserContext, pageUserID uuid.UUID) error {
	//get user from db
	usr, err := s.repo.ByID(ctx, pageUserID)
	if err != nil {
		return err
	}

	//check permession to edit
	if usr.ID != userKeys.UserID && userKeys.UserRole != "admin" {
		return errors.New("forbidden")
	}

	return s.repo.Update(ctx, input, usr.ID)
}
