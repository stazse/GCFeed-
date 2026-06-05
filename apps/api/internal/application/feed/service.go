package applicationfeed

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	domainfeed "GCFeed/internal/domain/feed"
)

const (
	defaultFeedLimit = 10 // 默认每页 10 条
	maxFeedLimit     = 50 // 最多 50 条
)

var (
	ErrLoadFeedFailed = errors.New("failed to load feed")
)

// cursorData 游标的内部结构（JSON → base64）
type cursorData struct {
	LastID int64 `json:"last_id"`
}

// Service Feed 应用服务。
type Service struct {
	repo domainfeed.Repository
}

func New(repo domainfeed.Repository) *Service {
	return &Service{repo: repo}
}

// GetFeed 获取 Feed 流。
func (s *Service) GetFeed(ctx context.Context, scene string, rawCursor string, limit int) (*domainfeed.FeedPage, error) {
	// 标准化 scene
	scene = domainfeed.NormalizeScene(scene)

	// 裁剪 limit
	if limit <= 0 {
		limit = defaultFeedLimit
	}
	if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	// 解析游标
	cursor := decodeCursor(rawCursor)

	// 多取一条 → 用来判断 has_more
	// 比如 limit=10，我们查 11 条：
	//   - 返回 11 条 → 说明有下一页
	//   - 返回 ≤10 条 → 说明到末尾了
	items, err := s.repo.ListTimeline(ctx, cursor.LastID, limit+1)
	if err != nil {
		return nil, errors.Join(ErrLoadFeedFailed, err)
	}

	// 判断是否有更多
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit] // 只返回 limit 条
	}

	// 生成下一页游标
	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = encodeCursor(items[len(items)-1].VideoID)
	}

	return &domainfeed.FeedPage{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// decodeCursor 把 base64 字符串解码为游标数据。
func decodeCursor(raw string) cursorData {
	if raw == "" {
		return cursorData{}
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return cursorData{}
	}
	var c cursorData
	json.Unmarshal(decoded, &c)
	return c
}

// encodeCursor 把最后一条视频 ID 编码为 base64 字符串。
func encodeCursor(lastID int64) string {
	data, _ := json.Marshal(cursorData{LastID: lastID})
	return base64.RawURLEncoding.EncodeToString(data)
}
