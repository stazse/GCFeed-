package domainrelation

import "errors"

var (
	ErrInvalidUserID    = errors.New("user id must be positive")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrAlreadyFollowing = errors.New("already following")
	ErrNotFollowing     = errors.New("not following this user")
)