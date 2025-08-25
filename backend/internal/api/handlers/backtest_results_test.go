package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"qlib-backend/internal/services"
	"qlib-backend/internal/testutils"
)

// MockBacktestResultsService 模拟回测结果服务
type MockBacktestResultsService struct {
	mock.Mock
}

func (m *MockBacktestResultsService) GetDetailedResults(resultID, userID uint) (interface{}, error) {
	args := m.Called(resultID, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockBacktestResultsService) GetChartData(resultID uint, chartType services.ChartType, userID uint) (interface{}, error) {
	args := m.Called(resultID, chartType, userID)
	return args.Get(0), args.Error(1)
}

func (m *MockBacktestResultsService) ExportBacktestReport(req services.BacktestReportExportRequest, userID uint) (string, error) {
	args := m.Called(req, userID)
	return args.String(0), args.Error(1)
}

func TestBacktestResultsHandler_GetDetailedResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		resultID       string
		userID         uint
		setupMock      func(*MockBacktestResultsService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "成功获取详细结果",
			resultID: "1",
			userID:   1,
			setupMock: func(m *MockBacktestResultsService) {
				m.On("GetDetailedResults", uint(1), uint(1)).Return(map[string]interface{}{
					"returns":     []float64{0.1, 0.2, 0.15},
					"positions":   []string{"AAPL", "GOOGL"},
					"risk_metrics": map[string]float64{"sharpe": 1.5, "max_drawdown": -0.1},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的结果ID",
			resultID:       "invalid",
			userID:         1,
			setupMock:      func(m *MockBacktestResultsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的结果ID",
		},
		{
			name:     "服务错误",
			resultID: "1",
			userID:   1,
			setupMock: func(m *MockBacktestResultsService) {
				m.On("GetDetailedResults", uint(1), uint(1)).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockService := new(MockBacktestResultsService)
			tt.setupMock(mockService)

			// 创建处理器
			handler := NewBacktestResultsHandler(mockService)

			// 创建测试路由
			router := gin.New()
			router.GET("/backtest/results/:result_id/detailed", testutils.MockAuthMiddleware(tt.userID), handler.GetDetailedResults)

			// 创建请求
			req, _ := http.NewRequest("GET", "/backtest/results/"+tt.resultID+"/detailed", nil)
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

			mockService.AssertExpectations(t)
		})
	}
}

func TestBacktestResultsHandler_GetChartData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		resultID       string
		chartType      string
		userID         uint
		setupMock      func(*MockBacktestResultsService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "成功获取图表数据",
			resultID:  "1",
			chartType: "returns",
			userID:    1,
			setupMock: func(m *MockBacktestResultsService) {
				m.On("GetChartData", uint(1), services.ChartType("returns"), uint(1)).Return(map[string]interface{}{
					"x_data": []string{"2023-01", "2023-02", "2023-03"},
					"y_data": []float64{0.1, 0.15, 0.12},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "空的图表类型",
			resultID:       "1",
			chartType:      "",
			userID:         1,
			setupMock:      func(m *MockBacktestResultsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "图表类型不能为空",
		},
		{
			name:      "服务错误",
			resultID:  "1",
			chartType: "returns",
			userID:    1,
			setupMock: func(m *MockBacktestResultsService) {
				m.On("GetChartData", uint(1), services.ChartType("returns"), uint(1)).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockService := new(MockBacktestResultsService)
			tt.setupMock(mockService)

			// 创建处理器
			handler := NewBacktestResultsHandler(mockService)

			// 创建测试路由
			router := gin.New()
			router.GET("/backtest/charts/:result_id/:chart_type", testutils.MockAuthMiddleware(tt.userID), handler.GetChartData)

			// 创建请求
			req, _ := http.NewRequest("GET", "/backtest/charts/"+tt.resultID+"/"+tt.chartType, nil)
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

			mockService.AssertExpectations(t)
		})
	}
}

func TestBacktestResultsHandler_ExportBacktestReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		userID         uint
		setupMock      func(*MockBacktestResultsService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功导出报告",
			requestBody: map[string]interface{}{
				"result_ids":     []uint{1},
				"report_type":    "detailed",
				"format":         "pdf",
				"template":       "standard",
				"sections":       []string{"performance", "risk"},
				"include_charts": true,
				"benchmark":      "HS300",
				"language":       "zh-CN",
			},
			userID: 1,
			setupMock: func(m *MockBacktestResultsService) {
				m.On("ExportBacktestReport", mock.AnythingOfType("services.BacktestReportExportRequest"), uint(1)).Return("task_123", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的请求体",
			requestBody:    map[string]interface{}{"invalid": "data"},
			userID:         1,
			setupMock:      func(m *MockBacktestResultsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "请求参数无效",
		},
		{
			name: "服务错误",
			requestBody: map[string]interface{}{
				"result_ids":  []uint{1},
				"report_type": "detailed",
				"format":      "pdf",
			},
			userID: 1,
			setupMock: func(m *MockBacktestResultsService) {
				m.On("ExportBacktestReport", mock.AnythingOfType("services.BacktestReportExportRequest"), uint(1)).Return("", assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟服务
			mockService := new(MockBacktestResultsService)
			tt.setupMock(mockService)

			// 创建处理器
			handler := NewBacktestResultsHandler(mockService)

			// 创建测试路由
			router := gin.New()
			router.POST("/backtest/export-report", testutils.MockAuthMiddleware(tt.userID), handler.ExportBacktestReport)

			// 创建请求体
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/backtest/export-report", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
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

			mockService.AssertExpectations(t)
		})
	}
}