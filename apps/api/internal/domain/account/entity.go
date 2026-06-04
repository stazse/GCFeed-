package domainaccount

import (
	"strings" // 字符串处理函数

	"golang.org/x/crypto/bcrypt" // bcrypt 密码哈希
)

// 常量：用有意义的单词代替魔法数字
const (
	RoleUser  = "user"  // 普通用户角色
	RoleAdmin = "admin" // 管理员角色

	StatusNormal = 1 // 正常状态
)

// User 是账户聚合根。
// "聚合根" 是 DDD 术语，你可以简单理解为"这个模块最重要的那个对象"。
type User struct {
	ID        int64  // 数据库自增 ID
	Account   string // 登录账号
	Password  string // 密码的 bcrypt 哈希（不是明文！）
	Nickname  string // 显示昵称
	AvatarURL string // 头像地址
	Bio       string // 个人简介
	Status    int    // 状态：1=正常
	Role      string // 角色：user/admin
}

// New 创建新用户（注册时用）。
// 注意：这个函数不碰数据库，只做"输入对不对"的检查。
// 这是领域层的职责：定义业务规则。
func New(account, password, nickname string) (*User, error) {
	//第一步：清理输入（去掉首尾空格）：
	account = strings.TrimSpace(account)
	password = strings.TrimSpace(password)
	nickname = strings.TrimSpace(nickname)
	
	//第二步：逐项检查账号是否已存在：
	if account == "" {
		return nil, ErrEmptyAccount//账号不能为空
	}
	if password == "" {
		return nil, ErrEmptyPassword//密码不能为空
	}
	if nickname == "" {
		return nil, ErrEmptyNickname//昵称不能为空
	}

	//第三步：密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrHashPasswordFailed//密码哈希失败	
	}

	//第四步：组装用户对象。返回
	return &User{
		Account:   account,
		Password:  string(hashedPassword), // 哈希后的密码
		Nickname:  nickname,
		Role:      RoleUser,              // 默认角色通用户角色
		Status:    StatusNormal,          // 默认状态正常
	}, nil
}

//Authenticate 登陆校验密码是否正确
func (u *User) Authenticate(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return ErrEmptyPassword//密码不能为空
	}
	// bcrypt.CompareHashAndPassword：把用户输入的密码跟数据库存的哈希比对
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return ErrInvalidCredentials//密码错误
	}
	return nil
}