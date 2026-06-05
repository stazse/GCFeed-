package interfaceshttpaccount

import (
	"errors"
	"net/http"

	applicationaccount "GCFeed/internal/application/account"
	domainaccount "GCFeed/internal/domain/account"

	"github.com/gin-gonic/gin"
)

// Handler 处理账户相关的 HTTP 请求。
type Handler struct {
	service *applicationaccount.Service
}

func New(service *applicationaccount.Service) *Handler {
	return &Handler{service: service}
}

// Register 注册接口：POST /api/users
func (h *Handler) Register(c *gin.Context) {
	// 第一步：从请求体解析 JSON
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 第二步：调用业务服务
	result, err := h.service.Register(c.Request.Context(), req.Account, req.Password, req.Nickname)
	if err != nil {
		writeAccountError(c, err) // 根据错误类型返回不同状态码
		return
	}

	// 第三步：返回成功响应
	c.JSON(http.StatusCreated, authResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		UserID:      result.UserID,
		Nickname:    result.Nickname,
	})
}

// Login 登录接口：POST /api/sessions
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.service.Login(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		writeAccountError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		UserID:      result.UserID,
		Nickname:    result.Nickname,
	})
}

// writeAccountError 把错误映射为 HTTP 状态码。
// 这种写法在项目中很常见：一个模块一个错误映射函数。
func writeAccountError(c *gin.Context, err error) {
	// 参数错误 → 400
	if errors.Is(err, domainaccount.ErrEmptyAccount) ||
		errors.Is(err, domainaccount.ErrEmptyPassword) ||
		errors.Is(err, domainaccount.ErrEmptyNickname) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 账号密码错误 → 401
	if errors.Is(err, domainaccount.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid account or password"})
		return
	}
	// 账号已存在 → 409
	if errors.Is(err, domainaccount.ErrAccountExists) {
		c.JSON(http.StatusConflict, gin.H{"error": "account already exists"})
		return
	}
	// 其余错误 → 500
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
