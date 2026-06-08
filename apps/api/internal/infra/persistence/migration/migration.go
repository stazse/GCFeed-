package migration

import (
	infraaccount "GCFeed/internal/infra/persistence/account"
	infrainteraction "GCFeed/internal/infra/persistence/interaction"
	infrarelation "GCFeed/internal/infra/persistence/relation"
	infravideo "GCFeed/internal/infra/persistence/video"

	"gorm.io/gorm"
)

// AutoMigrate 根据所有模型自动创建/更新数据库表。
// 后续每增加新模块，就把模型加到这个列表里。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&infraaccount.AccountModel{},
		&infravideo.VideoModel{},
		&infravideo.VideoStatModel{},
		&infrainteraction.ActionModel{},
		&infrainteraction.CommentModel{},
		&infrarelation.FollowModel{},
		&infrarelation.RelationStatModel{},
	)
}
