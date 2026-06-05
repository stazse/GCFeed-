package domainaccount

import "context"

type Repository interface {
	// Save 保存一个新用户
	Save(ctx context.Context, user *User) error

	// FindByAccount 根据账号查找用户，登陆使用
	FindByAccount(ctx context.Context, account string) (*User, error)

	// FindByID 根据用户ID查找用户
	FindByID(ctx context.Context, userID uint) (*User, error)
}
