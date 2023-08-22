package connection_pool

import (
	"context"
	"database/sql"
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"golanglearning/new_project/connection-pool/pkg/pool/config"
	"log"
	"testing"
	"time"
)

func TestConnectionPoll(t *testing.T) {

	cfg := &config.ConnectionConfig{
		MaxConnections:      10,
		MaxIdleTime:         600 * time.Second,
		Timeout:             10 * time.Second,
		HealthCheckInterval: 2 * time.Second,
		CleanupInterval:     10 * time.Second,
	}

	// 创建 MySQL 连接池
	mysqlPool := NewConnectionPool(MysqlMode("mysql", "root:1234567@tcp(127.0.0.1:3306)/testdb", cfg))
	defer mysqlPool.Close()

	// 从 MySQL 连接池获取连接
	mysqlConn, err := mysqlPool.GetConnection()
	if err != nil {
		log.Fatal("Failed to get MySQL connection:", err)
	}

	mysqlDB := mysqlConn.(*sql.DB)
	defer mysqlPool.ReleaseConnection(mysqlConn)

	// 执行数据库查询操作
	rows, err := mysqlDB.Query("SELECT * FROM example")
	if err != nil {
		log.Fatal("Failed to execute MySQL query:", err)
	}
	fmt.Println(rows.Columns())
	defer rows.Close()

	// 创建 Redis 连接池
	redisPool := NewConnectionPool(RedisMode("127.0.0.1:6379", "", cfg))
	defer redisPool.Close()

	// 从 Redis 连接池获取连接
	redisConn, err := redisPool.GetConnection()
	if err != nil {
		log.Fatal("Failed to get Redis connection:", err)
	}
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
