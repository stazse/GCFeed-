package domainfeed

import (
	"strings"
)

// Feed 场景常量
const (
	SceneTimeline = "timeline" // 时间线（全部视频按时间）
)

// NormalizeScene 清洗场景参数（去空格、给默认值）
func NormalizeScene(scene string) string {
	scene = strings.TrimSpace(scene)
	if scene == "" {
		return SceneTimeline
	}
	return scene
}

// FeedItem 是 Feed 流中的一条视频摘要。
// 它不是完整的 Video 对象，只包含 Feed 列表需要展示的字段。
type FeedItem struct {
	VideoID      int64  `json:"video_id"`
	AuthorID     int64  `json:"author_id"`
	Title        string `json:"title"`
	MediaURL     string `json:"media_url"`
	CoverURL     string `json:"cover_url"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	PublishedAt  string `json:"published_at"`
}

// FeedPage 是一页 Feed 结果。
type FeedPage struct {
	Items      []*FeedItem // 这一页的视频列表
	NextCursor string      // 下一页的游标（空表示没有下一页）
	HasMore    bool        // 是否还有更多
}
