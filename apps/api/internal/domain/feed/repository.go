package domainfeed

import "context"

// Repository Feed 仓储接口。
type Repository interface {
	// ListTimeline 查询时间线 Feed。
	// cursorID: 上一页最后一条视频的 ID（首次为0）
	// limit: 期望返回多少条
	ListTimeline(ctx context.Context, cursorID int64, limit int) ([]*FeedItem, error)
}
