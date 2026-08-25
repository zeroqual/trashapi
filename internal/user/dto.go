package user

import (
	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type RequestRegisterUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

func (s RequestRegisterUser) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required, is.Email),
		validation.Field(&s.Password, validation.Required, validation.Length(5, 20)),
		validation.Field(&s.Name, validation.Required, validation.Length(1, 10)),
		validation.Field(&s.Surname, validation.Required, validation.Length(1, 10)),
	)
}

type RequestLoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s RequestLoginUser) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Email, validation.Required, is.Email),
		validation.Field(&s.Password, validation.Required, validation.Length(5, 20)),
	)
}

type UpdateUserRequest struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

func (u UpdateUserRequest) Validate() error {
	return validation.ValidateStruct(&u,
		validation.Field(&u.Name, validation.Required, validation.Length(2, 20)),
		validation.Field(&u.Surname, validation.Required, validation.Length(2, 20)),
	)
}
