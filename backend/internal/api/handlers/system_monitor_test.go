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

// MockSystemMonitorService 模拟系统监控服务
type MockSystemMonitorService struct {
	mock.Mock
}

func (m *MockSystemMonitorService) GetRealTimeData(userID uint, metrics []string, interval int, includeHistory bool) (interface{}, error) {
	args := m.Called(userID, metrics, interval, includeHistory)
	return args.Get(0), args.Error(1)
}

// MockNotificationService 模拟通知服务
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) GetNotifications(userID uint, unreadOnly bool, notificationType, priority string, page, pageSize int) (interface{}, error) {
	args := m.Called(userID, unreadOnly, notificationType, priority, page, pageSize)
	return args.Get(0), args.Error(1)
}

func (m *MockNotificationService) MarkAsRead(userID, notificationID uint) error {
	args := m.Called(userID, notificationID)
	return args.Error(0)
}

func TestSystemMonitorHandler_GetRealTimeMonitorData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		userID         uint
		setupMock      func(*MockSystemMonitorService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "成功获取实时监控数据",
			queryParams: "?metrics=cpu,memory&interval=10&include_history=true",
			userID:      1,
			setupMock: func(m *MockSystemMonitorService) {
				m.On("GetRealTimeData", uint(1), []string{"cpu", "memory"}, 10, true).Return(map[string]interface{}{
					"cpu":    85.5,
					"memory": 65.2,
					"timestamp": "2023-01-01T00:00:00Z",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "默认参数获取监控数据",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockSystemMonitorService) {
				m.On("GetRealTimeData", uint(1), []string{"cpu", "memory", "disk", "network", "tasks"}, 5, false).Return(map[string]interface{}{
					"cpu":    80.0,
					"memory": 70.0,
					"disk":   60.0,
					"network": 50.0,
					"tasks":  10,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "服务错误",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockSystemMonitorService) {
				m.On("GetRealTimeData", uint(1), []string{"cpu", "memory", "disk", "network", "tasks"}, 5, false).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockMonitorService := new(MockSystemMonitorService)
			mockNotificationService := new(MockNotificationService)
			tt.setupMock(mockMonitorService)

			// 创建处理器
			handler := NewSystemMonitorHandler(mockMonitorService, mockNotificationService)

			// 创建测试路由
			router := gin.New()
			router.GET("/system/monitor/real-time", testutils.MockAuthMiddleware(tt.userID), handler.GetRealTimeMonitorData)

			// 创建请求
			req, _ := http.NewRequest("GET", "/system/monitor/real-time"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockMonitorService.AssertExpectations(t)
		})
	}
}

func TestSystemMonitorHandler_GetSystemNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		userID         uint
		setupMock      func(*MockNotificationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "成功获取系统通知",
			queryParams: "?unread_only=true&type=warning&priority=high&page=1&page_size=10",
			userID:      1,
			setupMock: func(m *MockNotificationService) {
				m.On("GetNotifications", uint(1), true, "warning", "high", 1, 10).Return([]map[string]interface{}{
					{
						"id":      1,
						"title":   "系统警告",
						"message": "CPU使用率过高",
						"type":    "warning",
						"priority": "high",
						"read":    false,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "默认参数获取通知",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockNotificationService) {
				m.On("GetNotifications", uint(1), false, "", "", 1, 20).Return([]map[string]interface{}{
					{
						"id":      1,
						"title":   "系统信息",
						"message": "任务完成",
						"type":    "info",
						"priority": "normal",
						"read":    true,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "服务错误",
			queryParams: "",
			userID:      1,
			setupMock: func(m *MockNotificationService) {
				m.On("GetNotifications", uint(1), false, "", "", 1, 20).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockMonitorService := new(MockSystemMonitorService)
			mockNotificationService := new(MockNotificationService)
			tt.setupMock(mockNotificationService)

			// 创建处理器
			handler := NewSystemMonitorHandler(mockMonitorService, mockNotificationService)

			// 创建测试路由
			router := gin.New()
			router.GET("/system/notifications", testutils.MockAuthMiddleware(tt.userID), handler.GetSystemNotifications)

			// 创建请求
			req, _ := http.NewRequest("GET", "/system/notifications"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockNotificationService.AssertExpectations(t)
		})
	}
}

func TestSystemMonitorHandler_MarkNotificationAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		notificationID string
		userID         uint
		setupMock      func(*MockNotificationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "成功标记通知已读",
			notificationID: "1",
			userID:         1,
			setupMock: func(m *MockNotificationService) {
				m.On("MarkAsRead", uint(1), uint(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的通知ID",
			notificationID: "invalid",
			userID:         1,
			setupMock:      func(m *MockNotificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的通知ID",
		},
		{
			name:           "服务错误",
			notificationID: "1",
			userID:         1,
			setupMock: func(m *MockNotificationService) {
				m.On("MarkAsRead", uint(1), uint(1)).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockMonitorService := new(MockSystemMonitorService)
			mockNotificationService := new(MockNotificationService)
			tt.setupMock(mockNotificationService)

			// 创建处理器
			handler := NewSystemMonitorHandler(mockMonitorService, mockNotificationService)

			// 创建测试路由
			router := gin.New()
			router.PUT("/system/notifications/:id/read", testutils.MockAuthMiddleware(tt.userID), handler.MarkNotificationAsRead)

			// 创建请求
			req, _ := http.NewRequest("PUT", "/system/notifications/"+tt.notificationID+"/read", nil)
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["message"], tt.expectedError)
			}

			mockNotificationService.AssertExpectations(t)
		})
	}
}