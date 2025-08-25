package config

import (
	"fmt"
	"time"
)

// ExtendedDatabaseConfig 扩展数据库配置结构
type ExtendedDatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	Charset         string        `json:"charset"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// GetDSN 获取数据库连接字符串
func (dc *ExtendedDatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dc.Username,
		dc.Password,
		dc.Host,
		dc.Port,
		dc.Database,
		dc.Charset,
	)
}

// Validate 验证数据库配置
func (dc *ExtendedDatabaseConfig) Validate() error {
	if dc.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	if dc.Port <= 0 || dc.Port > 65535 {
		return fmt.Errorf("数据库端口必须在1-65535之间")
	}
	if dc.Username == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if dc.Database == "" {
		return fmt.Errorf("数据库名不能为空")
	}
	if dc.Charset == "" {
		dc.Charset = "utf8mb4"
	}
	if dc.MaxIdleConns <= 0 {
		dc.MaxIdleConns = 10
	}
	if dc.MaxOpenConns <= 0 {
		dc.MaxOpenConns = 100
	}
	if dc.ConnMaxLifetime <= 0 {
		dc.ConnMaxLifetime = time.Hour
	}
	if dc.ConnMaxIdleTime <= 0 {
		dc.ConnMaxIdleTime = 10 * time.Minute
	}
	return nil
}

// DefaultDatabaseConfig 默认数据库配置
func DefaultDatabaseConfig() ExtendedDatabaseConfig {
	return ExtendedDatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "password",
		Database:        "qlib",
		Charset:         "utf8mb4",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}