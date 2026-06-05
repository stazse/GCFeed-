package domainvideo

import "errors"

var (
	ErrInvalidAuthorID       = errors.New("author id must be positive")  //作者ID必须为正数
	ErrEmptyTitle            = errors.New("title is required")           //标题不能为空
	ErrTitleTooLong          = errors.New("title is too long")           //标题不能超过 255 个字符
	ErrIdempotencyKeyTooLong = errors.New("idempotency key is too long") //幂等性键不能超过 255 个字符
	ErrVideoNotFound         = errors.New("video not found")             //视频不存在
	ErrNotVideoOwner         = errors.New("not the owner of this video") //不是视频所有者
)
