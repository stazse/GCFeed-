package infrarelation

import (
	"context"

	domainrelation "GCFeed/internal/domain/relation"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

// Follow 关注（事务：写关系 + 更新双方计数）。
func (r *Repository) Follow(ctx context.Context, followerID, followeeID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 幂等：用结构体条件（GORM 才会把条件值合并到 INSERT 中）
		var follow FollowModel
		result := tx.Where(FollowModel{
			FollowerID: followerID,
			FolloweeID: followeeID,
		}).Attrs(FollowModel{
			Status: domainrelation.FollowStatusActive,
		}).FirstOrCreate(&follow)
		if result.Error != nil {
			return result.Error
		}

		// RowsAffected == 0 → 记录已存在
		if result.RowsAffected == 0 {
			// 已经是关注状态 → 幂等返回
			if follow.Status == domainrelation.FollowStatusActive {
				return nil
			}
			// 重新激活（status: 0 → 1）
			if err := tx.Model(&follow).Update("status", domainrelation.FollowStatusActive).Error; err != nil {
				return err
			}
		}
		// RowsAffected == 1 → 新创建（状态已是 Active，无需更新）
		// 两种情况都需要更新计数

		// 更新双方计数
		r.ensureStat(tx, followerID)
		r.ensureStat(tx, followeeID)
		tx.Model(&RelationStatModel{}).Where("user_id = ?", followerID).
			Update("following_count", gorm.Expr("following_count + 1"))
		tx.Model(&RelationStatModel{}).Where("user_id = ?", followeeID).
			Update("follower_count", gorm.Expr("follower_count + 1"))
		return nil
	})
}

// Unfollow 取关（事务：更新关系 + 更新双方计数）。
func (r *Repository) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&FollowModel{}).
			Where("follower_id = ? AND followee_id = ? AND status = ?",
				followerID, followeeID, domainrelation.FollowStatusActive).
			Update("status", domainrelation.FollowStatusCanceled)
		if result.RowsAffected == 0 {
			return domainrelation.ErrNotFollowing
		}

		tx.Model(&RelationStatModel{}).Where("user_id = ?", followerID).
			Update("following_count", gorm.Expr("GREATEST(following_count - 1, 0)"))
		tx.Model(&RelationStatModel{}).Where("user_id = ?", followeeID).
			Update("follower_count", gorm.Expr("GREATEST(follower_count - 1, 0)"))
		return nil
	})
}

// ListFollowing 关注列表。
func (r *Repository) ListFollowing(ctx context.Context, userID int64, cursor int64, limit int) ([]*domainrelation.Follow, error) {
	var models []*FollowModel
	query := r.db.WithContext(ctx).
		Where("follower_id = ? AND status = ?", userID, domainrelation.FollowStatusActive)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Order("id DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	return toDomainFollows(models), nil
}

// ListFollowers 粉丝列表。
func (r *Repository) ListFollowers(ctx context.Context, userID int64, cursor int64, limit int) ([]*domainrelation.Follow, error) {
	var models []*FollowModel
	query := r.db.WithContext(ctx).
		Where("followee_id = ? AND status = ?", userID, domainrelation.FollowStatusActive)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Order("id DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	return toDomainFollows(models), nil
}

// IsFollowing 检查是否已关注。
func (r *Repository) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&FollowModel{}).
		Where("follower_id = ? AND followee_id = ? AND status = ?",
			followerID, followeeID, domainrelation.FollowStatusActive).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) ensureStat(tx *gorm.DB, userID int64) {
	tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&RelationStatModel{UserID: userID})
}

func toDomainFollows(models []*FollowModel) []*domainrelation.Follow {
	result := make([]*domainrelation.Follow, 0, len(models))
	for _, m := range models {
		result = append(result, &domainrelation.Follow{
			FollowerID: m.FollowerID, FolloweeID: m.FolloweeID,
			Status: m.Status, CreatedAt: m.CreatedAt,
		})
	}
	return result
}

var _ domainrelation.Repository = (*Repository)(nil)