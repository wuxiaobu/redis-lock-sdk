package lock

import "time"

// Options 锁配置选项
type Options struct {
	Expiration  time.Duration // 锁过期时间
	RetryCount  int           // 重试次数
	RetryDelay  time.Duration // 重试延迟
	Timeout     time.Duration // 超时时间
	ValuePrefix string        // 锁值前缀
}

// Option 配置函数类型
type Option func(*Options)

// DefaultOptions 返回默认配置
func DefaultOptions() Options {
	return Options{
		Expiration:  30 * time.Second,
		RetryCount:  3,
		RetryDelay:  100 * time.Millisecond,
		Timeout:     5 * time.Second,
		ValuePrefix: "lock",
	}
}

// WithExpiration 设置锁过期时间
func WithExpiration(expiration time.Duration) Option {
	return func(o *Options) {
		o.Expiration = expiration
	}
}

// WithRetryCount 设置重试次数
func WithRetryCount(retryCount int) Option {
	return func(o *Options) {
		o.RetryCount = retryCount
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(retryDelay time.Duration) Option {
	return func(o *Options) {
		o.RetryDelay = retryDelay
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithValuePrefix 设置锁值前缀
func WithValuePrefix(prefix string) Option {
	return func(o *Options) {
		o.ValuePrefix = prefix
	}
}
