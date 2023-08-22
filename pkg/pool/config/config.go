package config

import "time"

// ConnectionConfig 连接池通用配置
type ConnectionConfig struct {
	// MaxConnections 最大连接数量
	MaxConnections int
	// Timeout 获取连接时的超时时间
	Timeout time.Duration
	// MaxIdleTime 连接最长空闲时间
	MaxIdleTime time.Duration
	// HealthCheckInterval 心跳检查时间
	HealthCheckInterval time.Duration
	// CleanupInterval 清理空闲连接触发时间
	CleanupInterval time.Duration
}
