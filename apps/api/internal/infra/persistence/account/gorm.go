package infraaccount

import (
	"context"
	"errors"

	domainaccount "GCFeed/internal/domain/account"

	"gorm.io/gorm"
)

// Repository 是仓储接口的具体实现——用 GORM 操作 MySQL。
type Repository struct {
	db *gorm.DB // GORM 的数据库连接对象
}

// New 创建仓储实例（工厂函数）。
func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Save 保存一个新用户（注册）
func (r *Repository) Save(ctx context.Context, user *domainaccount.User) error {
	//把领域对象转为数据库模型
	model := &AccountModel{
		Account:   user.Account,
		Password:  user.Password,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Status:    user.Status,
		Role:      user.Role,
	}

	// r.db.WithContext(ctx)：告诉 GORM 这个操作属于哪个请求
	// Create：生成 INSERT INTO account ... 语句
	// Error：返回操作是否成功的错误
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		//如果错误是“唯一键冲突”（账号已存在），则返回领域错误
		if isDuplicateKeyError(err) {
			return domainaccount.ErrAccountExists
		}
		//其他错误，直接返回
		return err
	}

	//把数据库生成的id回填到领域对象中
	user.ID = model.ID
	return nil
}

// FindByAccount 根据账号查找用户，登陆使用
func (r *Repository) FindByAccount(ctx context.Context, account string) (*domainaccount.User, error) {
	//根据账号查询数据库
	var model AccountModel
	// Where + First：生成 SELECT * FROM account WHERE account = ? LIMIT 1
	if err := r.db.WithContext(ctx).Where("account = ?", account).First(&model).Error; err != nil {
		//如果查询失败，返回领域错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainaccount.ErrUserNotFound
		}
		return nil, err
	}
	return restoreUser(&model), nil
}

// FindByID 根据 ID 查找用户。
func (r *Repository) FindByID(ctx context.Context, userID uint) (*domainaccount.User, error) {
	//根据ID查询数据库
	var model AccountModel
	// Where + First：生成 SELECT * FROM account WHERE id = ? LIMIT 1
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&model).Error; err != nil {
		//如果查询失败，返回领域错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainaccount.ErrUserNotFound
		}
		return nil, err
	}
	return restoreUser(&model), nil
}

// restoreUser 把数据库模型还原为领域对象（内部函数）。
func restoreUser(m *AccountModel) *domainaccount.User {
	return &domainaccount.User{
		ID:        m.ID,
		Account:   m.Account,
		Password:  m.Password,
		Nickname:  m.Nickname,
		AvatarURL: m.AvatarURL,
		Bio:       m.Bio,
		Status:    m.Status,
		Role:      m.Role,
	}
}

// isDuplicateKeyError 判断是否为 MySQL 唯一键冲突（错误码 1062）。
func isDuplicateKeyError(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrDuplicatedKey)
}

// 编译期校验：确保 Repository 实现了 domainaccount.Repository 接口
// 如果没实现，编译时会报错——这是一个很好的保护机制。
var _ domainaccount.Repository = (*Repository)(nil)
