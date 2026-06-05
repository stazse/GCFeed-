package infrafeed

import (
	"context"

	domainfeed "GCFeed/internal/domain/feed"
	domainvideo "GCFeed/internal/domain/video"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListTimeline 按发布时间倒序查询 Feed。
// 游标分页：用 id < cursorID 保证翻页稳定。
func (r *Repository) ListTimeline(ctx context.Context, cursorID int64, limit int) ([]*domainfeed.FeedItem, error) {
	query := r.db.WithContext(ctx).
		Table("video").
		Select("video.id, video.author_id, video.title, video.media_url, video.cover_url, "+
			"COALESCE(video_stat.like_count, 0) AS like_count, "+
			"COALESCE(video_stat.comment_count, 0) AS comment_count, "+
			"video.published_at").
		Joins("LEFT JOIN video_stat ON video.id = video_stat.video_id").
		Where("video.status = ?", domainvideo.StatusPublished)

	// 游标条件：只查比 cursorID 更早的视频
	if cursorID > 0 {
		query = query.Where("video.id < ?", cursorID)
	}

	// 按 ID 倒序（因为 ID 自增，越新的 ID 越大）
	query = query.Order("video.id DESC").Limit(limit)

	// 用一个临时的查询结果结构来接收数据
	type row struct {
		ID           int64  `gorm:"column:id"`
		AuthorID     int64  `gorm:"column:author_id"`
		Title        string `gorm:"column:title"`
		MediaURL     string `gorm:"column:media_url"`
		CoverURL     string `gorm:"column:cover_url"`
		LikeCount    int    `gorm:"column:like_count"`
		CommentCount int    `gorm:"column:comment_count"`
		PublishedAt  string `gorm:"column:published_at"`
	}

	var rows []row
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]*domainfeed.FeedItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, &domainfeed.FeedItem{
			VideoID:      r.ID,
			AuthorID:     r.AuthorID,
			Title:        r.Title,
			MediaURL:     r.MediaURL,
			CoverURL:     r.CoverURL,
			LikeCount:    r.LikeCount,
			CommentCount: r.CommentCount,
			PublishedAt:  r.PublishedAt,
		})
	}
	return items, nil
}

var _ domainfeed.Repository = (*Repository)(nil)
