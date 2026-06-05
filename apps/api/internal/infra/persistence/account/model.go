package infraaccount

// AccountModel 是 user 表在 Go 中的映射。
// GORM 会根据这个结构体自动创建和维护数据库表。
type AccountModel struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`                    // 主键，自增
	Account   string `gorm:"column:account;type:varchar(191);uniqueIndex;not null"` // 唯一索引，不能重复
	Password  string `gorm:"column:password;type:varchar(255);not null"`            // 哈希后的密码
	Nickname  string `gorm:"column:nickname;type:varchar(64);not null"`
	AvatarURL string `gorm:"column:avatar_url;type:varchar(512);default:''"`
	Bio       string `gorm:"column:bio;type:varchar(512);default:''"`
	Status    int    `gorm:"column:status;default:1"`
	Role      string `gorm:"column:role;type:varchar(32);default:'user'"`
}

// TableName 告诉 GORM 这个模型对应哪张数据库表。
// 如果不写这个函数，GORM 默认用结构体名的复数（account_models）。
func (AccountModel) TableName() string {
	return "account"
}
