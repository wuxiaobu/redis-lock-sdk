package lock

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisLock Redis分布式锁实现
type RedisLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
	options    Options
}

// NewRedisLock 创建Redis锁实例
func NewRedisLock(client *redis.Client, key string, opts ...Option) *RedisLock {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &RedisLock{
		client:     client,
		key:        key,
		value:      generateLockValue(options.ValuePrefix),
		expiration: options.Expiration,
		options:    options,
	}
}

// generateLockValue 生成锁的唯一值
func generateLockValue(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s:%d:%d", prefix, time.Now().UnixNano(), rand.Int63())
}

// Acquire 获取锁
func (rl *RedisLock) Acquire(ctx context.Context) (bool, error) {
	result, err := rl.client.SetNX(ctx, rl.key, rl.value, rl.expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return result, nil
}

// TryAcquire 尝试获取锁（立即返回）
func (rl *RedisLock) TryAcquire(ctx context.Context) (bool, error) {
	return rl.Acquire(ctx)
}

// AcquireWithRetry 带重试获取锁
func (rl *RedisLock) AcquireWithRetry(ctx context.Context, retryCount int, retryDelay time.Duration) (bool, error) {
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
func (rl *RedisLock) TryAcquireWithTimeout(ctx context.Context, timeout time.Duration) (bool, error) {
	if timeout <= 0 {
		timeout = rl.options.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
			acquired, err := rl.Acquire(ctx)
			if err != nil {
				return false, err
			}
			if acquired {
				return true, nil
			}
		}
	}
}

// Release 释放锁
func (rl *RedisLock) Release(ctx context.Context) error {
	script := `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`
	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// GetLockKey 获取锁的key
func (rl *RedisLock) GetLockKey() string {
	return rl.key
}

// GetLockValue 获取锁的value
func (rl *RedisLock) GetLockValue() string {
	return rl.value
}

// Refresh 刷新锁过期时间
func (rl *RedisLock) Refresh(ctx context.Context) error {
	script := `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("pexpire", KEYS[1], ARGV[2])
else
    return 0
end
`
	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value, rl.expiration.Milliseconds()).Result()
	if err != nil {
		return fmt.Errorf("failed to refresh lock: %w", err)
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// 错误定义
var (
	ErrLockNotHeld = fmt.Errorf("lock not held by this client")
)
