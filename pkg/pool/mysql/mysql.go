package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golanglearning/new_project/connection-pool/pkg/pool/config"
	"sync"
	"time"
)

// MySQLConnectionPool 实现 ConnectionPool 接口，用于 MySQL 连接池
type MySQLConnectionPool struct {
	pool          chan *sql.DB
	config        *config.ConnectionConfig
	mysqlOpts     *mysqlOpt
	mu            sync.Mutex
	connectionNum int
	lastAccessed  map[*sql.DB]time.Time
}

type mysqlOpt struct {
	driver string
	dsn    string
}

// NewMySQLConnectionPool 创建 MySQL 连接池
func NewMySQLConnectionPool(driver, dsn string, cfg *config.ConnectionConfig) (*MySQLConnectionPool, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	pool := make(chan *sql.DB, cfg.MaxConnections)
	for i := 0; i < cfg.MaxConnections; i++ {
		conn := db
		pool <- conn
	}

	p := &MySQLConnectionPool{
		pool:         pool,
		config:       cfg,
		mysqlOpts:    &mysqlOpt{driver: driver, dsn: dsn},
		lastAccessed: make(map[*sql.DB]time.Time),
	}

	// 启动定时任务
	go p.startCleanupTask(cfg.CleanupInterval)
	go p.startCheckAndModifyConnectionNum()
	go p.startHealthCheckTask()

	return p, nil
}

// GetConnection 从 MySQL 连接池获取连接
func (p *MySQLConnectionPool) GetConnection() (interface{}, error) {
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

// ReleaseConnection 释放 MySQL 连接到连接池
func (p *MySQLConnectionPool) ReleaseConnection(conn interface{}) {
	p.pool <- conn.(*sql.DB)
}

// ReclaimConnections 回收空闲连接
func (p *MySQLConnectionPool) ReclaimConnections() {
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
func (p *MySQLConnectionPool) startCleanupTask(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		p.ReclaimConnections()
	}
}

// Close 关闭 MySQL 连接池
func (p *MySQLConnectionPool) Close() {
	close(p.pool)
}

// startHealthCheckTask 启动定时任务来定期进行连接池的健康检查
func (p *MySQLConnectionPool) startHealthCheckTask() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkConnectionsHealth()
	}
}

// checkConnectionsHealth 检查连接池中连接的健康状态
func (p *MySQLConnectionPool) checkConnectionsHealth() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for conn := range p.lastAccessed {
		// 健康检查逻辑
		if err := conn.Ping(); err != nil {
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
func (p *MySQLConnectionPool) startCheckAndModifyConnectionNum() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkAndModifyConnectionNum()
	}
}

// checkAndModifyConnectionNum 检查连接池中连接数量，不够则创建
func (p *MySQLConnectionPool) checkAndModifyConnectionNum() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if (p.config.MaxConnections - p.connectionNum) <= 0 {
		return
	}
	newConnectionNum := p.config.MaxConnections - p.connectionNum
	db, err := sql.Open(p.mysqlOpts.driver, p.mysqlOpts.dsn)
	if err != nil {
		return
	}
	for i := 0; i < newConnectionNum; i++ {
		conn := db
		p.pool <- conn
	}
}
