package domainvideo

import (
	"strings"
	"time"
)

// 视频状态常量
const (
	StatusPublished = 2 // 已发布
	StatusDraft     = 1 // 草稿
	StatusDeleted   = 3 // 已删除
)

// 输入限制
const (
	MaxTitleLength       = 128
	MaxIdempotencyKeyLen = 128
)

// Video 是视频聚合根。
type Video struct {
	ID             int64
	AuthorID       int64 // 谁发的
	Title          string
	Description    string
	MediaURL       string // 视频文件路径
	CoverURL       string // 封面图路径
	Status         int
	PublishedAt    time.Time
	IdempotencyKey string // 幂等键
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// 下面这些字段来自 video_stat 表（JOIN 查询时填充）
	LikeCount     int
	CommentCount  int
	FavoriteCount int
}

// NewPublished 创建一个"待发布"的视频实体。
// 所有输入在这里校验，不符合规则的不让创建。
func NewPublished(authorID int64, title, mediaURL, coverURL, idempotencyKey string) (*Video, error) {
	if authorID <= 0 {
		return nil, ErrInvalidAuthorID
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if len([]rune(title)) > MaxTitleLength {
		return nil, ErrTitleTooLong
	}

	mediaURL = strings.TrimSpace(mediaURL)
	coverURL = strings.TrimSpace(coverURL)

	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > MaxIdempotencyKeyLen {
		return nil, ErrIdempotencyKeyTooLong
	}

	now := time.Now().UTC()
	return &Video{
		AuthorID:       authorID,
		Title:          title,
		MediaURL:       mediaURL,
		CoverURL:       coverURL,
		Status:         StatusPublished,
		PublishedAt:    now,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
	}, nil
}
