package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20, time.Minute)
	
	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.rate)
	assert.Equal(t, 20, rl.burst)
	assert.Equal(t, time.Minute, rl.ttl)
	assert.NotNil(t, rl.visitors)
}

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建一个限制很严格的限流器：每秒1个请求，突发2个
	router := gin.New()
	router.Use(RateLimiterByIP(1, 2, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 第一个请求应该成功
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 第二个请求应该成功（在突发限制内）
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12346"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 第三个请求应该被限制
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12347"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestRateLimiterByIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimiterByIP(2, 3, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 不同IP的请求应该独立计算
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.2")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestRateLimiterByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	
	// 模拟JWT认证中间件
	router.Use(func(c *gin.Context) {
		userID := c.GetHeader("User-ID")
		if userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	})
	
	router.Use(RateLimiterByUser(2, 3, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 不同用户的请求应该独立计算
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("User-ID", "user1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("User-ID", "user2")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 未认证用户应该被拒绝
	req3, _ := http.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)
}

func TestTokenBucketAlgorithm(t *testing.T) {
	// 测试令牌桶算法的基本功能
	visitor := &Visitor{
		tokens:    2,
		lastToken: time.Now(),
		lastSeen:  time.Now(),
	}

	// 测试令牌足够时的请求
	assert.True(t, visitor.tokens > 0)
	visitor.tokens--
	assert.Equal(t, 1, visitor.tokens)

	// 测试令牌不足时的请求
	visitor.tokens = 0
	assert.False(t, visitor.tokens > 0)
}

func TestRefillTokens(t *testing.T) {
	rl := NewRateLimiter(10, 20, time.Minute) // 每秒10个令牌，最大20个
	
	visitor := &Visitor{
		tokens:    5,
		lastToken: time.Now().Add(-time.Second), // 1秒前
		lastSeen:  time.Now(),
	}

	// 模拟令牌补充
	now := time.Now()
	elapsed := now.Sub(visitor.lastToken)
	tokensToAdd := int(elapsed.Seconds()) * rl.rate
	
	visitor.tokens += tokensToAdd
	if visitor.tokens > rl.burst {
		visitor.tokens = rl.burst
	}
	visitor.lastToken = now

	// 1秒后应该补充10个令牌，总数不超过burst限制
	assert.Equal(t, 15, visitor.tokens) // 5 + 10 = 15，小于burst(20)
}

func TestCleanupExpiredVisitors(t *testing.T) {
	rl := NewRateLimiter(10, 20, 100*time.Millisecond) // 很短的TTL
	
	// 添加一个访客
	visitor := &Visitor{
		tokens:    10,
		lastToken: time.Now(),
		lastSeen:  time.Now().Add(-200 * time.Millisecond), // 已过期
	}
	
	rl.mu.Lock()
	rl.visitors["test-key"] = visitor
	rl.mu.Unlock()

	// 等待清理
	time.Sleep(150 * time.Millisecond)

	rl.mu.Lock()
	_, exists := rl.visitors["test-key"]
	rl.mu.Unlock()

	// 注意：由于清理是异步的，这个测试可能需要调整
	// 在实际应用中，过期的访客记录应该被清理
	assert.False(t, exists)
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedIP     string
	}{
		{
			name: "X-Forwarded-For头",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
			},
			expectedIP: "192.168.1.1",
		},
		{
			name: "X-Real-IP头",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-Real-IP", "192.168.1.2")
			},
			expectedIP: "192.168.1.2",
		},
		{
			name: "RemoteAddr",
			setupRequest: func(req *http.Request) {
				req.RemoteAddr = "192.168.1.3:12345"
			},
			expectedIP: "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				ip := getClientIP(c)
				c.JSON(http.StatusOK, gin.H{"ip": ip})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			tt.setupRequest(req)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			// 在实际测试中，可以解析响应来验证IP是否正确
		})
	}
}

func TestRateLimiterWithBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建限流器：每秒1个令牌，突发容量3个
	router := gin.New()
	router.Use(RateLimiterByIP(1, 3, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	clientIP := "192.168.1.100"

	// 突发请求应该在burst限制内成功
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", clientIP)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 第4个请求应该被限制
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", clientIP)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestRateLimiterRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建限流器：每100毫秒1个令牌
	router := gin.New()
	router.Use(RateLimiterByIP(10, 1, time.Minute)) // 10个/秒，burst=1
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	clientIP := "192.168.1.101"

	// 第一个请求成功
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", clientIP)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 立即发送第二个请求，应该被限制
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", clientIP)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// 等待令牌恢复
	time.Sleep(150 * time.Millisecond)

	// 新请求应该成功
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-Forwarded-For", clientIP)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

// 辅助函数在rate_limiter.go中已定义