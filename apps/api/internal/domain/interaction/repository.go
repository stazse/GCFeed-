package domaininteraction

import "context"

// Repository 互动仓储接口。
type Repository interface {
	// SetAction 设置点赞/收藏状态（PUT=喜欢, DELETE=取消）。
	// 返回：是否有效、新的计数、错误
	SetAction(ctx context.Context, userID, videoID int64, actionType string, active bool, idempotencyKey string) (bool, int, int, error)

	// SaveComment 保存评论。
	SaveComment(ctx context.Context, comment *Comment) error

	// FindCommentByID 查评论。
	FindCommentByID(ctx context.Context, id int64) (*Comment, error)

	// ListComments 评论列表（游标分页）。
	ListComments(ctx context.Context, videoID int64, cursor int64, limit int) ([]*Comment, error)

	// DeleteComment 软删除评论。
	DeleteComment(ctx context.Context, id int64, userID int64) error
}