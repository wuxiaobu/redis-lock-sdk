package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedLock 多节点Redis锁实现
type RedLock struct {
	locks   []*RedisLock
	key     string
	options Options
}

// NewRedLock 创建RedLock实例
func NewRedLock(clients []*redis.Client, key string, opts ...Option) *RedLock {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	locks := make([]*RedisLock, len(clients))
	for i, client := range clients {
		locks[i] = NewRedisLock(client, key, opts...)
	}

	return &RedLock{
		locks:   locks,
		key:     key,
		options: options,
	}
}

// Acquire 获取锁（需要大多数节点成功）
func (rl *RedLock) Acquire(ctx context.Context) (bool, error) {
	successCount := 0
	errors := make([]error, 0)

	for _, lock := range rl.locks {
		acquired, err := lock.Acquire(ctx)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if acquired {
			successCount++
		}
	}

	// 需要超过半数的节点成功
	quorum := len(rl.locks)/2 + 1
	if successCount >= quorum {
		return true, nil
	}

	// 如果获取失败，释放已经获取的锁
	rl.releaseAcquiredLocks(ctx)
	return false, fmt.Errorf("failed to acquire lock on enough nodes: %d/%d, errors: %v", successCount, quorum, errors)
}

// releaseAcquiredLocks 释放已经获取的锁
func (rl *RedLock) releaseAcquiredLocks(ctx context.Context) {
	for _, lock := range rl.locks {
		lock.Release(ctx) // 忽略错误，尽力释放
	}
}

// TryAcquire 尝试获取锁
func (rl *RedLock) TryAcquire(ctx context.Context) (bool, error) {
	return rl.Acquire(ctx)
}

// AcquireWithRetry 带重试获取锁
func (rl *RedLock) AcquireWithRetry(ctx context.Context, retryCount int, retryDelay time.Duration) (bool, error) {
	if retryCount <= 0 {
		retryCount = rl.options.RetryCount
	}
	if retryDelay <= 0 {
		retryDelay = rl.options.RetryDelay
	}

	for i := 0; i < retryCount; i++ {
		acquired, err := rl.Acquire(ctx)
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}

		if i < retryCount-1 {
			select {
			case <-time.After(retryDelay):
				// 继续重试
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}
	return false, nil
}

// TryAcquireWithTimeout 带超时获取锁
func (rl *RedLock) TryAcquireWithTimeout(ctx context.Context, timeout time.Duration) (bool, error) {
	if timeout <= 0 {
		timeout = rl.options.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return rl.AcquireWithRetry(ctx, 0, 100*time.Millisecond)
}

// Release 释放锁
func (rl *RedLock) Release(ctx context.Context) error {
	errors := make([]error, 0)
	successCount := 0

	for _, lock := range rl.locks {
		err := lock.Release(ctx)
		if err != nil && err != ErrLockNotHeld {
			errors = append(errors, err)
		} else {
			successCount++
		}
	}

	// 需要超过半数的节点成功释放
	quorum := len(rl.locks)/2 + 1
	if successCount >= quorum {
		return nil
	}

	return fmt.Errorf("failed to release lock on enough nodes: %d/%d, errors: %v", successCount, quorum, errors)
}

// GetLockKey 获取锁的key
func (rl *RedLock) GetLockKey() string {
	return rl.key
}

// GetLockValue 获取锁的value
func (rl *RedLock) GetLockValue() string {
	if len(rl.locks) > 0 {
		return rl.locks[0].GetLockValue()
	}
	return ""
}
