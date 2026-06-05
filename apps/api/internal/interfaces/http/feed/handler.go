package interfaceshttpfeed

import (
	"net/http"
	"strconv"

	applicationfeed "GCFeed/internal/application/feed"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *applicationfeed.Service
}

func New(service *applicationfeed.Service) *Handler {
	return &Handler{service: service}
}

// ListFeedItems Feed 流接口：GET /api/feed-items?scene=timeline&cursor=&limit=10
func (h *Handler) ListFeedItems(c *gin.Context) {
	scene := c.DefaultQuery("scene", "timeline")
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	page, err := h.service.GetFeed(c.Request.Context(), scene, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load feed"})
		return
	}

	// 转换领域对象 → HTTP 响应
	items := make([]*feedItemResponse, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, &feedItemResponse{
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

	c.JSON(http.StatusOK, feedListResponse{
		Items:      items,
		NextCursor: page.NextCursor,
		HasMore:    page.HasMore,
	})
}
