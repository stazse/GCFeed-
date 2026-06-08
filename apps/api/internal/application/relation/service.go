package applicationrelation

import (
	"context"
	"errors"

	domainrelation "GCFeed/internal/domain/relation"
)

var (
	ErrFollowFailed   = errors.New("failed to follow")
	ErrUnfollowFailed = errors.New("failed to unfollow")
	ErrListFailed     = errors.New("failed to list follows")
)

type Service struct {
	repo domainrelation.Repository
}

func New(repo domainrelation.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID int64) error {
	if _, err := domainrelation.NewFollow(followerID, followeeID); err != nil {
		return err
	}
	if err := s.repo.Follow(ctx, followerID, followeeID); err != nil {
		return errors.Join(ErrFollowFailed, err)
	}
	return nil
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	if err := s.repo.Unfollow(ctx, followerID, followeeID); err != nil {
		return errors.Join(ErrUnfollowFailed, err)
	}
	return nil
}

func (s *Service) ListFollowing(ctx context.Context, userID int64, cursor int64, limit int) ([]*domainrelation.Follow, error) {
	if limit <= 0 || limit > 50 { limit = 20 }
	return s.repo.ListFollowing(ctx, userID, cursor, limit)
}

func (s *Service) ListFollowers(ctx context.Context, userID int64, cursor int64, limit int) ([]*domainrelation.Follow, error) {
	if limit <= 0 || limit > 50 { limit = 20 }
	return s.repo.ListFollowers(ctx, userID, cursor, limit)
}