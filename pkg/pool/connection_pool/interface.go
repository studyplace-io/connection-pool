package connection_pool

import (
	"sync"
)

// IConnectionPool 接口定义连接池方法
type IConnectionPool interface {
	// GetConnection 获取连接实例
	GetConnection() (interface{}, error)
	// ReleaseConnection 释放连接实例
	ReleaseConnection(interface{})
	// Close 关闭连接池
	Close()
}

// ConnectionPool 连接池对象
type ConnectionPool struct {
	// ConnectionPool 连接池接口对象
	ConnectionPool IConnectionPool
	lock           sync.Mutex
}

func NewConnectionPool(connectionPool IConnectionPool) *ConnectionPool {
	return &ConnectionPool{
		ConnectionPool: connectionPool,
		lock:           sync.Mutex{},
	}
}

// GetConnection 获取连接实例
func (c *ConnectionPool) GetConnection() (interface{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.ConnectionPool.GetConnection()
}

// ReleaseConnection 释放连接实例
func (c *ConnectionPool) ReleaseConnection(connection interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ConnectionPool.ReleaseConnection(connection)
}

// Close 关闭连接池
func (c *ConnectionPool) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ConnectionPool.Close()
}
