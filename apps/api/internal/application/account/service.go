package applicationaccount

import (
	"context"
	"errors"

	domainaccount "GCFeed/internal/domain/account"
)

// 应用层自己的错误（表达用例失败，而不是业务规则违反）
var (
	ErrRegisterFailed = errors.New("failed to register")
	ErrLoginFailed    = errors.New("failed to login")
)

// TokenSigner 签发 token 的最小接口。
type TokenSigner interface {
	IssueAccessToken(userID int64, role string) (string, error)
}

// RegisterResult 注册成功后的返回内容。
type RegisterResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	UserID      int64  `json:"user_id"`
	Nickname    string `json:"nickname"`
}

// LoginResult 登录成功后的返回内容。
type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	UserID      int64  `json:"user_id"`
	Nickname    string `json:"nickname"`
}

// Service 账户模块的应用服务。
// 它只依赖领域层的 Repository 接口，不依赖 GORM。
type Service struct {
	repo        domainaccount.Repository
	tokenSigner TokenSigner // 签发 JWT 的能力（由 JWT Manager 提供）
}

// New 创建账户服务。
func New(repo domainaccount.Repository, tokenSigner TokenSigner) *Service {
	return &Service{repo: repo, tokenSigner: tokenSigner}
}

//Register 注册用户。
func (s *Service) Register(ctx context.Context, account, password, nickname string) (*RegisterResult, error) {
	//第一步：用领域层规则创建用户（检查输入，哈希密码）
	user, err := domainaccount.New(account, password, nickname)
	if err != nil {
		return nil, err//参不合法，返回错误
	}

	//第二步：保存用户到数据库
	if err := s.repo.Save(ctx, user); err != nil {
		if errors.Is(err, domainaccount.ErrAccountExists) {
			return nil, err // 账号已存在
		}
		return nil, errors.Join(ErrRegisterFailed, err)
	}

	//第三步：签发访问令牌
	token, err := s.tokenSigner.IssueAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.Join(ErrRegisterFailed, err)
	}

	//第四步：返回注册结果
	return &RegisterResult{
		AccessToken: token,
		TokenType:   "Bearer",
		UserID:      user.ID,
		Nickname:    user.Nickname,
	}, nil
}

// Login 登录用户。
func (s *Service) Login(ctx context.Context, account, password string) (*LoginResult, error) {
	//第一步：用领域层规则检查用户是否存在（检查输入，哈希密码）
	user, err := s.repo.FindByAccount(ctx, account)
	if err != nil {
		if errors.Is(err, domainaccount.ErrUserNotFound) {
			return nil, domainaccount.ErrInvalidCredentials // 不暴露"用户不存在"
		}
		return nil, errors.Join(ErrLoginFailed, err)
	}

	//第二步：检查密码是否匹配
	if err := user.Authenticate(password); err != nil {
		return nil, err // 密码错误
	}

	//第三步：签发访问令牌
	token, err := s.tokenSigner.IssueAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.Join(ErrLoginFailed, err)
	}

	//第四步：返回登录结果
	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		UserID:      user.ID,
		Nickname:    user.Nickname,
	}, nil
}