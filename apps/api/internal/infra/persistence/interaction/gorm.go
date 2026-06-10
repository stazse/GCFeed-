package infrainteraction

import (
	"context"
	"errors"

	domaininteraction "GCFeed/internal/domain/interaction"
	infravideo "GCFeed/internal/infra/persistence/video"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// SetAction 设置点赞/收藏（事务 + 行锁）。
func (r *Repository) SetAction(ctx context.Context, userID, videoID int64, actionType string, active bool, idempotencyKey string) (bool, int, int, error) {
	var isActive bool
	var likeCount, favoriteCount int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 第一步：查找或创建行为记录
		// 关键：用 struct 作 Where 条件，GORM 才会在 INSERT 时把字段值填入新行
		action := ActionModel{
			UserID:  userID,
			VideoID: videoID,
			Action:  actionType,
		}
		err := tx.Where(&action).
			Attrs(ActionModel{
				Status:         domaininteraction.ActionStatusCanceled,
				IdempotencyKey: idempotencyKey,
			}).
			FirstOrCreate(&action).Error
		if err != nil {
			return err
		}

		// 幂等检查：如果已处理过相同幂等键，直接返回
		if idempotencyKey != "" && action.IdempotencyKey == idempotencyKey &&
			action.Status == boolToStatus(active) {
			isActive = active
			// 读取当前计数
			var stat infravideo.VideoStatModel
			tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("video_id = ?", videoID).First(&stat)
			likeCount = stat.LikeCount
			favoriteCount = stat.FavoriteCount
			return nil
		}

		// 第二步：更新行为状态
		newStatus := boolToStatus(active)
		if err := tx.Model(&action).Updates(map[string]interface{}{
			"status":          newStatus,
			"idempotency_key": idempotencyKey,
		}).Error; err != nil {
			return err
		}
		isActive = active

		// 第三步：更新 video_stat 计数
		delta := 1
		if !active {
			delta = -1
		}

		var statField string
		switch actionType {
		case domaininteraction.ActionLike:
			statField = "like_count"
		case domaininteraction.ActionFavorite:
			statField = "favorite_count"
		}

		if _, err := updateVideoStatCounter(tx, videoID, statField, delta); err != nil {
			return err
		}

		// 回读最终计数
		var stat infravideo.VideoStatModel
		tx.Where("video_id = ?", videoID).First(&stat)
		likeCount = stat.LikeCount
		favoriteCount = stat.FavoriteCount
		return nil
	})

	return isActive, likeCount, favoriteCount, err
}

// SaveComment 保存评论，同时更新评论计数。
func (r *Repository) SaveComment(ctx context.Context, comment *domaininteraction.Comment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := &CommentModel{
			VideoID:        comment.VideoID,
			UserID:         comment.UserID,
			Content:        comment.Content,
			Status:         comment.Status,
			IdempotencyKey: comment.IdempotencyKey,
		}
		if err := tx.Create(model).Error; err != nil {
			return err
		}
		_, err := updateVideoStatCounter(tx, comment.VideoID, "comment_count", 1)
		return err
	})
}

// ListComments 评论列表（游标：按 ID 升序，旧→新）。
func (r *Repository) ListComments(ctx context.Context, videoID int64, cursor int64, limit int) ([]*domaininteraction.Comment, error) {
	var models []*CommentModel
	query := r.db.WithContext(ctx).
		Where("video_id = ? AND status = ?", videoID, domaininteraction.CommentStatusNormal)
	if cursor > 0 {
		query = query.Where("id > ?", cursor) // 只查比游标更新的评论
	}
	if err := query.Order("id ASC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}

	comments := make([]*domaininteraction.Comment, 0, len(models))
	for _, m := range models {
		comments = append(comments, &domaininteraction.Comment{
			ID:        m.ID,
			VideoID:   m.VideoID,
			UserID:    m.UserID,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	return comments, nil
}

// DeleteComment 软删除，同时更新评论计数。
func (r *Repository) DeleteComment(ctx context.Context, id int64, userID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cm CommentModel
		if err := tx.Where("id = ? AND user_id = ? AND status = ?",
			id, userID, domaininteraction.CommentStatusNormal).First(&cm).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domaininteraction.ErrCommentNotFound
			}
			return err
		}

		if err := tx.Model(&cm).Update("status", domaininteraction.CommentStatusDeleted).Error; err != nil {
			return err
		}

		_, err := updateVideoStatCounter(tx, cm.VideoID, "comment_count", -1)
		return err
	})
}

// FindCommentByID 按 ID 查评论。
func (r *Repository) FindCommentByID(ctx context.Context, id int64) (*domaininteraction.Comment, error) {
	var m CommentModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domaininteraction.ErrCommentNotFound
		}
		return nil, err
	}
	return &domaininteraction.Comment{
		ID: m.ID, VideoID: m.VideoID, UserID: m.UserID,
		Content: m.Content, Status: m.Status, CreatedAt: m.CreatedAt,
	}, nil
}

// ========== 辅助函数 ==========

// updateVideoStatCounter 更新 video_stat 表中的计数字段（行锁 + 防负数）。
func updateVideoStatCounter(tx *gorm.DB, videoID int64, field string, delta int) (int, error) {
	var stat infravideo.VideoStatModel
	tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("video_id = ?", videoID).Take(&stat)

	switch field {
	case "like_count":
		stat.LikeCount = clampCount(stat.LikeCount + delta)
	case "favorite_count":
		stat.FavoriteCount = clampCount(stat.FavoriteCount + delta)
	case "comment_count":
		stat.CommentCount = clampCount(stat.CommentCount + delta)
	}

	if err := tx.Save(&stat).Error; err != nil {
		return 0, err
	}
	return statCounter(stat, field), nil
}

// statCounter 根据字段名返回对应的计数值。
func statCounter(stat infravideo.VideoStatModel, field string) int {
	switch field {
	case "like_count":
		return stat.LikeCount
	case "favorite_count":
		return stat.FavoriteCount
	case "comment_count":
		return stat.CommentCount
	default:
		return 0
	}
}

func boolToStatus(active bool) int {
	if active {
		return domaininteraction.ActionStatusActive
	}
	return domaininteraction.ActionStatusCanceled
}

func clampCount(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

var _ domaininteraction.Repository = (*Repository)(nil)
