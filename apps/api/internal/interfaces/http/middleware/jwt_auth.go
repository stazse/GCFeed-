package interfaceshttpmiddleware

import (
	infrajwt "GCFeed/internal/infra/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContextUserIDKey 是在 gin.Context 中存用户 ID 的 key。
const ContextUserIDKey = "auth_user_id"

// NewJWTAuth 创建强制鉴权中间件。
// 如果请求没有合法 token，直接返回 401，不让通过。
func NewJWTAuth(jwtManager *infrajwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头中提取 token
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		// Bearer <token> 格式
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		// 验证 token
		claims, err := jwtManager.ParseAndValidateToken(strings.TrimSpace(parts[1]), infrajwt.TokenTypeAccess)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// 把用户信息写入上下文，后续 Handler 可以读取
		c.Set(ContextUserIDKey, claims.UserID)
		c.Next() // 放行，继续处理请求
	}
}