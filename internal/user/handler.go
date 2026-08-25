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
	jwtManager  helpers.JwtManager
}

func NewUserHandler(userService *UserService, jwtManager helpers.JwtManager) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtManager:  jwtManager,
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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var dto RequestLoginUser

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

	//get user

	usr, err := h.userService.GetUser(r.Context(), dto.Email, dto.Password)
	if err != nil {
		slog.Error("failed to get user", "error", err.OriginalError)
		switch err.ReturnError {
		case ErrInternalServer:
			helpers.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		case ErrUserNotFound, ErrInvalidCredentionals:
			helpers.WriteError(w, "invalid credentionals", http.StatusUnauthorized)
			return
		default:
			helpers.WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	//gen tokens
	tokenPair, tokenErr := h.jwtManager.CreateTokenPair(usr.ID, usr.Role)
	if tokenErr != nil {
		slog.Error("failed to create token pair", "error", tokenErr)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, map[string]string{
		"access_token":  tokenPair.AccessToken.Raw,
		"refresh_token": tokenPair.RefreshToken.Raw,
	}, http.StatusOK)
}
