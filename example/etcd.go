package main

import (
	"context"
	"fmt"
	"github.com/practice/connection-pool/pkg/pool/config"
	"github.com/practice/connection-pool/pkg/pool/connection_pool"
	clientv3 "go.etcd.io/etcd/client/v3"
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
