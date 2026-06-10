package infracache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"strconv"

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

// RecordHotScore 记录一次互动热度。
// videoID: 被操作的视频
// delta: 热度变化（点赞=3, 收藏=4, 评论=5, 取消=负数）
func (c *FeedCache) RecordHotScore(ctx context.Context, videoID int64, delta int) error {
	now := time.Now().UTC()
	// 格式化成 yyyyMMddHHmm（分钟粒度）
	bucketKey := fmt.Sprintf("feed:hot:minute:v1:%s", now.Format("200601021504"))
	member := strconv.FormatInt(videoID, 10)

	// ZINCRBY：给成员加分数（delta 可以是负数）
	if err := c.client.ZIncrBy(ctx, bucketKey, float64(delta), member).Err(); err != nil {
		return err
	}

	// 设置过期时间：2 小时（保证窗口足够覆盖 1 小时 + 缓冲）
	c.client.Expire(ctx, bucketKey, 2*time.Hour)

	return nil
}

// GetHotRanking 获取最近 windowMinutes 分钟的热度排行榜。
// 返回 video_id 列表（按热度从高到低），以及用于分页的当前最后一名分数。
func (c *FeedCache) GetHotRanking(ctx context.Context, windowMinutes int, limit int) ([]int64, error) {
	now := time.Now().UTC()
	bucketKeys := make([]string, 0, windowMinutes)

	// 收集最近 N 个分钟桶的 key
	for i := 0; i < windowMinutes; i++ {
		t := now.Add(-time.Duration(i) * time.Minute)
		key := fmt.Sprintf("feed:hot:minute:v1:%s", t.Format("200601021504"))
		bucketKeys = append(bucketKeys, key)
	}

	if len(bucketKeys) == 0 {
		return nil, nil
	}

	// 生成临时合并 key
	windowKey := fmt.Sprintf("feed:hot:window:v1:%d", now.Unix())

	// ZUNIONSTORE：合并所有桶（AGGREGATE SUM 表示同一 video_id 分数累加）
	c.client.ZUnionStore(ctx, windowKey, &redis.ZStore{
		Keys:      bucketKeys,
		Aggregate: "SUM",
	})

	// 设置合并结果的过期时间（1 分钟，因为下一分钟窗口就变了）
	c.client.Expire(ctx, windowKey, 1*time.Minute)

	// 取热度前 N 名（按分数从高到低）
	members, err := c.client.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   windowKey,
		Start: 0,
		Stop:  int64(limit - 1),
		Rev:   true, // 倒序 = 高分在前
	}).Result()
	if err != nil {
		return nil, err
	}

	videoIDs := make([]int64, 0, len(members))
	for _, m := range members {
		id, _ := strconv.ParseInt(m, 10, 64)
		if id > 0 {
			videoIDs = append(videoIDs, id)
		}
	}

	return videoIDs, nil
}