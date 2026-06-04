package interfaceshttpaccount

// registerRequest 注册请求体。
// json 标签告诉 Gin："请求里的 JSON 字段名叫 account，对应 Go 的 Account 字段"。
type registerRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// loginRequest 登录请求体。
type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// authResponse 认证成功后的响应（注册和登录成功后返回的一样）。
type authResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	UserID      int64  `json:"user_id"`
	Nickname    string `json:"nickname"`
}