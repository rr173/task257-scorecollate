package model

import "errors"

// 领域错误集合，供仓储与服务层统一返回。
var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidState = errors.New("invalid state")
	ErrConflict     = errors.New("concurrent version conflict")
	ErrDuplicate    = errors.New("duplicate resource")
	ErrInvalidInput = errors.New("invalid input")
	ErrSealed       = errors.New("project is sealed and cannot be modified")
	ErrFrozen       = errors.New("edition is frozen and cannot be modified")
	ErrSelfCycle    = errors.New("source genealogy self-cycle detected")
	ErrUnreadable   = errors.New("fragment is unreadable and cannot be aligned")
)
