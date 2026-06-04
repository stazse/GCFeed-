package domainvideo

import "context"

// Repository 视频仓储接口。
type Repository interface {
	// Save 保存视频（同时创建 video 和 video_stat 记录，放在一个事务里）
	Save(ctx context.Context, video *Video) error

	// FindByID 按 ID 查视频（含计数）
	FindByID(ctx context.Context, id int64) (*Video, error)

	// FindByIdempotencyKey 按幂等键查找（用于防止重复提交）
	FindByIdempotencyKey(ctx context.Context, key string) (*Video, error)

	// ListByAuthor 按作者分页查视频
	ListByAuthor(ctx context.Context, authorID int64, cursor int64, limit int) ([]*Video, error)
}