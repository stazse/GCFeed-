package domainrelation

import (
	"time"
)

const (
	FollowStatusActive   = 1
	FollowStatusCanceled = 0
)

// Follow 表示一条关注关系。
type Follow struct {
	FollowerID int64 // 谁关注的
	FolloweeID int64 // 被关注的
	Status     int
	CreatedAt  time.Time
}

// NewFollow 创建一个合法的关注。
func NewFollow(followerID, followeeID int64) (*Follow, error) {
	if followerID <= 0 || followeeID <= 0 {
		return nil, ErrInvalidUserID
	}
	if followerID == followeeID {
		return nil, ErrCannotFollowSelf // 自己不能关注自己
	}
	return &Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		Status:     FollowStatusActive,
		CreatedAt:  time.Now().UTC(),
	}, nil
}