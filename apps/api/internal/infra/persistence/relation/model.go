package infrarelation

import "time"

type FollowModel struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement"`
	FollowerID int64     `gorm:"column:follower_id;not null;uniqueIndex:idx_follow"`
	FolloweeID int64     `gorm:"column:followee_id;not null;uniqueIndex:idx_follow"`
	Status     int       `gorm:"column:status;not null;default:1"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (FollowModel) TableName() string {
	return "user_follow"
}

type RelationStatModel struct {
	UserID         int64 `gorm:"column:user_id;primaryKey"`
	FollowingCount int   `gorm:"column:following_count;default:0"`
	FollowerCount  int   `gorm:"column:follower_count;default:0"`
}

func (RelationStatModel) TableName() string {
	return "user_relation_stat"
}