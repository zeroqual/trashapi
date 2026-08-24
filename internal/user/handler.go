package user

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"trash/api/pkg/helpers"
)

type UserHandler struct {
	userService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var dto RequestRegisterUser

	//decode
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		slog.Error("failed to decode json request", "error", err)
		helpers.WriteError(w, "Bad Request", http.StatusBadRequest)
		return
	}

	//validate
	if err := dto.Validate(); err != nil {
		slog.Error("failed to validate json request", "error", err)
		helpers.WriteError(w, "Bad Request", http.StatusBadRequest)
		return
	}

	//create user
	usr, err := h.userService.CreateUser(r.Context(), dto.Email, dto.Password, dto.Name, dto.Surname)
	if err != nil {
		if errors.Is(err.ReturnError, ErrUserAlreadyExists) {
			slog.Warn("user in db already exists", "error", err.OriginalError)
			helpers.WriteError(w, "User already exists", http.StatusConflict)
			return
		}
		if errors.Is(err.OriginalError, context.Canceled) {
			slog.Warn("request canceled")
			return
		}
		slog.Error("internal server error(database)", "error", err.OriginalError)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, map[string]string{
		"id": usr.ID.String(),
	}, http.StatusCreated)
}
