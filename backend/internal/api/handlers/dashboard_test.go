package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"qlib-backend/internal/testutils"
)

func TestDashboardHandlers(t *testing.T) {
	// 设置测试环境
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 设置测试路由器
	router := testutils.SetupTestRouter()
	
	// 添加认证中间件
	router.Use(testutils.MockAuthMiddleware())

	// 添加路由
	router.GET("/dashboard/overview", GetDashboardOverview)
	router.GET("/dashboard/market-overview", GetMarketOverview)
	router.GET("/dashboard/performance-chart", GetPerformanceChart)
	router.GET("/dashboard/recent-tasks", GetRecentTasks)

	// 测试用例
	testCases := []testutils.TestCase{
		{
			Name:           "获取总览统计数据-成功",
			Method:         "GET",
			URL:            "/dashboard/overview",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取市场数据概览-成功",
			Method:         "GET", 
			URL:            "/dashboard/market-overview",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取性能图表数据-成功",
			Method:         "GET",
			URL:            "/dashboard/performance-chart",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取最近任务列表-成功",
			Method:         "GET",
			URL:            "/dashboard/recent-tasks",
			ExpectedStatus: http.StatusOK,
		},
	}

	// 运行测试用例
	testutils.RunTestCases(t, router, testCases)
}

func TestGetDashboardOverview(t *testing.T) {
	// 设置测试环境
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/dashboard/overview", GetDashboardOverview)

	// 创建请求
	req, _ := testutils.CreateJSONRequest("GET", "/dashboard/overview", nil)
	w := testutils.PerformRequest(router, req)

	// 断言响应
	testutils.AssertStatusCode(t, http.StatusOK, w.Code)
	
	// 验证响应结构
	var response map[string]interface{}
	if err := testutils.ParseJSONResponse(w, &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// 验证必要字段存在
	requiredFields := []string{"success", "data"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Response missing required field: %s", field)
		}
	}
}

func TestGetMarketOverview(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/dashboard/market-overview", GetMarketOverview)

	req, _ := testutils.CreateJSONRequest("GET", "/dashboard/market-overview", nil)
	w := testutils.PerformRequest(router, req)

	testutils.AssertStatusCode(t, http.StatusOK, w.Code)
}

func TestGetPerformanceChart(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/dashboard/performance-chart", GetPerformanceChart)

	// 测试不同的时间范围参数
	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认参数", "", http.StatusOK},
		{"7天时间范围", "?period=7d", http.StatusOK},
		{"30天时间范围", "?period=30d", http.StatusOK},
		{"无效时间范围", "?period=invalid", http.StatusOK}, // 应该使用默认值
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/dashboard/performance-chart" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestGetRecentTasks(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/dashboard/recent-tasks", GetRecentTasks)

	// 测试不同的限制参数
	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认限制", "", http.StatusOK},
		{"自定义限制", "?limit=5", http.StatusOK},
		{"最大限制", "?limit=100", http.StatusOK},
		{"无效限制", "?limit=invalid", http.StatusOK}, // 应该使用默认值
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/dashboard/recent-tasks" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}