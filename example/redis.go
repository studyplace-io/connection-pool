package main

import (
	"context"
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/practice/connection-pool/pkg/pool/config"
	"github.com/practice/connection-pool/pkg/pool/connection_pool"
	"log"
	"time"
)

func main() {
	cfg := &config.ConnectionConfig{
		MaxConnections:      10,
		MaxIdleTime:         600 * time.Second,
		Timeout:             10 * time.Second,
		HealthCheckInterval: 2 * time.Second,
		CleanupInterval:     10 * time.Second,
	}

	// 创建 Redis 连接池
	redisPool := connection_pool.NewConnectionPool(connection_pool.RedisMode("127.0.0.1:6379", "", cfg))
	defer redisPool.Close()

	// 从 Redis 连接池获取连接
	redisConn, err := redisPool.GetConnection()
	if err != nil {
		log.Fatal("Failed to get Redis connection:", err)
	}

	// 获取后需要先转回连接对象
	redisClient := redisConn.(*redis2.Client)
	defer redisPool.ReleaseConnection(redisConn)

	// 执行 Redis 操作
	err = redisClient.Set(context.Background(), "my-key", "my-value", 0).Err()
	if err != nil {
		log.Fatal("Failed to set Redis key:", err)
	}

	cc := redisClient.Get(context.Background(), "my-key")
	if cc.Err() != nil {
		fmt.Println("err: ", err)
		return
	}
	fmt.Println(cc.String())
}
