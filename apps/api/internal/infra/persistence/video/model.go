package infravideo

import "time"

// VideoModel video 表的 GORM 模型。
type VideoModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	AuthorID       int64     `gorm:"column:author_id;not null;index"`
	Title          string    `gorm:"column:title;type:varchar(256);not null"`
	Description    string    `gorm:"column:description;type:text"`
	MediaURL       string    `gorm:"column:media_url;type:varchar(512)"`
	CoverURL       string    `gorm:"column:cover_url;type:varchar(512)"`
	Status         int       `gorm:"column:status;not null;default:1"`
	PublishedAt    time.Time `gorm:"column:published_at"`
	IdempotencyKey string    `gorm:"column:idempotency_key;type:varchar(128);uniqueIndex"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (VideoModel) TableName() string {
	return "video"
}

// VideoStatModel video_stat 表的 GORM 模型。
type VideoStatModel struct {
	VideoID       int64 `gorm:"column:video_id;primaryKey"`
	LikeCount     int   `gorm:"column:like_count;default:0"`
	CommentCount  int   `gorm:"column:comment_count;default:0"`
	FavoriteCount int   `gorm:"column:favorite_count;default:0"`
}

func (VideoStatModel) TableName() string {
	return "video_stat"
}