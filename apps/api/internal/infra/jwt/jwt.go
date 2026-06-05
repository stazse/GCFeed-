package infrajwt

import (
	"fmt"
	"time"

	infraconfig "GCFeed/internal/infra/config"

	"github.com/golang-jwt/jwt/v5"
)

// TokenTypeAccess 表示这是"访问令牌"。
const TokenTypeAccess = "access"

// Claims 是 JWT 里面存的自定义数据。
// 你可以把它理解为"手环上写的信息"。
type Claims struct {
	UserID               int64  `json:"user_id"`
	Role                 string `json:"role"`
	TokenType            string `json:"token_type"`
	jwt.RegisteredClaims        // JWT 标准字段（过期时间等）
}

// Manager 负责签发和验证 JWT。
type Manager struct {
	secret    []byte        // 签名密钥
	accessTTL time.Duration // token 有效期
}

// NewManager 创建 JWT 管理器。
func NewManager(cfg *infraconfig.JWTConfig) (*Manager, error) {
	if cfg.Secret == "" {
		return nil, fmt.Errorf("JWT secret is empty")
	}
	ttl, err := time.ParseDuration(cfg.AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("JWT access TTL is invalid: %w", err)
	}
	return &Manager{
		secret:    []byte(cfg.Secret),
		accessTTL: ttl,
	}, nil
}

// IssueAccessToken 签发访问令牌（登录成功后调用）。
func (m *Manager) IssueAccessToken(userID int64, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)), //过期时间
			IssuedAt:  jwt.NewNumericDate(now),                  //签发时间
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 用密钥签名，生成最终的 token 字符串
	return token.SignedString(m.secret)
}

// ParseAndValidateToken 验证并解析 token（每次请求时调用）。
func (m *Manager) ParseAndValidateToken(tokenString string, tokenType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	if claims.TokenType != tokenType {
		return nil, fmt.Errorf("invalid token type ")
	}
	return claims, nil
}
