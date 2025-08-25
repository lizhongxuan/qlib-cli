package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器结构
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.Mutex
	rate     int           // 每秒允许的请求数
	burst    int           // 突发请求数
	ttl      time.Duration // 访客记录的生存时间
}

// Visitor 访客信息
type Visitor struct {
	tokens    int       // 当前令牌数
	lastToken time.Time // 上次添加令牌时间
	lastSeen  time.Time // 上次访问时间
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(rate, burst int, ttl time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
		ttl:      ttl,
	}

	// 启动清理goroutine
	go rl.cleanupVisitors()
	
	return rl
}

// cleanupVisitors 定期清理过期的访客记录
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, visitor := range rl.visitors {
				if now.Sub(visitor.lastSeen) > rl.ttl {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	visitor, exists := rl.visitors[ip]

	if !exists {
		// 新访客，分配初始令牌
		visitor = &Visitor{
			tokens:    rl.burst - 1, // 使用一个令牌处理当前请求
			lastToken: now,
			lastSeen:  now,
		}
		rl.visitors[ip] = visitor
		return true
	}

	// 更新访客的最后访问时间
	visitor.lastSeen = now

	// 计算应该添加的令牌数
	elapsed := now.Sub(visitor.lastToken)
	tokensToAdd := int(elapsed.Seconds()) * rl.rate

	if tokensToAdd > 0 {
		visitor.tokens += tokensToAdd
		if visitor.tokens > rl.burst {
			visitor.tokens = rl.burst
		}
		visitor.lastToken = now
	}

	// 检查是否有足够的令牌
	if visitor.tokens > 0 {
		visitor.tokens--
		return true
	}

	return false
}

// RateLimit 限流中间件
func RateLimit(rate, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst, time.Hour) // 1小时TTL

	return func(c *gin.Context) {
		// 获取客户端IP
		ip := getClientIP(c)

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithConfig 自定义配置的限流中间件
func RateLimitWithConfig(rate, burst int, ttl time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst, ttl)

	return func(c *gin.Context) {
		ip := getClientIP(c)

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"code":    http.StatusTooManyRequests,
				"message": "请求过于频繁，请稍后再试",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIRateLimit API专用限流中间件（更严格的限制）
func APIRateLimit() gin.HandlerFunc {
	// API限制：每秒10个请求，突发20个
	return RateLimit(10, 20)
}

// WebSocketRateLimit WebSocket连接限流中间件
func WebSocketRateLimit() gin.HandlerFunc {
	// WebSocket限制：每秒5个连接请求，突发10个
	return RateLimit(5, 10)
}

// UploadRateLimit 文件上传限流中间件
func UploadRateLimit() gin.HandlerFunc {
	// 上传限制：每分钟5个请求，突发2个
	limiter := NewRateLimiter(5/60, 2, time.Hour) // 每分钟5个转换为每秒

	return func(c *gin.Context) {
		ip := getClientIP(c)

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"code":    http.StatusTooManyRequests,
				"message": "文件上传过于频繁，请稍后再试",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP 获取客户端真实IP
func getClientIP(c *gin.Context) string {
	// 检查X-Forwarded-For header
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		return xForwardedFor
	}

	// 检查X-Real-IP header
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 检查X-Forwarded header
	xForwarded := c.GetHeader("X-Forwarded")
	if xForwarded != "" {
		return xForwarded
	}

	// 检查Forwarded-For header
	forwardedFor := c.GetHeader("Forwarded-For")
	if forwardedFor != "" {
		return forwardedFor
	}

	// 检查Forwarded header
	forwarded := c.GetHeader("Forwarded")
	if forwarded != "" {
		return forwarded
	}

	// 使用RemoteAddr
	return c.RemoteIP()
}

// RateLimiterMiddleware 限流中间件
func RateLimiterMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		rl.mu.Lock()
		visitor, exists := rl.visitors[clientIP]
		if !exists {
			visitor = &Visitor{
				tokens:    rl.burst,
				lastToken: time.Now(),
				lastSeen:  time.Now(),
			}
			rl.visitors[clientIP] = visitor
		}
		
		// 更新最后访问时间
		visitor.lastSeen = time.Now()
		
		// 补充令牌
		now := time.Now()
		elapsed := now.Sub(visitor.lastToken)
		tokensToAdd := int(elapsed.Seconds()) * rl.rate
		visitor.tokens += tokensToAdd
		if visitor.tokens > rl.burst {
			visitor.tokens = rl.burst
		}
		visitor.lastToken = now
		
		// 检查是否有令牌
		if visitor.tokens > 0 {
			visitor.tokens--
			rl.mu.Unlock()
			c.Next()
		} else {
			rl.mu.Unlock()
			c.JSON(429, gin.H{
				"success": false,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
		}
	}
}

// RateLimiterByIP 基于IP的限流中间件
func RateLimiterByIP(rate, burst int, ttl time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(rate, burst, ttl)
	return RateLimiterMiddleware(rl)
}

// RateLimiterByUser 基于用户的限流中间件
func RateLimiterByUser(rate, burst int, ttl time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(rate, burst, ttl)
	
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"success": false,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}
		
		key := fmt.Sprintf("user_%v", userID)
		
		rl.mu.Lock()
		visitor, exists := rl.visitors[key]
		if !exists {
			visitor = &Visitor{
				tokens:    rl.burst,
				lastToken: time.Now(),
				lastSeen:  time.Now(),
			}
			rl.visitors[key] = visitor
		}
		
		// 更新最后访问时间
		visitor.lastSeen = time.Now()
		
		// 补充令牌
		now := time.Now()
		elapsed := now.Sub(visitor.lastToken)
		tokensToAdd := int(elapsed.Seconds()) * rl.rate
		visitor.tokens += tokensToAdd
		if visitor.tokens > rl.burst {
			visitor.tokens = rl.burst
		}
		visitor.lastToken = now
		
		// 检查是否有令牌
		if visitor.tokens > 0 {
			visitor.tokens--
			rl.mu.Unlock()
			c.Next()
		} else {
			rl.mu.Unlock()
			c.JSON(429, gin.H{
				"success": false,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
		}
	}
}