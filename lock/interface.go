package lock

import (
	"context"
	"time"
)

// DistributedLock 分布式锁接口
type DistributedLock interface {
	// Acquire 获取锁
	Acquire(ctx context.Context) (bool, error)

	// Release 释放锁
	Release(ctx context.Context) error

	// TryAcquire 尝试获取锁（立即返回）
	TryAcquire(ctx context.Context) (bool, error)

	// AcquireWithRetry 带重试获取锁
	AcquireWithRetry(ctx context.Context, retryCount int, retryDelay time.Duration) (bool, error)

	// TryAcquireWithTimeout 带超时获取锁
	TryAcquireWithTimeout(ctx context.Context, timeout time.Duration) (bool, error)

	// GetLockKey 获取锁的key
	GetLockKey() string

	// GetLockValue 获取锁的value
	GetLockValue() string
}

// LockManager 锁管理器接口
type LockManager interface {
	// NewLock 创建新的分布式锁
	NewLock(key string, opts ...Option) DistributedLock

	// NewRedLock 创建RedLock（多节点锁）
	NewRedLock(key string, opts ...Option) DistributedLock
}
