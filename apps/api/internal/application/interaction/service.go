package applicationinteraction

import (
	"context"
	"errors"

	domaininteraction "GCFeed/internal/domain/interaction"
)

var (
	ErrSetActionFailed    = errors.New("failed to set action")
	ErrCreateCommentFailed = errors.New("failed to create comment")
	ErrListCommentsFailed  = errors.New("failed to list comments")
	ErrDeleteCommentFailed = errors.New("failed to delete comment")
)

type Service struct {
	repo domaininteraction.Repository
}

func New(repo domaininteraction.Repository) *Service {
	return &Service{repo: repo}
}

// SetLike 点赞（active=true）或取消点赞（active=false）。
func (s *Service) SetLike(ctx context.Context, userID, videoID int64, active bool, idempotencyKey string) (bool, int, error) {
	isActive, likeCount, _, err := s.repo.SetAction(ctx, userID, videoID, domaininteraction.ActionLike, active, idempotencyKey)
	return isActive, likeCount, err
}

// SetFavorite 收藏或取消收藏。
func (s *Service) SetFavorite(ctx context.Context, userID, videoID int64, active bool, idempotencyKey string) (bool, int, error) {
	isActive, _, favCount, err := s.repo.SetAction(ctx, userID, videoID, domaininteraction.ActionFavorite, active, idempotencyKey)
	return isActive, favCount, err
}

// CreateComment 发表评论。
func (s *Service) CreateComment(ctx context.Context, videoID, userID int64, content, idempotencyKey string) (*domaininteraction.Comment, error) {
	comment, err := domaininteraction.NewComment(videoID, userID, content, idempotencyKey)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveComment(ctx, comment); err != nil {
		return nil, errors.Join(ErrCreateCommentFailed, err)
	}
	return comment, nil
}

// ListComments 评论列表。
func (s *Service) ListComments(ctx context.Context, videoID int64, cursor int64, limit int) ([]*domaininteraction.Comment, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.ListComments(ctx, videoID, cursor, limit)
}

// DeleteComment 删除评论。
func (s *Service) DeleteComment(ctx context.Context, commentID, userID int64) error {
	if err := s.repo.DeleteComment(ctx, commentID, userID); err != nil {
		return errors.Join(ErrDeleteCommentFailed, err)
	}
	return nil
}