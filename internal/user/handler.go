package user

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"trash/api/internal/refresh"
	"trash/api/pkg/helpers"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService    *UserService
	refreshService *refresh.RefreshService
	jwtManager     helpers.JwtManager
}

func NewUserHandler(userService *UserService, jwtManager helpers.JwtManager, refreshService *refresh.RefreshService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		refreshService: refreshService,
		jwtManager:     jwtManager,
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

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	//get cookies
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	} else {
		//todo  /  refresh token from body? idk if it really need to do(
	}

	if refreshToken == "" {
		slog.Error("empty refresh token", "error", err)
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	//parse token
	token, err := h.jwtManager.Parse(refreshToken)
	if err != nil {
		slog.Error("failed to parse refresh token", "error", err)
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	//maybe this is better to do in service?
	if token.TokenType != "refresh" {
		slog.Error("wrong token type", "type", token.TokenType)
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(token.Subject)
	if err != nil {
		slog.Error("failed to parse userID", "error", err)
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	//get user
	usr, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get user from db", "error", err)
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	//generate new tokens
	tokenPair, err := h.jwtManager.CreateTokenPair(usr.ID, usr.Role)
	if err != nil {
		slog.Error("failed to create token pair", "error", err)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//insert into db refresh tokens
	expiresAt, expiresErr := tokenPair.RefreshToken.Claims.GetExpirationTime()
	if expiresErr != nil {
		slog.Error("failed to get expires token t ime", "error", expiresErr)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//make tx to db
	err = h.refreshService.RefreshTokens(r.Context(), refreshToken, usr.ID, tokenPair.RefreshToken.Raw, expiresAt.Time)
	if err != nil {
		slog.Error("failed to make tx refresh tokens", "error", err)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken.Raw,
		HttpOnly: false,
		Secure:   false,
		// Path:     "/api/auth/refresh",
		MaxAge: int(24 * 60 * 30 * time.Hour),
	})
	helpers.WriteResponse(w, map[string]string{
		"access_token": tokenPair.AccessToken.Raw,
	}, http.StatusOK)
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

	//insert into db refresh tokens
	expiresAt, expiresErr := tokenPair.RefreshToken.Claims.GetExpirationTime()
	if expiresErr != nil {
		slog.Error("failed to get expires token t ime", "error", expiresErr)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, refreshErr := h.refreshService.CreateRefresh(r.Context(), usr.ID, tokenPair.RefreshToken.Raw, expiresAt.Time)
	if refreshErr != nil {
		slog.Error("failed to create refresh token", "error", refreshErr)
		helpers.WriteError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	//
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken.Raw,
		HttpOnly: false,
		Secure:   false,
		// Path:     "/api/auth/refresh",
		MaxAge: int(24 * 60 * 30 * time.Hour),
	})

	helpers.WriteResponse(w, map[string]string{
		"access_token": tokenPair.AccessToken.Raw,
	}, http.StatusOK)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userKeys, ok := r.Context().Value(helpers.UserKey).(helpers.UserContext)
	if !ok {
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	helpers.WriteResponse(w, map[string]string{
		"user_id": userKeys.UserID.String(),
		"role":    userKeys.UserRole,
	}, http.StatusOK)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var dto UpdateUserRequest
	userKeys, ok := r.Context().Value(helpers.UserKey).(helpers.UserContext)
	if !ok {
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	//get id from requests
	user_id := chi.URLParam(r, "id")
	if user_id == "" {
		slog.Error("userid cannot be empty")
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user_idUUID, err := uuid.Parse(user_id)
	if err != nil {
		slog.Error("user not found")
		helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	//decode requests
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&dto); err != nil {
		slog.Error("failed to decode req", "error", err)
		helpers.WriteError(w, "bad request", http.StatusBadRequest)
		return
	}

	//validate
	if err := dto.Validate(); err != nil {
		slog.Error("failed to validate req", "error", err)
		helpers.WriteError(w, "bad request", http.StatusBadRequest)
		return
	}

	err = h.userService.UpdateUser(r.Context(), UserUpdate{
		Name:    dto.Name,
		Surname: dto.Surname,
	}, userKeys, user_idUUID)
	if err != nil {
		slog.Error("failed to update user", "error", err)
		helpers.WriteError(w, "forbidden", http.StatusForbidden)
		return
	}

	helpers.WriteResponse(w, map[string]string{
		"success": "true",
	}, http.StatusCreated)
}
