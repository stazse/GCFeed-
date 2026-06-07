package infracache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// FeedCache 负责 Feed 相关的缓存操作。
type FeedCache struct {
	client *redis.Client
}

// NewFeedCache 创建 Feed 缓存。
func NewFeedCache(client *redis.Client) *FeedCache {
	return &FeedCache{client: client}
}

// ========== Feed 页缓存 ==========

// CachedFeedPage Redis 中存储的 Feed 页结构。
// 只存 video_id 列表（轻量），卡片信息另外批量查。
type CachedFeedPage struct {
	VideoIDs []int64 `json:"video_ids"`
	HasMore  bool    `json:"has_more"`
}

// GetTimelinePage 读取时间线首页缓存。
// key 格式：feed:timeline:page:0:{limit}
func (c *FeedCache) GetTimelinePage(ctx context.Context, cursorID int64, limit int) (*CachedFeedPage, error) {
	key := fmt.Sprintf("feed:timeline:page:%d:%d", cursorID, limit)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err // key 不存在时 redis 返回 Nil
	}
	var page CachedFeedPage
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// SetTimelinePage 写入时间线页缓存，TTL = 3 秒。
// 为什么是 3 秒？首页是高频访问点，3 秒意味着最多 3 秒后就能看到新发布的视频。
// 对用户体验来说，3 秒的"不实时"在 Feed 流中完全感觉不到。
func (c *FeedCache) SetTimelinePage(ctx context.Context, cursorID int64, limit int, page *CachedFeedPage) error {
	key := fmt.Sprintf("feed:timeline:page:%d:%d", cursorID, limit)
	data, err := json.Marshal(page)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 3*time.Second).Err()
}

// ========== 视频卡片缓存 ==========

// VideoCard 视频卡片（Feed 中展示一条视频需要的所有信息）。
type VideoCard struct {
	VideoID      int64  `json:"video_id"`
	AuthorID     int64  `json:"author_id"`
	Title        string `json:"title"`
	MediaURL     string `json:"media_url"`
	CoverURL     string `json:"cover_url"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	PublishedAt  string `json:"published_at"`
}

// GetVideoCards 批量获取视频卡片（MGET，一次网络往返）。
func (c *FeedCache) GetVideoCards(ctx context.Context, videoIDs []int64) (map[int64]*VideoCard, []int64, error) {
	if len(videoIDs) == 0 {
		return nil, nil, nil
	}

	// 构建 MGET 的 key 列表
	keys := make([]string, len(videoIDs))
	for i, id := range videoIDs {
		keys[i] = fmt.Sprintf("video:card:v1:%d", id)
	}

	// MGET：一次请求获取多个 key
	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}

	cards := make(map[int64]*VideoCard)
	var missed []int64

	for i, result := range results {
		if result == nil {
			missed = append(missed, videoIDs[i])
			continue
		}
		var card VideoCard
		if err := json.Unmarshal([]byte(result.(string)), &card); err != nil {
			missed = append(missed, videoIDs[i])
			continue
		}
		cards[videoIDs[i]] = &card
	}

	return cards, missed, nil
}

// SetVideoCard 写入单条视频卡片缓存，TTL = 15 分钟。
// 15 分钟意味着一个视频发布后，最多 15 分钟就会自动被下一次回源刷新。
// 视频标题、封面等不经常变，15 分钟完全可接受。
func (c *FeedCache) SetVideoCard(ctx context.Context, card *VideoCard) error {
	key := fmt.Sprintf("video:card:v1:%d", card.VideoID)
	data, err := json.Marshal(card)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 15*time.Minute).Err()
}