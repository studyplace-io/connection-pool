package pool

// ConnectionPool 接口定义连接池方法
type ConnectionPool interface {
	// GetConnection 获取连接实例
	GetConnection() (interface{}, error)
	// ReleaseConnection 释放连接实例
	ReleaseConnection(interface{})
	// Close 关闭连接池
	Close()
}
