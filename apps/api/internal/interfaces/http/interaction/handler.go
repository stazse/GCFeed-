package interfaceshttpinteraction

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	applicationinteraction "GCFeed/internal/application/interaction"
	domaininteraction "GCFeed/internal/domain/interaction"
	interfaceshttpmiddleware "GCFeed/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *applicationinteraction.Service
}

func New(service *applicationinteraction.Service) *Handler {
	return &Handler{service: service}
}

// Like 点赞：PUT /api/videos/:videoId/like
func (h *Handler) Like(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	key := c.GetHeader("Idempotency-Key")

	active, count, err := h.service.SetLike(c.Request.Context(), userID.(int64), videoID, true, key)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_liked": active, "like_count": count})
}

// Unlike 取消点赞：DELETE /api/videos/:videoId/like
func (h *Handler) Unlike(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	key := c.GetHeader("Idempotency-Key")

	active, count, err := h.service.SetLike(c.Request.Context(), userID.(int64), videoID, false, key)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_liked": active, "like_count": count})
}

// Favorite 收藏：PUT /api/videos/:videoId/favorite
func (h *Handler) Favorite(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	key := c.GetHeader("Idempotency-Key")

	active, count, err := h.service.SetFavorite(c.Request.Context(), userID.(int64), videoID, true, key)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_favorited": active, "favorite_count": count})
}

// Unfavorite 取消收藏：DELETE /api/videos/:videoId/favorite
func (h *Handler) Unfavorite(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	key := c.GetHeader("Idempotency-Key")

	active, count, err := h.service.SetFavorite(c.Request.Context(), userID.(int64), videoID, false, key)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_favorited": active, "favorite_count": count})
}

// CreateComment 发评论：POST /api/videos/:videoId/comments
func (h *Handler) CreateComment(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)
	key := c.GetHeader("Idempotency-Key")

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	comment, err := h.service.CreateComment(c.Request.Context(), videoID, userID.(int64), req.Content, key)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": comment.ID, "content": comment.Content})
}

// ListComments 评论列表：GET /api/videos/:videoId/comments
func (h *Handler) ListComments(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	cursor := parseCommentCursor(c.Query("cursor"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	comments, err := h.service.ListComments(c.Request.Context(), videoID, cursor, limit)
	if err != nil {
		writeInteractionError(c, err)
		return
	}

	type resp struct {
		ID        int64  `json:"id"`
		UserID    int64  `json:"user_id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}
	items := make([]*resp, 0, len(comments))
	for _, c := range comments {
		items = append(items, &resp{
			ID: c.ID, UserID: c.UserID, Content: c.Content,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// DeleteComment 删除评论：DELETE /api/comments/:commentId
func (h *Handler) DeleteComment(c *gin.Context) {
	commentID, _ := strconv.ParseInt(c.Param("commentId"), 10, 64)
	userID, _ := c.Get(interfaceshttpmiddleware.ContextUserIDKey)

	if err := h.service.DeleteComment(c.Request.Context(), commentID, userID.(int64)); err != nil {
		writeInteractionError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ========== 辅助函数 ==========

func parseCommentCursor(raw string) int64 {
	if raw == "" {
		return 0
	}
	decoded, _ := base64.RawURLEncoding.DecodeString(raw)
	var c struct{ LastID int64 }
	json.Unmarshal(decoded, &c)
	return c.LastID
}

func writeInteractionError(c *gin.Context, err error) {
	if errors.Is(err, domaininteraction.ErrCommentNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if errors.Is(err, domaininteraction.ErrNotCommentOwner) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not comment owner"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}