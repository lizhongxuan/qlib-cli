package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		shouldSetUser  bool
	}{
		{
			name:           "有效的JWT Token",
			authHeader:     generateValidToken(t),
			expectedStatus: http.StatusOK,
			shouldSetUser:  true,
		},
		{
			name:           "缺少Authorization头",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			shouldSetUser:  false,
		},
		{
			name:           "Authorization头格式错误 - 缺少Bearer",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			shouldSetUser:  false,
		},
		{
			name:           "Authorization头格式错误 - 只有Bearer",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			shouldSetUser:  false,
		},
		{
			name:           "无效的JWT Token",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			shouldSetUser:  false,
		},
		{
			name:           "过期的JWT Token",
			authHeader:     generateExpiredToken(t),
			expectedStatus: http.StatusUnauthorized,
			shouldSetUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTAuth())
			router.GET("/test", func(c *gin.Context) {
				// 验证用户信息是否正确设置
				if tt.shouldSetUser {
					userID, exists := c.Get("user_id")
					assert.True(t, exists)
					assert.Equal(t, uint(123), userID)

					username, exists := c.Get("username")
					assert.True(t, exists)
					assert.Equal(t, "testuser", username)

					role, exists := c.Get("role")
					assert.True(t, exists)
					assert.Equal(t, "user", role)
				}
				
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGenerateToken(t *testing.T) {
	userID := uint(123)
	username := "testuser"
	role := "admin"

	token, err := GenerateToken(userID, username, role)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证生成的token是否可以被解析
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, role, claims.Role)
}

func TestRefreshToken(t *testing.T) {
	// 生成一个原始token
	originalToken, err := GenerateToken(123, "testuser", "user")
	assert.NoError(t, err)

	// 刷新token
	newToken, err := RefreshToken(originalToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, originalToken, newToken)

	// 验证新token包含相同的用户信息
	parsedToken, err := jwt.ParseWithClaims(newToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)

	// 测试无效token的刷新
	_, err = RefreshToken("invalid.token.here")
	assert.Error(t, err)
}

func TestValidateToken(t *testing.T) {
	// 测试有效token
	validToken, err := GenerateToken(123, "testuser", "user")
	assert.NoError(t, err)

	claims, err := ValidateToken(validToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)

	// 测试无效token
	_, err = ValidateToken("invalid.token.here")
	assert.Error(t, err)

	// 测试过期token
	expiredToken := generateExpiredToken(t)
	expiredTokenStr := expiredToken[7:] // 移除"Bearer "前缀
	_, err = ValidateToken(expiredTokenStr)
	assert.Error(t, err)
}

func TestRoleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       string
		requiredRoles  []string
		expectedStatus int
	}{
		{
			name:           "管理员访问用户权限",
			userRole:       "admin",
			requiredRoles:  []string{"user"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "用户访问管理员权限",
			userRole:       "user",
			requiredRoles:  []string{"admin"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "用户访问用户权限",
			userRole:       "user",
			requiredRoles:  []string{"user"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "多角色权限检查 - 有权限",
			userRole:       "editor",
			requiredRoles:  []string{"admin", "editor", "user"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "多角色权限检查 - 无权限",
			userRole:       "guest",
			requiredRoles:  []string{"admin", "editor"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			// 模拟已通过JWT认证的用户
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(123))
				c.Set("username", "testuser")
				c.Set("role", tt.userRole)
				c.Next()
			})
			
			router.Use(RoleAuth(tt.requiredRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRoleAuthWithoutUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RoleAuth("admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOptionalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		shouldSetUser  bool
	}{
		{
			name:           "有效token的可选认证",
			authHeader:     generateValidToken(t),
			expectedStatus: http.StatusOK,
			shouldSetUser:  true,
		},
		{
			name:           "无token的可选认证",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			shouldSetUser:  false,
		},
		{
			name:           "无效token的可选认证",
			authHeader:     "Bearer invalid.token",
			expectedStatus: http.StatusOK,
			shouldSetUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(OptionalAuth())
			router.GET("/test", func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				if tt.shouldSetUser {
					assert.True(t, exists)
					assert.Equal(t, uint(123), userID)
				} else {
					assert.False(t, exists)
				}
				
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// 辅助函数

func generateValidToken(t *testing.T) string {
	token, err := GenerateToken(123, "testuser", "user")
	assert.NoError(t, err)
	return "Bearer " + token
}

func generateExpiredToken(t *testing.T) string {
	claims := &Claims{
		UserID:   123,
		Username: "testuser",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 过期1小时
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	assert.NoError(t, err)
	
	return "Bearer " + tokenString
}