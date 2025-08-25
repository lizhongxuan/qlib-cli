package config

import (
	"log"
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Qlib     QlibConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Name string
	Port string
	Mode string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Charset  string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string
	Expire int // 小时
}

// QlibConfig Qlib配置
type QlibConfig struct {
	PythonPath string
	DataPath   string
	CachePath  string
}

// Load 加载配置
func Load() *Config {
	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "qlib-backend"),
			Port: getEnv("APP_PORT", "8000"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			Username: getEnv("DB_USERNAME", "root"),
			Password: getEnv("DB_PASSWORD", "password"),
			Database: getEnv("DB_DATABASE", "qlib"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "qlib-secret-key"),
			Expire: getEnvInt("JWT_EXPIRE", 24),
		},
		Qlib: QlibConfig{
			PythonPath: getEnv("QLIB_PYTHON_PATH", "/usr/bin/python3"),
			DataPath:   getEnv("QLIB_DATA_PATH", "~/.qlib/qlib_data"),
			CachePath:  getEnv("QLIB_CACHE_PATH", "~/.qlib/cache"),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整型环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}