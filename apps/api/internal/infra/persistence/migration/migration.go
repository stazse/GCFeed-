package migration

import (
	"log"

	infraaccount "GCFeed/internal/infra/persistence/account"
	infrainteraction "GCFeed/internal/infra/persistence/interaction"
	infrarelation "GCFeed/internal/infra/persistence/relation"
	infravideo "GCFeed/internal/infra/persistence/video"

	"gorm.io/gorm"
)

// AutoMigrate 根据所有模型自动创建/更新数据库表。
// 每个表独立迁移，避免一个表失败阻塞后续表。
func AutoMigrate(db *gorm.DB) error {
	models := []any{
		&infraaccount.AccountModel{},
		&infravideo.VideoModel{},
		&infravideo.VideoStatModel{},
		&infrainteraction.ActionModel{},
		&infrainteraction.CommentModel{},
		&infrarelation.FollowModel{},
		&infrarelation.RelationStatModel{},
	}

	var lastErr error
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Printf("auto-migrate %T warning: %v", m, err)
			lastErr = err
		}
	}

	return lastErr
}
