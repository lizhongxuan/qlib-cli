package handlers

import (
	"net/http"
	"testing"

	"qlib-backend/internal/testutils"
)

func TestStrategyHandlers(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())

	// 添加策略路由
	router.POST("/strategies/backtest", StartBacktest)
	router.GET("/strategies", GetStrategies)
	router.GET("/strategies/:id/results", GetBacktestResults)
	router.GET("/strategies/:id/progress", GetBacktestProgress)
	router.POST("/strategies/:id/stop", StopBacktest)
	router.GET("/strategies/:id/attribution", GetStrategyAttribution)
	router.POST("/strategies/compare", CompareStrategies)
	router.POST("/strategies/:id/optimize", OptimizeStrategy)
	router.POST("/strategies/export", ExportBacktestReport)

	testCases := []testutils.TestCase{
		{
			Name:   "启动策略回测",
			Method: "POST",
			URL:    "/strategies/backtest",
			Body: map[string]interface{}{
				"strategy_name": "TopkDropout Strategy",
				"model_id":      1,
				"start_date":    "2022-01-01",
				"end_date":      "2023-12-31",
				"initial_cash":  1000000,
				"benchmark":     "SH000300",
				"universe":      "csi300",
				"strategy_config": map[string]interface{}{
					"topk":   30,
					"n_drop": 3,
				},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取策略列表",
			Method:         "GET",
			URL:            "/strategies",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取回测结果",
			Method:         "GET",
			URL:            "/strategies/1/results",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取回测进度",
			Method:         "GET",
			URL:            "/strategies/1/progress",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "停止回测",
			Method:         "POST",
			URL:            "/strategies/1/stop",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "策略归因分析",
			Method:         "GET",
			URL:            "/strategies/1/attribution",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "策略对比",
			Method: "POST",
			URL:    "/strategies/compare",
			Body: map[string]interface{}{
				"strategy_ids": []int{1, 2},
				"metrics":      []string{"total_return", "sharpe_ratio", "max_drawdown"},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "参数优化",
			Method: "POST", 
			URL:    "/strategies/1/optimize",
			Body: map[string]interface{}{
				"parameters": map[string]interface{}{
					"topk":   []int{20, 30, 40, 50},
					"n_drop": []int{2, 3, 5},
				},
				"optimization_method": "grid_search",
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	testutils.RunTestCases(t, router, testCases)
}

func TestStartBacktest(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/strategies/backtest", StartBacktest)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "完整的回测配置",
			body: map[string]interface{}{
				"strategy_name": "Complete Strategy Test",
				"model_id":      1,
				"start_date":    "2022-01-01",
				"end_date":      "2023-12-31",
				"initial_cash":  1000000,
				"benchmark":     "SH000300",
				"universe":      "csi300",
				"commission":    0.003,
				"strategy_config": map[string]interface{}{
					"class":  "TopkDropoutStrategy",
					"topk":   30,
					"n_drop": 3,
					"method": "top",
				},
				"exchange_config": map[string]interface{}{
					"limit_threshold": 0.095,
					"deal_price":      "close",
					"open_cost":       0.0005,
					"close_cost":      0.0015,
				},
			},
			status: http.StatusOK,
		},
		{
			name: "最小必要配置",
			body: map[string]interface{}{
				"strategy_name": "Minimal Strategy Test",
				"model_id":      1,
				"start_date":    "2022-01-01",
				"end_date":      "2023-12-31",
				"initial_cash":  1000000,
			},
			status: http.StatusOK,
		},
		{
			name: "使用自定义策略",
			body: map[string]interface{}{
				"strategy_name": "Custom Strategy Test",
				"custom_strategy": map[string]interface{}{
					"class":  "CustomStrategy",
					"config": map[string]interface{}{
						"param1": "value1",
						"param2": 123,
					},
				},
				"start_date":   "2022-01-01",
				"end_date":     "2023-12-31",
				"initial_cash": 1000000,
			},
			status: http.StatusOK,
		},
		{
			name: "缺少策略名称",
			body: map[string]interface{}{
				"model_id":     1,
				"start_date":   "2022-01-01",
				"end_date":     "2023-12-31",
				"initial_cash": 1000000,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的日期格式",
			body: map[string]interface{}{
				"strategy_name": "Invalid Date Test",
				"model_id":      1,
				"start_date":    "invalid-date",
				"end_date":      "2023-12-31",
				"initial_cash":  1000000,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "结束日期早于开始日期",
			body: map[string]interface{}{
				"strategy_name": "Invalid Date Range Test",
				"model_id":      1,
				"start_date":    "2023-12-31",
				"end_date":      "2022-01-01",
				"initial_cash":  1000000,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的初始资金",
			body: map[string]interface{}{
				"strategy_name": "Invalid Cash Test",
				"model_id":      1,
				"start_date":    "2022-01-01",
				"end_date":      "2023-12-31",
				"initial_cash":  -1000000,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "既没有模型ID也没有自定义策略",
			body: map[string]interface{}{
				"strategy_name": "No Strategy Test",
				"start_date":    "2022-01-01",
				"end_date":      "2023-12-31",
				"initial_cash":  1000000,
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/strategies/backtest", tc.body)
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

				// 验证响应包含任务ID
				if _, exists := data["task_id"]; !exists {
					t.Error("Response should contain task_id")
				}

				if _, exists := data["status"]; !exists {
					t.Error("Response should contain status")
				}
			}
		})
	}
}

func TestGetStrategies(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/strategies", GetStrategies)

	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认参数", "", http.StatusOK},
		{"分页参数", "?page=1&limit=10", http.StatusOK},
		{"按状态过滤", "?status=completed", http.StatusOK},
		{"按策略类型过滤", "?strategy_type=TopkDropout", http.StatusOK},
		{"按时间范围过滤", "?start_date=2023-01-01&end_date=2023-12-31", http.StatusOK},
		{"组合过滤", "?status=running&page=1&limit=5", http.StatusOK},
		{"排序", "?sort=created_at&order=desc", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/strategies" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

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

			if _, exists := data["strategies"]; !exists {
				t.Error("Response should contain strategies field")
			}

			if _, exists := data["total"]; !exists {
				t.Error("Response should contain total field")
			}
		})
	}
}

func TestGetBacktestResults(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/strategies/:id/results", GetBacktestResults)

	testCases := []struct {
		name   string
		id     string
		params string
		status int
	}{
		{"基本结果获取", "1", "", http.StatusOK},
		{"详细结果获取", "1", "?detailed=true", http.StatusOK},
		{"包含持仓数据", "1", "?include_positions=true", http.StatusOK},
		{"包含交易记录", "1", "?include_trades=true", http.StatusOK},
		{"完整数据", "1", "?detailed=true&include_positions=true&include_trades=true", http.StatusOK},
		{"不存在的策略", "999999", "", http.StatusNotFound},
		{"无效的策略ID", "invalid", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/strategies/" + tc.id + "/results" + tc.params
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

				// 验证基本性能指标
				requiredFields := []string{"performance", "strategy_name", "start_date", "end_date"}
				for _, field := range requiredFields {
					if _, exists := data[field]; !exists {
						t.Errorf("Results response should contain %s field", field)
					}
				}

				// 验证性能指标
				performance, ok := data["performance"].(map[string]interface{})
				if !ok {
					t.Error("Performance should be an object")
					return
				}

				performanceFields := []string{"total_return", "annual_return", "sharpe_ratio", "max_drawdown"}
				for _, field := range performanceFields {
					if _, exists := performance[field]; !exists {
						t.Errorf("Performance should contain %s field", field)
					}
				}
			}
		})
	}
}

func TestCompareStrategies(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/strategies/compare", CompareStrategies)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "对比两个策略",
			body: map[string]interface{}{
				"strategy_ids": []int{1, 2},
				"metrics":      []string{"total_return", "sharpe_ratio", "max_drawdown"},
			},
			status: http.StatusOK,
		},
		{
			name: "对比多个策略",
			body: map[string]interface{}{
				"strategy_ids": []int{1, 2, 3, 4},
				"metrics":      []string{"total_return", "annual_return", "volatility", "win_rate"},
			},
			status: http.StatusOK,
		},
		{
			name: "默认指标对比",
			body: map[string]interface{}{
				"strategy_ids": []int{1, 2},
			},
			status: http.StatusOK,
		},
		{
			name: "包含基准对比",
			body: map[string]interface{}{
				"strategy_ids":      []int{1, 2},
				"include_benchmark": true,
				"benchmark":         "SH000300",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少策略ID",
			body: map[string]interface{}{
				"metrics": []string{"total_return"},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "空的策略ID列表",
			body: map[string]interface{}{
				"strategy_ids": []int{},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "只有一个策略ID",
			body: map[string]interface{}{
				"strategy_ids": []int{1},
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/strategies/compare", tc.body)
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

func TestOptimizeStrategy(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/strategies/:id/optimize", OptimizeStrategy)

	testCases := []struct {
		name   string
		id     string
		body   map[string]interface{}
		status int
	}{
		{
			name: "网格搜索优化",
			id:   "1",
			body: map[string]interface{}{
				"parameters": map[string]interface{}{
					"topk":   []int{20, 30, 40, 50},
					"n_drop": []int{2, 3, 5},
				},
				"optimization_method": "grid_search",
				"objective":           "sharpe_ratio",
			},
			status: http.StatusOK,
		},
		{
			name: "随机搜索优化",
			id:   "1",
			body: map[string]interface{}{
				"parameters": map[string]interface{}{
					"topk": map[string]interface{}{
						"type":  "range",
						"min":   10,
						"max":   100,
						"step":  5,
					},
					"n_drop": map[string]interface{}{
						"type": "choice",
						"values": []int{1, 2, 3, 5, 8},
					},
				},
				"optimization_method": "random_search",
				"max_trials":          50,
				"objective":           "total_return",
			},
			status: http.StatusOK,
		},
		{
			name: "贝叶斯优化",
			id:   "1",
			body: map[string]interface{}{
				"parameters": map[string]interface{}{
					"topk": map[string]interface{}{
						"type": "range",
						"min":  10,
						"max":  100,
					},
				},
				"optimization_method": "bayesian",
				"max_trials":          30,
			},
			status: http.StatusOK,
		},
		{
			name: "缺少参数",
			id:   "1",
			body: map[string]interface{}{
				"optimization_method": "grid_search",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "不支持的优化方法",
			id:   "1",
			body: map[string]interface{}{
				"parameters": map[string]interface{}{
					"topk": []int{20, 30},
				},
				"optimization_method": "unsupported_method",
			},
			status: http.StatusBadRequest,
		},
		{
			name:   "不存在的策略",
			id:     "999999",
			body:   map[string]interface{}{"parameters": map[string]interface{}{"topk": []int{20, 30}}},
			status: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/strategies/" + tc.id + "/optimize"
			req, _ := testutils.CreateJSONRequest("POST", url, tc.body)
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

				// 验证优化任务响应
				if _, exists := data["optimization_task_id"]; !exists {
					t.Error("Response should contain optimization_task_id")
				}
			}
		})
	}
}