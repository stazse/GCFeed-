package domainrelation

import "context"

type Repository interface {
	// Follow 关注。已关注的重复调用幂等。
	Follow(ctx context.Context, followerID, followeeID int64) error

	// Unfollow 取关。
	Unfollow(ctx context.Context, followerID, followeeID int64) error

	// ListFollowing 关注列表（游标分页）。
	ListFollowing(ctx context.Context, userID int64, cursor int64, limit int) ([]*Follow, error)

	// ListFollowers 粉丝列表。
	ListFollowers(ctx context.Context, userID int64, cursor int64, limit int) ([]*Follow, error)

	// IsFollowing 检查是否已关注。
	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)
}