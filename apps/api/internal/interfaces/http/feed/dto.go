package interfaceshttpfeed

// feedItemResponse 单条 Feed 的响应格式。
type feedItemResponse struct {
	VideoID      int64  `json:"video_id"`
	AuthorID     int64  `json:"author_id"`
	Title        string `json:"title"`
	MediaURL     string `json:"media_url"`
	CoverURL     string `json:"cover_url"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	PublishedAt  string `json:"published_at"`
}

// feedListResponse Feed 列表响应格式。
// 业界常用的三个字段：items, next_cursor, has_more。
type feedListResponse struct {
	Items      []*feedItemResponse `json:"items"`
	NextCursor string              `json:"next_cursor"`
	HasMore    bool                `json:"has_more"`
}
