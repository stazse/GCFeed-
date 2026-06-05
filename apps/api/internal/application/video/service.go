package applicationvideo

import (
	"context"
	"errors"

	domainvideo "GCFeed/internal/domain/video"
)

var (
	ErrCreateVideoFailed = errors.New("failed to create video")
	ErrLoadVideoFailed   = errors.New("failed to load video")
)

type Service struct {
	repo domainvideo.Repository
}

func New(repo domainvideo.Repository) *Service {
	return &Service{repo: repo}
}

// Create 发布视频。
func (s *Service) Create(ctx context.Context, authorID int64, title, mediaURL, coverURL, idempotencyKey string) (*domainvideo.Video, error) {
	// 第一步：幂等检查（客户端是不是重复提交了？）
	if idempotencyKey != "" {
		existing, err := s.repo.FindByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			return existing, nil // 找到了，直接返回之前的视频
		}
		if !errors.Is(err, domainvideo.ErrVideoNotFound) {
			return nil, errors.Join(ErrCreateVideoFailed, err)
		}
		// 没找到 → 说明是新的请求，继续往下
	}

	// 第二步：用领域规则创建视频实体
	video, err := domainvideo.NewPublished(authorID, title, mediaURL, coverURL, idempotencyKey)
	if err != nil {
		return nil, err
	}

	// 第三步：存入数据库
	if err := s.repo.Save(ctx, video); err != nil {
		return nil, errors.Join(ErrCreateVideoFailed, err)
	}

	return video, nil
}

// GetByID 查看视频详情。
func (s *Service) GetByID(ctx context.Context, id int64) (*domainvideo.Video, error) {
	video, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrLoadVideoFailed, err)
	}
	return video, nil
}

// ListByAuthor 查看某用户的作品列表。
func (s *Service) ListByAuthor(ctx context.Context, authorID int64, cursor int64, limit int) ([]*domainvideo.Video, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.ListByAuthor(ctx, authorID, cursor, limit)
}
