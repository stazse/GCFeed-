package infravideo

import (
	"context"
	"errors"

	domainvideo "GCFeed/internal/domain/video"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Save 保存视频。用事务保证 video 和 video_stat 要么都成功，要么都失败。
func (r *Repository) Save(ctx context.Context, video *domainvideo.Video) error {
	// 开启事务
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先插入 video
		vm := &VideoModel{
			AuthorID:    video.AuthorID,
			Title:       video.Title,
			Description: video.Description,
			MediaURL:    video.MediaURL,
			CoverURL:    video.CoverURL,
			Status:      video.Status,
			PublishedAt: video.PublishedAt,
		}
		// 幂等键为空时存 NULL，避免空字符串触发唯一索引冲突
		if video.IdempotencyKey != "" {
			vm.IdempotencyKey = &video.IdempotencyKey
		}
		if err := tx.Create(vm).Error; err != nil {
			return err
		}

		// 再插入 video_stat（初始计数都是 0）
		sm := &VideoStatModel{VideoID: vm.ID}
		if err := tx.Create(sm).Error; err != nil {
			return err
		}

		video.ID = vm.ID
		return nil
	})
}

// FindByID 按 ID 查视频。
func (r *Repository) FindByID(ctx context.Context, id int64) (*domainvideo.Video, error) {
	var vm VideoModel
	if err := r.db.WithContext(ctx).First(&vm, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainvideo.ErrVideoNotFound
		}
		return nil, err
	}

	var sm VideoStatModel
	// 查统计表（可能不存在，新视频还没统计记录时不算错）
	r.db.WithContext(ctx).Where("video_id = ?", id).First(&sm)

	return restoreVideo(&vm, &sm), nil
}

// FindByIdempotencyKey 按幂等键查找。
func (r *Repository) FindByIdempotencyKey(ctx context.Context, key string) (*domainvideo.Video, error) {
	if key == "" {
		return nil, domainvideo.ErrVideoNotFound
	}
	var vm VideoModel
	if err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&vm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainvideo.ErrVideoNotFound
		}
		return nil, err
	}
	return restoreVideo(&vm, nil), nil
}

// ListByAuthor 按作者分页查视频。
func (r *Repository) ListByAuthor(ctx context.Context, authorID int64, cursor int64, limit int) ([]*domainvideo.Video, error) {
	var models []*VideoModel
	query := r.db.WithContext(ctx).
		Where("author_id = ? AND status = ?", authorID, domainvideo.StatusPublished)

	// cursor 分页：只查 cursor 之后的（published_at 更早的）
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}

	if err := query.Order("id DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	videos := make([]*domainvideo.Video, 0, len(models))
	for _, m := range models {
		videos = append(videos, restoreVideo(m, nil))
	}
	return videos, nil
}

// restoreVideo 把数据库模型转回领域对象。
func restoreVideo(vm *VideoModel, sm *VideoStatModel) *domainvideo.Video {
	v := &domainvideo.Video{
		ID:             vm.ID,
		AuthorID:       vm.AuthorID,
		Title:          vm.Title,
		Description:    vm.Description,
		MediaURL:       vm.MediaURL,
		CoverURL:       vm.CoverURL,
		Status:         vm.Status,
		PublishedAt:    vm.PublishedAt,
		IdempotencyKey: stringPtrToStr(vm.IdempotencyKey),
		CreatedAt:      vm.CreatedAt,
		UpdatedAt:      vm.UpdatedAt,
	}
	if sm != nil {
		v.LikeCount = sm.LikeCount
		v.CommentCount = sm.CommentCount
		v.FavoriteCount = sm.FavoriteCount
	}
	return v
}

// stringPtrToStr 将 *string 安全转为 string（nil → ""）。
func stringPtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

var _ domainvideo.Repository = (*Repository)(nil)
