package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"trash/api/pkg/helpers"

	"github.com/google/uuid"
)

func AuthMiddleware(manager helpers.JwtManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "" {
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			parsedToken, err := manager.Parse(token)
			if err != nil {
				slog.Error("failed to parse token", "error", err)
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if parsedToken == nil {
				slog.Error("parsed token is nil")
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			//tokentype
			if parsedToken.TokenType != "access" {
				slog.Error("token type not access", "type", parsedToken.TokenType)
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
			}
			if parsedToken.Subject == "" {
				slog.Error("parsedtoken subject zero", "subject", parsedToken.Subject)
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			//time
			if time.Now().After(parsedToken.ExpiresAt.Time) {
				slog.Error("token already expires")
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(parsedToken.Subject)
			if err != nil {
				slog.Error("failed to parse userid", "error", err)
				helpers.WriteError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), helpers.UserKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
