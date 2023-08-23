package etcd

import (
	"context"
	"fmt"
	"github.com/practice/connection-pool/pkg/pool/config"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

// ETCDConnectionPool 实现 ConnectionPool 接口，用于 ETCD 连接池
type ETCDConnectionPool struct {
	// pool 存放连接池chan
	pool chan *clientv3.Client
	// config 连接池通用配置
	config *config.ConnectionConfig
	// etcdOpts etcd私有配置，不对外暴露
	etcdOpts *etcdOpt
	// connectionNum 记录当下池中的连接数
	connectionNum int
	// lastAccessed 记录每个连接实例的最后使用时间
	lastAccessed map[*clientv3.Client]time.Time
	mu           sync.Mutex
}

type etcdOpt struct {
	config clientv3.Config
}

// NewETCDConnectionPool 创建 ETCD 连接池
func NewETCDConnectionPool(config clientv3.Config, cfg *config.ConnectionConfig) (*ETCDConnectionPool, error) {
	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	pool := make(chan *clientv3.Client, cfg.MaxConnections)
	for i := 0; i < cfg.MaxConnections; i++ {
		conn := client
		pool <- conn
	}

	p := &ETCDConnectionPool{
		pool:         pool,
		config:       cfg,
		etcdOpts:     &etcdOpt{config: config},
		lastAccessed: make(map[*clientv3.Client]time.Time),
	}

	// 启动定时任务
	go p.startCleanupTask(cfg.CleanupInterval)
	go p.startCheckAndModifyConnectionNum()
	go p.startHealthCheckTask()

	return p, nil
}

// GetConnection 从 MySQL 连接池获取连接
func (p *ETCDConnectionPool) GetConnection() (interface{}, error) {
	select {
	case conn := <-p.pool:
		p.mu.Lock()
		p.lastAccessed[conn] = time.Now()
		p.mu.Unlock()
		return conn, nil
	case <-time.After(p.config.Timeout):
		return nil, fmt.Errorf("timeout: failed to get ETCD connection")
	}
}

// ReleaseConnection 释放 MySQL 连接到连接池
func (p *ETCDConnectionPool) ReleaseConnection(conn interface{}) {
	p.pool <- conn.(*clientv3.Client)
}

// ReclaimConnections 回收空闲连接
func (p *ETCDConnectionPool) reclaimConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for conn, lastAccessed := range p.lastAccessed {
		if now.Sub(lastAccessed) > p.config.MaxIdleTime {
			conn.Close()
			delete(p.lastAccessed, conn)
		}
	}
}

// startCleanupTask 启动定时任务来定期调用回收空闲连接的方法
func (p *ETCDConnectionPool) startCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		p.reclaimConnections()
	}
}

// Close 关闭 ETCD 连接池
func (p *ETCDConnectionPool) Close() {
	close(p.pool)
}

// startHealthCheckTask 启动定时任务来定期进行连接池的健康检查
func (p *ETCDConnectionPool) startHealthCheckTask() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkConnectionsHealth()
	}
}

// checkConnectionsHealth 检查连接池中连接的健康状态
func (p *ETCDConnectionPool) checkConnectionsHealth() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for conn := range p.lastAccessed {
		// 健康检查逻辑
		if _, err := conn.Get(context.Background(), "", clientv3.WithSerializable()); err != nil {
			// 连接无效，关闭连接并从连接池中移除
			conn.Close()
			delete(p.lastAccessed, conn)
		} else if now.Sub(p.lastAccessed[conn]) > p.config.MaxIdleTime {
			// 连接超时，关闭连接并从连接池中移除
			conn.Close()
			delete(p.lastAccessed, conn)
		}
	}
}

// startCheckAndModifyConnectionNum 启动定时任务来定期进行连接池的连接数检查
func (p *ETCDConnectionPool) startCheckAndModifyConnectionNum() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkAndModifyConnectionNum()
	}
}

// checkAndModifyConnectionNum 检查连接池中连接数量，不够则创建
func (p *ETCDConnectionPool) checkAndModifyConnectionNum() {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 代表不需要新增连接
	if (p.config.MaxConnections - p.connectionNum) <= 0 {
		return
	}
	newConnectionNum := p.config.MaxConnections - p.connectionNum

	ccc, err := clientv3.New(p.etcdOpts.config)
	if err != nil {
		return
	}
	for i := 0; i < newConnectionNum; i++ {
		conn := ccc
		p.pool <- conn
	}
}
