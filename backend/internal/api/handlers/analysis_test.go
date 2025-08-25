package handlers

import (
	"net/http"
	"testing"

	"qlib-backend/internal/testutils"
)

func TestAnalysisHandlers(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())

	// 添加分析路由
	router.GET("/analysis/overview", GetAnalysisOverview)
	router.POST("/analysis/models/compare", CompareModelAnalysis)
	router.GET("/analysis/models/:result_id/factor-importance", GetFactorImportance)
	router.GET("/analysis/strategies/:result_id/performance", GetStrategyPerformance)
	router.POST("/analysis/strategies/compare", CompareStrategyAnalysis)
	router.POST("/analysis/reports/generate", GenerateAnalysisReport)
	router.GET("/analysis/reports/:task_id/status", GetReportStatus)
	router.GET("/analysis/results/summary-stats", GetSummaryStats)
	router.POST("/analysis/results/multi-compare", MultiResultCompare)

	testCases := []testutils.TestCase{
		{
			Name:           "获取分析概览",
			Method:         "GET",
			URL:            "/analysis/overview",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "模型性能对比",
			Method: "POST",
			URL:    "/analysis/models/compare",
			Body: map[string]interface{}{
				"model_ids": []int{1, 2, 3},
				"metrics":   []string{"ic", "rank_ic", "sharpe"},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取因子重要性",
			Method:         "GET",
			URL:            "/analysis/models/1/factor-importance",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取策略绩效",
			Method:         "GET",
			URL:            "/analysis/strategies/1/performance",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "策略对比分析",
			Method: "POST",
			URL:    "/analysis/strategies/compare",
			Body: map[string]interface{}{
				"strategy_ids": []int{1, 2},
				"metrics":      []string{"total_return", "sharpe_ratio", "max_drawdown"},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "生成分析报告",
			Method: "POST",
			URL:    "/analysis/reports/generate",
			Body: map[string]interface{}{
				"report_type": "comprehensive",
				"targets": map[string]interface{}{
					"models":     []int{1, 2},
					"strategies": []int{1, 2},
				},
				"format": "pdf",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取报告状态",
			Method:         "GET",
			URL:            "/analysis/reports/task123/status",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取汇总统计",
			Method:         "GET",
			URL:            "/analysis/results/summary-stats",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "多结果对比",
			Method: "POST",
			URL:    "/analysis/results/multi-compare",
			Body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "model", "id": 1},
					{"type": "strategy", "id": 1},
					{"type": "model", "id": 2},
				},
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	testutils.RunTestCases(t, router, testCases)
}

func TestGetAnalysisOverview(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/analysis/overview", GetAnalysisOverview)

	// 测试不同的时间范围
	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认概览", "", http.StatusOK},
		{"最近30天", "?period=30d", http.StatusOK},
		{"最近90天", "?period=90d", http.StatusOK},
		{"自定义时间范围", "?start_date=2023-01-01&end_date=2023-12-31", http.StatusOK},
		{"按类型过滤", "?types=models,strategies", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/analysis/overview" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				// 验证概览数据结构
				requiredFields := []string{"summary", "top_performers", "recent_analysis"}
				for _, field := range requiredFields {
					if _, exists := data[field]; !exists {
						t.Errorf("Overview should contain %s field", field)
					}
				}
			}
		})
	}
}

func TestCompareModelAnalysis(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/analysis/models/compare", CompareModelAnalysis)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "标准模型对比",
			body: map[string]interface{}{
				"model_ids": []int{1, 2, 3},
				"metrics":   []string{"ic", "rank_ic", "sharpe", "mse"},
			},
			status: http.StatusOK,
		},
		{
			name: "包含时间序列分析",
			body: map[string]interface{}{
				"model_ids":           []int{1, 2},
				"metrics":             []string{"ic", "rank_ic"},
				"include_time_series": true,
				"start_date":          "2023-01-01",
				"end_date":            "2023-12-31",
			},
			status: http.StatusOK,
		},
		{
			name: "分组对比",
			body: map[string]interface{}{
				"model_ids": []int{1, 2, 3, 4},
				"metrics":   []string{"ic", "sharpe"},
				"group_by":  "model_type",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少模型ID",
			body: map[string]interface{}{
				"metrics": []string{"ic"},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "空的模型列表",
			body: map[string]interface{}{
				"model_ids": []int{},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的指标",
			body: map[string]interface{}{
				"model_ids": []int{1, 2},
				"metrics":   []string{"invalid_metric"},
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/analysis/models/compare", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				if _, exists := data["comparison"]; !exists {
					t.Error("Response should contain comparison field")
				}
			}
		})
	}
}

func TestGetFactorImportance(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/analysis/models/:result_id/factor-importance", GetFactorImportance)

	testCases := []struct {
		name      string
		resultID  string
		params    string
		status    int
	}{
		{"基本因子重要性", "1", "", http.StatusOK},
		{"前10个重要因子", "1", "?top=10", http.StatusOK},
		{"按重要性排序", "1", "?sort=importance&order=desc", http.StatusOK},
		{"包含相关性分析", "1", "?include_correlation=true", http.StatusOK},
		{"不存在的结果ID", "999999", "", http.StatusNotFound},
		{"无效的结果ID", "invalid", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/analysis/models/" + tc.resultID + "/factor-importance" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				// 验证因子重要性数据结构
				if _, exists := data["factors"]; !exists {
					t.Error("Response should contain factors field")
				}

				if _, exists := data["chart_data"]; !exists {
					t.Error("Response should contain chart_data field")
				}
			}
		})
	}
}

