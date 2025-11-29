# Redis Distributed Lock SDK

一个高性能、易用的Go语言Redis分布式锁SDK。

## 特性

- ✅ 单节点Redis锁
- ✅ 多节点RedLock算法
- ✅ 自动重试机制
- ✅ 超时控制
- ✅ 锁自动续期
- ✅ 健康检查
- ✅ 线程安全

## 安装

```bash
go get github.com/your-username/redis-lock-sdk
```

## 快速开始
```go
package main

import (
    "context"
    "time"
    "github.com/your-username/redis-lock-sdk/client"
)

func main() {
    // 创建Redis管理器
    redisManager := client.NewSingleRedisManager("localhost:6379", "", 0)
    defer redisManager.Close()

    // 创建锁
    lockManager := redisManager.LockManager()
    lock := lockManager.NewLock(
        "my_lock",
        client.WithExpiration(10*time.Second),
        client.WithRetryCount(3),
    )

    ctx := context.Background()
    
    // 获取锁
    if acquired, err := lock.Acquire(ctx); err == nil && acquired {
        defer lock.Release(ctx)
        // 执行业务逻辑
    }
}
```

## 使用说明

这个SDK提供了：

1. **简单易用的API**：清晰的接口设计，易于理解和使用
2. **灵活的配置**：支持多种配置选项
3. **高可用性**：支持单节点和多节点部署
4. **完善的错误处理**：详细的错误信息和处理机制
5. **健康检查**：内置Redis连接健康检查
6. **自动重试**：内置重试机制，提高获取锁的成功率

你可以根据需要选择使用单节点锁或多节点RedLock，SDK会自动处理底层的复杂逻辑。
