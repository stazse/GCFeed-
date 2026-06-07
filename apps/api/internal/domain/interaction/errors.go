package domaininteraction

import "errors"

var (
	ErrInvalidVideoID          = errors.New("video id must be positive")
	ErrInvalidUserID           = errors.New("user id must be positive")
	ErrEmptyContent            = errors.New("comment content is required")
	ErrCommentNotFound         = errors.New("comment not found")
	ErrNotCommentOwner         = errors.New("not the owner of this comment")
	ErrIdempotencyKeyTooLong   = errors.New("idempotency key too long")
)