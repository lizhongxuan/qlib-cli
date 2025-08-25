package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	// 允许的域名
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:8080", 
		"http://127.0.0.1:3000",
		"http://127.0.0.1:8080",
	}
	
	// 允许的HTTP方法
	config.AllowMethods = []string{
		"GET", 
		"POST", 
		"PUT", 
		"DELETE", 
		"OPTIONS", 
		"PATCH",
	}
	
	// 允许的请求头
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
		"X-Request-ID",
		"X-User-ID",
	}
	
	// 暴露的响应头
	config.ExposeHeaders = []string{
		"Content-Length",
		"Content-Type",
		"X-Request-ID",
	}
	
	// 允许携带认证信息
	config.AllowCredentials = true
	
	// 预检请求的缓存时间
	config.MaxAge = 86400 // 24小时
	
	return cors.New(config)
}

// CORSWithConfig 自定义CORS配置
func CORSWithConfig(allowOrigins []string, allowMethods []string, allowHeaders []string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	if len(allowOrigins) > 0 {
		config.AllowOrigins = allowOrigins
	} else {
		config.AllowAllOrigins = true
	}
	
	if len(allowMethods) > 0 {
		config.AllowMethods = allowMethods
	}
	
	if len(allowHeaders) > 0 {
		config.AllowHeaders = allowHeaders
	}
	
	config.AllowCredentials = true
	config.MaxAge = 86400
	
	return cors.New(config)
}