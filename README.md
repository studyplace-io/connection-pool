### connection-pool
### 介绍
`connection-pool`是基于golang实现的连接池，让调用者在使用中间件的连接时，达到限制过多连接的问题。调用方只需要在初始化后，使用`GetConnection`,`ReleaseConnection`即可获取与释放连接。

![](https://github.com/studyplace-io/connection-pool/blob/main/image/%E6%97%A0%E6%A0%87%E9%A2%98-2023-08-10-2343.png?raw=true)

### 项目功能
- 自定义连接数量
- 自定义获取连接超时时间
- 自定义空闲连接时间(超过时间会内部自动回收连接)
- 自定义心跳检查时间(内部定时检查心跳与检查连接数量)
- 支持**mysql** **redis** **etcd**连接池

### 使用
- mysql模式
```go
func main() {
	cfg := &config.ConnectionConfig{
		MaxConnections:      10,
		MaxIdleTime:         600 * time.Second,
		Timeout:             10 * time.Second,
		HealthCheckInterval: 2 * time.Second,
		CleanupInterval:     10 * time.Second,
	}

	// 创建 MySQL 连接池
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
```
- redis模式
```go
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

```
- etcd模式
```go
func main() {

	cfg := &config.ConnectionConfig{
		MaxConnections:      10,
		MaxIdleTime:         600 * time.Second,
		Timeout:             10 * time.Second,
		HealthCheckInterval: 2 * time.Second,
		CleanupInterval:     10 * time.Second,
	}

	etcdCfg := clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	// 创建 ETCD 连接池
	etcdPool := connection_pool.NewConnectionPool(connection_pool.EtcdMode(etcdCfg, cfg))
	defer etcdPool.Close()

	// 从 ETCD 连接池获取连接
	etcdConn, err := etcdPool.GetConnection()
	if err != nil {
		log.Fatal("Failed to get ETCD connection:", err)
	}
	
	// 获取后需要先转回连接对象
	etcdClient := etcdConn.(*clientv3.Client)
	defer etcdPool.ReleaseConnection(etcdConn)

	// 执行 ETCD 操作
	_, err = etcdClient.Put(context.Background(), "aaa", "aaa")
	if err != nil {
		log.Fatal("Failed to set ETCD key:", err)
	}

	rr, err := etcdClient.Get(context.Background(), "aaa")
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	fmt.Println(rr.Kvs[0].String())
}

```