package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"golanglearning/new_project/connection-pool/pkg/pool/config"
	"sync"
	"time"
)

// RedisConnectionPool redis连接池，实现ConnectionPool接口
type RedisConnectionPool struct {
	// pool 存放连接池chan
	pool          chan *redis.Client
	// config 连接池通用配置
	config        *config.ConnectionConfig
	// redisOpts redis私有配置，不对外暴露
	redisOpts     *redisOpt
	// connectionNum 记录当下池中的连接数
	connectionNum int
	// lastAccessed 记录每个连接实例的最后使用时间
	lastAccessed  map[*redis.Client]time.Time
	mu            sync.Mutex
}

// redisOpt redis私有配置，不对外暴露
type redisOpt struct {
	addr     string
	password string
}

// NewRedisConnectionPool 创建 Redis 连接池
func NewRedisConnectionPool(addr, password string, cfg *config.ConnectionConfig) *RedisConnectionPool {
	// 1. 优先创建出指定连接数
	pool := make(chan *redis.Client, cfg.MaxConnections)
	for i := 0; i < cfg.MaxConnections; i++ {
		client := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		})
		pool <- client
	}

	p := &RedisConnectionPool{
		pool:          pool,
		config:        cfg,
		redisOpts:     &redisOpt{addr: addr, password: password},
		connectionNum: cfg.MaxConnections,
		lastAccessed:  make(map[*redis.Client]time.Time),
	}

	// 2. 启动定时任务
	go p.startCleanupTask(cfg.CleanupInterval)
	go p.startCheckAndModifyConnectionNum()
	go p.startHealthCheckTask()

	return p
}

// GetConnection 从 Redis 连接池获取连接
func (p *RedisConnectionPool) GetConnection() (interface{}, error) {
	select {
	case conn := <-p.pool:
		p.mu.Lock()
		p.lastAccessed[conn] = time.Now()
		p.mu.Unlock()
		return conn, nil
	case <-time.After(p.config.Timeout):
		return nil, fmt.Errorf("timeout: failed to get MySQL connection")
	}
}

// ReclaimConnections 回收空闲连接
func (p *RedisConnectionPool) ReclaimConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for conn, lastAccessed := range p.lastAccessed {
		if now.Sub(lastAccessed) > p.config.MaxIdleTime {
			conn.Close()
			delete(p.lastAccessed, conn)
			p.connectionNum--
		}
	}
}

// startCleanupTask 启动定时任务来定期调用回收空闲连接的方法
func (p *RedisConnectionPool) startCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		p.ReclaimConnections()
	}
}

// ReleaseConnection 释放 Redis 连接到连接池
func (p *RedisConnectionPool) ReleaseConnection(conn interface{}) {
	p.pool <- conn.(*redis.Client)
}

// Close 关闭 Redis 连接池
func (p *RedisConnectionPool) Close() {
	close(p.pool)
}

// startHealthCheckTask 启动定时任务来定期进行连接池的健康检查
func (p *RedisConnectionPool) startHealthCheckTask() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkConnectionsHealth()
	}
}

// checkConnectionsHealth 检查连接池中连接的健康状态
func (p *RedisConnectionPool) checkConnectionsHealth() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for conn := range p.lastAccessed {
		// 健康检查逻辑
		if err := conn.Ping(context.Background()); err != nil {
			// 连接无效，关闭连接并从连接池中移除
			conn.Close()
			delete(p.lastAccessed, conn)
			p.connectionNum--
		} else if now.Sub(p.lastAccessed[conn]) > p.config.MaxIdleTime {
			// 连接超时，关闭连接并从连接池中移除
			conn.Close()
			delete(p.lastAccessed, conn)
			p.connectionNum--
		}
	}
}

// startCheckAndModifyConnectionNum 启动定时任务来定期进行连接池的连接数检查
func (p *RedisConnectionPool) startCheckAndModifyConnectionNum() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkAndModifyConnectionNum()
	}
}

// checkAndModifyConnectionNum 检查连接池中连接数量，不够则创建
func (p *RedisConnectionPool) checkAndModifyConnectionNum() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if (p.config.MaxConnections - p.connectionNum) <= 0 {
		return
	}
	newConnectionNum := p.config.MaxConnections - p.connectionNum
	for i := 0; i < newConnectionNum; i++ {
		client := redis.NewClient(&redis.Options{
			Addr:     p.redisOpts.addr,
			Password: p.redisOpts.password,
		})
		p.pool <- client
	}
}
