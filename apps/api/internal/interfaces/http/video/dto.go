package interfaceshttpvideo

// createVideoRequest 发布视频请求体。
type createVideoRequest struct {
	Title    string `json:"title"`
	MediaURL string `json:"media_url"`
	CoverURL string `json:"cover_url"`
}

// videoResponse 视频响应体（对外返回的格式）。
type videoResponse struct {
	ID            int64  `json:"id"`
	AuthorID      int64  `json:"author_id"`
	Title         string `json:"title"`
	MediaURL      string `json:"media_url"`
	CoverURL      string `json:"cover_url"`
	Status        int    `json:"status"`
	LikeCount     int    `json:"like_count"`
	CommentCount  int    `json:"comment_count"`
	FavoriteCount int    `json:"favorite_count"`
	PublishedAt   string `json:"published_at"`
}

// videoListResponse 视频列表响应。
type videoListResponse struct {
	Videos []*videoResponse `json:"videos"`
	Cursor int64            `json:"cursor"`
}
