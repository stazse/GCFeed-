package applicationfeed

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	domainfeed "GCFeed/internal/domain/feed"
	infracache "GCFeed/internal/infra/cache"
)

const (
	defaultFeedLimit = 10 // 默认每页 10 条
	maxFeedLimit     = 50 // 最多 50 条
)

var (
	ErrLoadFeedFailed = errors.New("failed to load feed")
)

// FeedCache 是缓存能力的最小接口。
type FeedCache interface {
	GetTimelinePage(ctx context.Context, cursorID int64, limit int) (*infracache.CachedFeedPage, error)
	SetTimelinePage(ctx context.Context, cursorID int64, limit int, page *infracache.CachedFeedPage) error
	GetVideoCards(ctx context.Context, videoIDs []int64) (map[int64]*infracache.VideoCard, []int64, error)
	SetVideoCard(ctx context.Context, card *infracache.VideoCard) error
}

type Service struct {
	repo      domainfeed.Repository
	feedCache FeedCache // 新增：缓存（可以为 nil，nil 时走纯数据库模式）
}

// WithFeedCache 注入缓存（函数选项模式）。
// 这样设计与 nil 检查配合：没 Redis 也能跑，有 Redis 就更快。
func WithFeedCache(cache FeedCache) func(*Service) {
	return func(s *Service) {
		s.feedCache = cache
	}
}

func New(repo domainfeed.Repository, opts ...func(*Service)) *Service {
	s := &Service{repo: repo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// cursorData 游标的内部结构（JSON → base64）
type cursorData struct {
	LastID int64 `json:"last_id"`
}

// GetFeed 获取 Feed 流。
// GetFeed 带缓存的 Feed 查询。
func (s *Service) GetFeed(ctx context.Context, scene string, rawCursor string, limit int) (*domainfeed.FeedPage, error) {
	scene = domainfeed.NormalizeScene(scene)
	if limit <= 0 {
		limit = defaultFeedLimit
	}
	if limit > maxFeedLimit {
		limit = maxFeedLimit
	}

	cursor := decodeCursor(rawCursor)

	// ---- 尝试从缓存读取 ----
	if s.feedCache != nil {
		cached, err := s.feedCache.GetTimelinePage(ctx, cursor.LastID, limit)
		if err == nil && cached != nil {
			// 缓存命中！但还需要批量获取视频卡片（因为页缓存只存了 video_id）
			cards, missed, _ := s.feedCache.GetVideoCards(ctx, cached.VideoIDs)
			if len(missed) == 0 {
				// 全部命中，直接组装返回
				return s.buildPageFromCards(cached.VideoIDs, cards, cached.HasMore), nil
			}
			// 有部分缺失，需要回源补充
		}
	}

	// ---- 缓存未命中，回源数据库 ----
	items, err := s.repo.ListTimeline(ctx, cursor.LastID, limit+1)
	if err != nil {
		return nil, errors.Join(ErrLoadFeedFailed, err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	// 把回源结果写回缓存
	if s.feedCache != nil && len(items) > 0 {
		videoIDs := make([]int64, len(items))
		for i, item := range items {
			videoIDs[i] = item.VideoID
			// 同时把每条视频卡片也缓存起来
			s.feedCache.SetVideoCard(ctx, &infracache.VideoCard{
				VideoID:      item.VideoID,
				AuthorID:     item.AuthorID,
				Title:        item.Title,
				MediaURL:     item.MediaURL,
				CoverURL:     item.CoverURL,
				LikeCount:    item.LikeCount,
				CommentCount: item.CommentCount,
				PublishedAt:  item.PublishedAt,
			})
		}
		s.feedCache.SetTimelinePage(ctx, cursor.LastID, limit, &infracache.CachedFeedPage{
			VideoIDs: videoIDs,
			HasMore:  hasMore,
		})
	}

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

// buildPageFromCards 从缓存的卡片数据组装 Feed 页。
func (s *Service) buildPageFromCards(videoIDs []int64, cards map[int64]*infracache.VideoCard, hasMore bool) *domainfeed.FeedPage {
	items := make([]*domainfeed.FeedItem, 0, len(videoIDs))
	for _, vid := range videoIDs {
		card, ok := cards[vid]
		if !ok {
			continue
		}
		items = append(items, &domainfeed.FeedItem{
			VideoID:      card.VideoID,
			AuthorID:     card.AuthorID,
			Title:        card.Title,
			MediaURL:     card.MediaURL,
			CoverURL:     card.CoverURL,
			LikeCount:    card.LikeCount,
			CommentCount: card.CommentCount,
			PublishedAt:  card.PublishedAt,
		})
	}

	// 生成游标
	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = encodeCursor(items[len(items)-1].VideoID)
	}

	return &domainfeed.FeedPage{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
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
