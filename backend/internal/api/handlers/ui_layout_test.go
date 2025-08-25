package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"qlib-backend/internal/testutils"
)

// MockUIConfigService 模拟UI配置服务
type MockUIConfigService struct {
	mock.Mock
}

func (m *MockUIConfigService) GetLayoutConfig(userID uint, configType, platform, theme string) (interface{}, error) {
	args := m.Called(userID, configType, platform, theme)
	return args.Get(0), args.Error(1)
}

func TestUILayoutHandler_GetLayoutConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		userID         uint
		setupMock      func(*MockUIConfigService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "成功获取默认布局配置",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockUIConfigService) {
				m.On("GetLayoutConfig", uint(1), "default", "web", "light").Return(map[string]interface{}{
					"layout": "grid",
					"sidebar": map[string]interface{}{
						"collapsed": false,
						"width":     280,
					},
					"header": map[string]interface{}{
						"height": 64,
						"fixed":  true,
					},
					"theme": map[string]interface{}{
						"primary_color":   "#1890ff",
						"background_color": "#ffffff",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "成功获取自定义布局配置",
			queryParams: "?type=dashboard&platform=mobile&theme=dark",
			userID:      1,
			setupMock: func(m *MockUIConfigService) {
				m.On("GetLayoutConfig", uint(1), "dashboard", "mobile", "dark").Return(map[string]interface{}{
					"layout": "flex",
					"sidebar": map[string]interface{}{
						"collapsed": true,
						"width":     240,
					},
					"header": map[string]interface{}{
						"height": 56,
						"fixed":  false,
					},
					"theme": map[string]interface{}{
						"primary_color":   "#722ed1",
						"background_color": "#141414",
					},
					"responsive": map[string]interface{}{
						"breakpoints": map[string]int{
							"xs": 480,
							"sm": 576,
							"md": 768,
							"lg": 992,
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "平板端布局配置",
			queryParams: "?platform=tablet&theme=light",
			userID:      1,
			setupMock: func(m *MockUIConfigService) {
				m.On("GetLayoutConfig", uint(1), "default", "tablet", "light").Return(map[string]interface{}{
					"layout": "adaptive",
					"sidebar": map[string]interface{}{
						"collapsed": false,
						"width":     260,
						"mode":      "drawer",
					},
					"header": map[string]interface{}{
						"height": 60,
						"fixed":  true,
					},
					"grid": map[string]interface{}{
						"columns": 12,
						"gutter":  16,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "服务错误",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockUIConfigService) {
				m.On("GetLayoutConfig", uint(1), "default", "web", "light").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockService := new(MockUIConfigService)
			tt.setupMock(mockService)

			// 创建处理器
			handler := NewUILayoutHandler(mockService)

			// 创建测试路由
			router := gin.New()
			router.GET("/ui/layout/config", testutils.MockAuthMiddleware(tt.userID), handler.GetLayoutConfig)

			// 创建请求
			req, _ := http.NewRequest("GET", "/ui/layout/config"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, true, response["success"])
				assert.NotNil(t, response["data"])
			}
			
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUILayoutHandler_GetLayoutConfig_UnauthorizedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建处理器
	mockService := new(MockUIConfigService)
	handler := NewUILayoutHandler(mockService)

	// 创建测试路由（不使用认证中间件）
	router := gin.New()
	router.GET("/ui/layout/config", handler.GetLayoutConfig)

	// 创建请求
	req, _ := http.NewRequest("GET", "/ui/layout/config", nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["message"], "用户未认证")
}