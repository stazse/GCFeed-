package interfaceshttpvideo

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	applicationvideo "GCFeed/internal/application/video"
	domainvideo "GCFeed/internal/domain/video"
	interfaceshttpmiddleware "GCFeed/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *applicationvideo.Service
}

func New(service *applicationvideo.Service) *Handler {
	return &Handler{service: service}
}

// Create 发布视频：POST /api/videos
func (h *Handler) Create(c *gin.Context) {
	// 从 JWT 中读取当前用户 ID
	userID, ok := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req createVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 读取幂等键（请求头）
	idempotencyKey := c.GetHeader("Idempotency-Key")

	video, err := h.service.Create(
		c.Request.Context(),
		userID.(int64),
		req.Title, req.MediaURL, req.CoverURL,
		idempotencyKey,
	)
	if err != nil {
		writeVideoError(c, err)
		return
	}

	c.JSON(http.StatusCreated, videoToResponse(video))
}

// Get 视频详情：GET /api/videos/:videoId
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("videoId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video id"})
		return
	}

	video, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		writeVideoError(c, err)
		return
	}

	c.JSON(http.StatusOK, videoToResponse(video))
}

// ListMine 我的作品列表：GET /api/users/me/videos
func (h *Handler) ListMine(c *gin.Context) {
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	cursor, _ := strconv.ParseInt(c.Query("cursor"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	videos, err := h.service.ListByAuthor(c.Request.Context(), userID.(int64), cursor, limit)
	if err != nil {
		writeVideoError(c, err)
		return
	}

	items := make([]*videoResponse, 0, len(videos))
	var lastID int64
	for _, v := range videos {
		items = append(items, videoToResponse(v))
		lastID = v.ID
	}

	c.JSON(http.StatusOK, videoListResponse{Videos: items, Cursor: lastID})
}

// videoToResponse 领域对象 → HTTP 响应。
func videoToResponse(v *domainvideo.Video) *videoResponse {
	return &videoResponse{
		ID:            v.ID,
		AuthorID:      v.AuthorID,
		Title:         v.Title,
		MediaURL:      v.MediaURL,
		CoverURL:      v.CoverURL,
		Status:        v.Status,
		LikeCount:     v.LikeCount,
		CommentCount:  v.CommentCount,
		FavoriteCount: v.FavoriteCount,
		PublishedAt:   v.PublishedAt.Format(time.RFC3339),
	}
}

func writeVideoError(c *gin.Context, err error) {
	if errors.Is(err, domainvideo.ErrVideoNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}
	if errors.Is(err, domainvideo.ErrInvalidAuthorID) ||
		errors.Is(err, domainvideo.ErrEmptyTitle) ||
		errors.Is(err, domainvideo.ErrTitleTooLong) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
