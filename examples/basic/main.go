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
	// 创建Redis管理器
	redisManager := client.NewSingleRedisManager("localhost:6379", "", 0)
	defer redisManager.Close()

	// 健康检查
	ctx := context.Background()
	redisManager.HealthCheck(ctx)

	// 创建锁管理器
	lockManager := redisManager.LockManager()

	// 创建分布式锁
	distributedLock := lockManager.NewLock(
		"my_business_lock",
		lock.WithExpiration(10*time.Second),
		lock.WithRetryCount(5),
		lock.WithRetryDelay(200*time.Millisecond),
	)

	// 尝试获取锁
	acquired, err := distributedLock.AcquireWithRetry(ctx, 0, 0)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}

	if acquired {
		fmt.Println("✅ Successfully acquired lock")

		// 确保在函数退出时释放锁
		defer func() {
			if err := distributedLock.Release(ctx); err != nil {
				log.Printf("Failed to release lock: %v", err)
			} else {
				fmt.Println("✅ Successfully released lock")
			}
		}()

		// 执行需要加锁的业务逻辑
		doBusinessLogic()
	} else {
		fmt.Println("❌ Failed to acquire lock")
	}
}

func doBusinessLogic() {
	fmt.Println("🔄 Doing business logic...")
	time.Sleep(3 * time.Second)
	fmt.Println("✅ Business logic completed")
}
