package main

import (
	"database/sql"
	"fmt"
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

	//// 创建 MySQL 连接池
	mysqlPool := connection_pool.NewConnectionPool(connection_pool.MysqlMode("mysql", "root:1234567@tcp(127.0.0.1:3306)/testdb", cfg))
	defer mysqlPool.Close()

	// 从 MySQL 连接池获取连接
	mysqlConn, err := mysqlPool.GetConnection()
	if err != nil {
		log.Fatal("Failed to get MySQL connection:", err)
	}

	// 获取后需要先转回连接对象
	mysqlDB := mysqlConn.(*sql.DB)
	defer mysqlPool.ReleaseConnection(mysqlConn)

	// 执行数据库查询操作
	rows, err := mysqlDB.Query("SELECT * FROM example")
	if err != nil {
		log.Fatal("Failed to execute MySQL query:", err)
	}
	fmt.Println(rows.Columns())
	defer rows.Close()
}
