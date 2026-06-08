package interfaceshttprelation

import (
	"net/http"
	"strconv"

	applicationrelation "GCFeed/internal/application/relation"
	interfaceshttpmiddleware "GCFeed/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *applicationrelation.Service
}

func New(service *applicationrelation.Service) *Handler {
	return &Handler{service: service}
}

// Follow 关注：PUT /api/users/me/following/:targetUserId
func (h *Handler) Follow(c *gin.Context) {
	userID := c.GetInt64(interfaceshttpmiddleware.ContextUserIDKey)
	targetID, _ := strconv.ParseInt(c.Param("targetUserId"), 10, 64)

	if err := h.service.Follow(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "followed"})
}

// Unfollow 取关：DELETE /api/users/me/following/:targetUserId
func (h *Handler) Unfollow(c *gin.Context) {
	userID := c.GetInt64(interfaceshttpmiddleware.ContextUserIDKey)
	targetID, _ := strconv.ParseInt(c.Param("targetUserId"), 10, 64)

	if err := h.service.Unfollow(c.Request.Context(), userID, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unfollowed"})
}

// ListFollowing 关注列表：GET /api/users/me/following
func (h *Handler) ListFollowing(c *gin.Context) {
	userID := c.GetInt64(interfaceshttpmiddleware.ContextUserIDKey)
	cursor, _ := strconv.ParseInt(c.Query("cursor"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	follows, err := h.service.ListFollowing(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": follows, "next_cursor": ""})
}

// ListFollowers 粉丝列表：GET /api/users/me/followers
func (h *Handler) ListFollowers(c *gin.Context) {
	userID := c.GetInt64(interfaceshttpmiddleware.ContextUserIDKey)
	cursor, _ := strconv.ParseInt(c.Query("cursor"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	follows, err := h.service.ListFollowers(c.Request.Context(), userID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": follows, "next_cursor": ""})
}