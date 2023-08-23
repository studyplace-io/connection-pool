package connection_pool

import (
	"github.com/practice/connection-pool/pkg/pool/config"
	"github.com/practice/connection-pool/pkg/pool/etcd"
	"github.com/practice/connection-pool/pkg/pool/mysql"
	"github.com/practice/connection-pool/pkg/pool/redis"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

// MysqlMode mysql模式
func MysqlMode(driver, dsn string, cfg *config.ConnectionConfig) IConnectionPool {
	c, err := mysql.NewMySQLConnectionPool(driver, dsn, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

// RedisMode redis模式
func RedisMode(addr, password string, cfg *config.ConnectionConfig) IConnectionPool {
	c := redis.NewRedisConnectionPool(addr, password, cfg)
	return c
}

// EtcdMode etcd模式
func EtcdMode(etcdConfig clientv3.Config, cfg *config.ConnectionConfig) IConnectionPool {
	c, err := etcd.NewETCDConnectionPool(etcdConfig, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return c
}
