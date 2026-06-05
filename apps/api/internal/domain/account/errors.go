package domainaccount

import (
	"errors"
)

// 用 errors.New 创建错误。每个错误有明确的名字，方便定位问题。
var (
	ErrEmptyAccount       = errors.New("account is empty")     // 账号不能为空
	ErrEmptyPassword      = errors.New("password is empty")    // 密码不能为空
	ErrEmptyNickname      = errors.New("nickname is empty")    // 昵称不能为空
	ErrHashPasswordFailed = errors.New("hash password failed") // 密码哈希失败
	ErrInvalidCredentials = errors.New("invalid credentials")  // 密码错误
	ErrAccountExists      = errors.New("account already exists")
	ErrUserNotFound       = errors.New("user not found")
)