func TestGenerateAnalysisReport(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/analysis/reports/generate", GenerateAnalysisReport)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "综合分析报告",
			body: map[string]interface{}{
				"report_type": "comprehensive",
				"targets": map[string]interface{}{
					"models":     []int{1, 2},
					"strategies": []int{1, 2},
				},
				"format":      "pdf",
				"language":    "zh-CN",
				"include_charts": true,
			},
			status: http.StatusOK,
		},
		{
			name: "模型专项报告",
			body: map[string]interface{}{
				"report_type": "model_analysis",
				"targets": map[string]interface{}{
					"models": []int{1, 2, 3},
				},
				"format": "html",
				"sections": []string{"performance", "factor_analysis", "risk_analysis"},
			},
			status: http.StatusOK,
		},
		{
			name: "策略专项报告",
			body: map[string]interface{}{
				"report_type": "strategy_analysis",
				"targets": map[string]interface{}{
					"strategies": []int{1, 2},
				},
				"format":      "pdf",
				"time_range": map[string]interface{}{
					"start_date": "2023-01-01",
					"end_date":   "2023-12-31",
				},
			},
			status: http.StatusOK,
		},
		{
			name: "自定义报告",
			body: map[string]interface{}{
				"report_type": "custom",
				"config": map[string]interface{}{
					"title":       "Custom Analysis Report",
					"description": "Custom report description",
					"sections": []string{"summary", "detailed_analysis"},
				},
				"targets": map[string]interface{}{
					"models": []int{1},
				},
				"format": "pdf",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少报告类型",
			body: map[string]interface{}{
				"targets": map[string]interface{}{
					"models": []int{1},
				},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "缺少目标",
			body: map[string]interface{}{
				"report_type": "comprehensive",
				"format":      "pdf",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "不支持的格式",
			body: map[string]interface{}{
				"report_type": "comprehensive",
				"targets": map[string]interface{}{
					"models": []int{1},
				},
				"format": "unsupported_format",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/analysis/reports/generate", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				// 验证报告生成响应
				if _, exists := data["task_id"]; !exists {
					t.Error("Response should contain task_id")
				}

				if _, exists := data["status"]; !exists {
					t.Error("Response should contain status")
				}

				if _, exists := data["estimated_time"]; !exists {
					t.Error("Response should contain estimated_time")
				}
			}
		})
	}
}

func TestMultiResultCompare(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/analysis/results/multi-compare", MultiResultCompare)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "混合类型对比",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "model", "id": 1},
					{"type": "strategy", "id": 1},
					{"type": "model", "id": 2},
					{"type": "strategy", "id": 2},
				},
				"metrics": []string{"performance", "risk", "efficiency"},
			},
			status: http.StatusOK,
		},
		{
			name: "仅模型对比",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "model", "id": 1},
					{"type": "model", "id": 2},
					{"type": "model", "id": 3},
				},
				"metrics": []string{"ic", "rank_ic", "sharpe"},
			},
			status: http.StatusOK,
		},
		{
			name: "仅策略对比",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "strategy", "id": 1},
					{"type": "strategy", "id": 2},
				},
				"metrics": []string{"total_return", "sharpe_ratio", "max_drawdown"},
			},
			status: http.StatusOK,
		},
		{
			name: "包含基准对比",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "strategy", "id": 1},
					{"type": "strategy", "id": 2},
				},
				"include_benchmark": true,
				"benchmark": "SH000300",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少目标",
			body: map[string]interface{}{
				"metrics": []string{"performance"},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "空的目标列表",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的目标类型",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "invalid_type", "id": 1},
				},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "目标缺少ID",
			body: map[string]interface{}{
				"targets": []map[string]interface{}{
					{"type": "model"},
				},
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/analysis/results/multi-compare", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				// 验证多结果对比响应
				if _, exists := data["comparison"]; !exists {
					t.Error("Response should contain comparison field")
				}

				if _, exists := data["summary"]; !exists {
					t.Error("Response should contain summary field")
				}
			}
		})
	}
}

func TestGetSummaryStats(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/analysis/results/summary-stats", GetSummaryStats)

	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认统计", "", http.StatusOK},
		{"按类型分组", "?group_by=type", http.StatusOK},
		{"按时间分组", "?group_by=date&period=month", http.StatusOK},
		{"指定时间范围", "?start_date=2023-01-01&end_date=2023-12-31", http.StatusOK},
		{"按用户过滤", "?user_id=1", http.StatusOK},
		{"组合过滤", "?group_by=type&start_date=2023-01-01&user_id=1", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/analysis/results/summary-stats" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				// 验证统计数据结构
				if _, exists := data["total_count"]; !exists {
					t.Error("Response should contain total_count field")
				}

				if _, exists := data["statistics"]; !exists {
					t.Error("Response should contain statistics field")
				}
			}
		})
	}
}