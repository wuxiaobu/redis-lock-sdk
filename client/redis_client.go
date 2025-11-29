package client

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/wuxiaobu/redis-lock-sdk/lock"
)

// RedisManager Redis管理器
type RedisManager struct {
	clients []*redis.Client
}

// NewRedisManager 创建Redis管理器
func NewRedisManager(addrs []string, password string, db int) *RedisManager {
	clients := make([]*redis.Client, len(addrs))
	for i, addr := range addrs {
		clients[i] = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})
	}
	return &RedisManager{clients: clients}
}

// NewSingleRedisManager 创建单节点Redis管理器
func NewSingleRedisManager(addr, password string, db int) *RedisManager {
	return NewRedisManager([]string{addr}, password, db)
}

// GetClient 获取Redis客户端
func (rm *RedisManager) GetClient(index int) (*redis.Client, error) {
	if index < 0 || index >= len(rm.clients) {
		return nil, fmt.Errorf("invalid client index: %d", index)
	}
	return rm.clients[index], nil
}

// GetAllClients 获取所有Redis客户端
func (rm *RedisManager) GetAllClients() []*redis.Client {
	return rm.clients
}

// HealthCheck 健康检查
func (rm *RedisManager) HealthCheck(ctx context.Context) map[string]error {
	errors := make(map[string]error)
	for _, client := range rm.clients {
		_, err := client.Ping(ctx).Result()
		errors[client.Options().Addr] = err
		if err != nil {
			fmt.Printf("Redis %s health check failed: %v\n", client.Options().Addr, err)
		} else {
			fmt.Printf("Redis %s health check passed\n", client.Options().Addr)
		}
	}
	return errors
}

// Close 关闭所有Redis连接
func (rm *RedisManager) Close() error {
	var lastErr error
	for _, client := range rm.clients {
		if err := client.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// LockManager 创建锁管理器
func (rm *RedisManager) LockManager() lock.LockManager {
	return &RedisLockManager{rm: rm}
}

// RedisLockManager Redis锁管理器
type RedisLockManager struct {
	rm *RedisManager
}

// NewLock 创建新的分布式锁
func (rlm *RedisLockManager) NewLock(key string, opts ...lock.Option) lock.DistributedLock {
	client, _ := rlm.rm.GetClient(0) // 默认使用第一个客户端
	return lock.NewRedisLock(client, key, opts...)
}

// NewRedLock 创建RedLock（多节点锁）
func (rlm *RedisLockManager) NewRedLock(key string, opts ...lock.Option) lock.DistributedLock {
	clients := rlm.rm.GetAllClients()
	return lock.NewRedLock(clients, key, opts...)
}
