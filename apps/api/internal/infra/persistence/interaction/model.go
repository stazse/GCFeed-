package infrainteraction

import "time"

// ActionModel interaction_action 表。
type ActionModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID         int64     `gorm:"column:user_id;not null;uniqueIndex:idx_user_video_action"`
	VideoID        int64     `gorm:"column:video_id;not null;uniqueIndex:idx_user_video_action"`
	Action         string    `gorm:"column:action;type:varchar(16);not null;uniqueIndex:idx_user_video_action"`
	Status         int       `gorm:"column:status;not null;default:1"`
	IdempotencyKey string    `gorm:"column:idempotency_key;type:varchar(128)"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ActionModel) TableName() string {
	return "interaction_action"
}

// CommentModel interaction_comment 表。
type CommentModel struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	VideoID        int64     `gorm:"column:video_id;not null;index"`
	UserID         int64     `gorm:"column:user_id;not null"`
	Content        string    `gorm:"column:content;type:text;not null"`
	Status         int       `gorm:"column:status;not null;default:1"`
	IdempotencyKey string    `gorm:"column:idempotency_key;type:varchar(128);uniqueIndex"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (CommentModel) TableName() string {
	return "interaction_comment"
}