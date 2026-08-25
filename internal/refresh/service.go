package refresh

import (
	"context"
	"time"
	"trash/api/pkg/helpers"

	"github.com/google/uuid"
)

type RefreshService struct {
	refreshRepo RefreshRepository
}

func NewRefreshService(refreshRepo RefreshRepository) *RefreshService {
	return &RefreshService{
		refreshRepo: refreshRepo,
	}
}

func (s *RefreshService) CreateRefresh(ctx context.Context, userID uuid.UUID, tokenString string, expiresAt time.Time) (*Refresh, error) {
	//hash token
	tokenHash := helpers.HashToken(tokenString)

	refresh, err := s.refreshRepo.Create(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return nil, err
	}
	return refresh, nil
}

func (s *RefreshService) RefreshTokens(ctx context.Context, tokenString string, userID uuid.UUID, newRefreshString string, expiresAt time.Time) error {
	tokenHash := helpers.HashToken(tokenString)

	//get refersh
	oldRefresh, err := s.refreshRepo.ByPrimary(ctx, userID, tokenHash)
	if err != nil {
		//todo revoke all user tokens
		return err
	}

	//hash new refreshString

	newTokenHash := helpers.HashToken(newRefreshString)

	//revoke refresh token(tokenString)

	//add new refresh token(for userID)

	return s.refreshRepo.RefreshRotateTokens(ctx, oldRefresh, newTokenHash, expiresAt)
}
