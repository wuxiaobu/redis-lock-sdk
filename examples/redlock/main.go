package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wuxiaobu/redis-lock-sdk/client"
	"github.com/wuxiaobu/redis-lock-sdk/lock"
)

func main() {
	// 创建多节点Redis管理器
	redisManager := client.NewRedisManager(
		[]string{
			"localhost:6379",
			"localhost:6380",
			"localhost:6381",
		},
		"", // password
		0,  // db
	)
	defer redisManager.Close()

	// 健康检查
	ctx := context.Background()
	redisManager.HealthCheck(ctx)

	// 创建锁管理器
	lockManager := redisManager.LockManager()

	// 创建RedLock
	redLock := lockManager.NewRedLock(
		"my_redlock",
		lock.WithExpiration(30*time.Second),
		lock.WithRetryCount(3),
	)

	// 尝试获取RedLock
	acquired, err := redLock.AcquireWithRetry(ctx, 0, 0)
	if err != nil {
		log.Fatalf("Failed to acquire redlock: %v", err)
	}

	if acquired {
		fmt.Println("✅ Successfully acquired redlock")

		defer func() {
			if err := redLock.Release(ctx); err != nil {
				log.Printf("Failed to release redlock: %v", err)
			} else {
				fmt.Println("✅ Successfully released redlock")
			}
		}()

		// 执行关键业务逻辑
		doCriticalBusinessLogic()
	} else {
		fmt.Println("❌ Failed to acquire redlock")
	}
}

func doCriticalBusinessLogic() {
	fmt.Println("🔄 Doing critical business logic...")
	time.Sleep(5 * time.Second)
	fmt.Println("✅ Critical business logic completed")
}
