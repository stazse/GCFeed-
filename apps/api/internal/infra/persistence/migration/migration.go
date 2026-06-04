package migration

import (
	infraaccount "GCFeed/internal/infra/persistence/account"

	"gorm.io/gorm"
)

// AutoMigrate 根据所有模型自动创建/更新数据库表。
// 后续每增加新模块，就把模型加到这个列表里。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&infraaccount.AccountModel{},
	)
}