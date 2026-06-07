package domaininteraction

import (
	"strings"
	"time"
)

// 行为类型常量
const (
	ActionLike     = "like"
	ActionFavorite = "favorite"
)

// 行为状态
const (
	ActionStatusActive   = 1 // 有效
	ActionStatusCanceled = 0 // 已取消
)

// 评论状态
const (
	CommentStatusNormal  = 1
	CommentStatusDeleted = 2
)

// 幂等键最大长度
const MaxIdempotencyKeyLength = 128

// Comment 评论实体。
type Comment struct {
	ID             int64
	VideoID        int64
	UserID         int64
	Content        string
	Status         int
	IdempotencyKey string
	CreatedAt      time.Time
}

// NewComment 创建评论。
func NewComment(videoID, userID int64, content, idempotencyKey string) (*Comment, error) {
	if videoID <= 0 {
		return nil, ErrInvalidVideoID
	}
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyContent
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > MaxIdempotencyKeyLength {
		return nil, ErrIdempotencyKeyTooLong
	}

	return &Comment{
		VideoID:        videoID,
		UserID:         userID,
		Content:        content,
		Status:         CommentStatusNormal,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now().UTC(),
	}, nil
}