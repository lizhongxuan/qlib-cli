package handlers

import (
	"net/http"
	"testing"

	"qlib-backend/internal/testutils"
)

func TestFactorHandlers(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())

	// 添加因子管理路由
	router.GET("/factors", GetFactors)
	router.POST("/factors", CreateFactor)
	router.PUT("/factors/:id", UpdateFactor)
	router.DELETE("/factors/:id", DeleteFactor)
	router.POST("/factors/test", TestFactor)
	router.GET("/factors/:id/analysis", GetFactorAnalysis)
	router.POST("/factors/batch-test", BatchTestFactors)
	router.GET("/factors/categories", GetFactorCategories)
	router.POST("/factors/import", ImportFactorLibrary)
	
	// 添加因子研究工作台路由
	router.POST("/factors/ai-chat", FactorAIChat)
	router.POST("/factors/validate-syntax", ValidateFactorSyntax)
	router.GET("/factors/qlib-functions", GetQlibFunctions)
	router.GET("/factors/syntax-reference", GetSyntaxReference)
	router.POST("/factors/save-workspace", SaveFactorWorkspace)

	testCases := []testutils.TestCase{
		{
			Name:           "获取因子列表",
			Method:         "GET",
			URL:            "/factors",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "创建新因子",
			Method: "POST",
			URL:    "/factors",
			Body: map[string]interface{}{
				"name":        "momentum_5d",
				"description": "5日动量因子",
				"expression":  "$close / Ref($close, 5) - 1",
				"category":    "momentum",
				"universe":    "csi300",
				"frequency":   "day",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "测试因子性能",
			Method: "POST",
			URL:    "/factors/test",
			Body: map[string]interface{}{
				"factor_id":   1,
				"start_date":  "2022-01-01",
				"end_date":    "2023-12-31",
				"universe":    "csi300",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取因子分类",
			Method:         "GET",
			URL:            "/factors/categories",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "AI因子研究助手",
			Method: "POST",
			URL:    "/factors/ai-chat",
			Body: map[string]interface{}{
				"message": "帮我构建一个价格动量因子",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "验证因子语法",
			Method: "POST",
			URL:    "/factors/validate-syntax",
			Body: map[string]interface{}{
				"expression": "$close / Ref($close, 1) - 1",
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	testutils.RunTestCases(t, router, testCases)
}

func TestCreateFactor(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/factors", CreateFactor)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "完整因子信息",
			body: map[string]interface{}{
				"name":        "complete_factor",
				"description": "完整的因子信息",
				"expression":  "($high + $low + $close) / 3",
				"category":    "technical",
				"universe":    "csi300",
				"frequency":   "day",
			},
			status: http.StatusOK,
		},
		{
			name: "最小必要信息",
			body: map[string]interface{}{
				"name":       "minimal_factor",
				"expression": "$close",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少名称",
			body: map[string]interface{}{
				"expression": "$close",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "缺少表达式",
			body: map[string]interface{}{
				"name": "no_expression_factor",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的表达式语法",
			body: map[string]interface{}{
				"name":       "invalid_syntax_factor",
				"expression": "invalid syntax here",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/factors", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestTestFactor(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/factors/test", TestFactor)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "完整测试参数",
			body: map[string]interface{}{
				"factor_id":  1,
				"start_date": "2022-01-01",
				"end_date":   "2023-12-31",
				"universe":   "csi300",
				"frequency":  "day",
			},
			status: http.StatusOK,
		},
		{
			name: "使用因子表达式测试",
			body: map[string]interface{}{
				"expression": "$close / Ref($close, 1) - 1",
				"start_date": "2022-01-01",
				"end_date":   "2023-12-31",
				"universe":   "csi300",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少因子信息",
			body: map[string]interface{}{
				"start_date": "2022-01-01",
				"end_date":   "2023-12-31",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的日期格式",
			body: map[string]interface{}{
				"factor_id":  1,
				"start_date": "invalid-date",
				"end_date":   "2023-12-31",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "结束日期早于开始日期",
			body: map[string]interface{}{
				"factor_id":  1,
				"start_date": "2023-12-31",
				"end_date":   "2022-01-01",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/factors/test", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestBatchTestFactors(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/factors/batch-test", BatchTestFactors)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "批量测试多个因子",
			body: map[string]interface{}{
				"factor_ids": []int{1, 2, 3},
				"start_date": "2022-01-01",
				"end_date":   "2023-12-31",
				"universe":   "csi300",
			},
			status: http.StatusOK,
		},
		{
			name: "批量测试表达式",
			body: map[string]interface{}{
				"expressions": []string{
					"$close / Ref($close, 1) - 1",
					"$close / Ref($close, 5) - 1",
					"Mean($close, 10)",
				},
				"start_date": "2022-01-01",
				"end_date":   "2023-12-31",
				"universe":   "csi300",
			},
			status: http.StatusOK,
		},
		{
			name: "空的因子列表",
			body: map[string]interface{}{
				"factor_ids": []int{},
				"start_date": "2022-01-01",
				"end_date":   "2023-12-31",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/factors/batch-test", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestFactorAIChat(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/factors/ai-chat", FactorAIChat)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "因子构建咨询",
			body: map[string]interface{}{
				"message": "如何构建一个价格动量因子？",
			},
			status: http.StatusOK,
		},
		{
			name: "因子优化建议",
			body: map[string]interface{}{
				"message": "这个因子表达式有什么问题：$close / $open",
				"context": map[string]interface{}{
					"current_expression": "$close / $open",
				},
			},
			status: http.StatusOK,
		},
		{
			name: "空消息",
			body: map[string]interface{}{
				"message": "",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "消息过长",
			body: map[string]interface{}{
				"message": string(make([]byte, 10000)), // 生成超长消息
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/factors/ai-chat", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}

func TestValidateFactorSyntax(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/factors/validate-syntax", ValidateFactorSyntax)

	testCases := []struct {
		name       string
		body       map[string]interface{}
		status     int
		shouldPass bool
	}{
		{
			name: "有效的基本表达式",
			body: map[string]interface{}{
				"expression": "$close",
			},
			status:     http.StatusOK,
			shouldPass: true,
		},
		{
			name: "有效的复杂表达式",
			body: map[string]interface{}{
				"expression": "($close - Ref($close, 1)) / Ref($close, 1)",
			},
			status:     http.StatusOK,
			shouldPass: true,
		},
		{
			name: "有效的函数调用",
			body: map[string]interface{}{
				"expression": "Mean($close, 5)",
			},
			status:     http.StatusOK,
			shouldPass: true,
		},
		{
			name: "无效的语法",
			body: map[string]interface{}{
				"expression": "invalid syntax here",
			},
			status:     http.StatusOK,
			shouldPass: false,
		},
		{
			name: "缺少表达式",
			body: map[string]interface{}{},
			status: http.StatusBadRequest,
		},
		{
			name: "空表达式",
			body: map[string]interface{}{
				"expression": "",
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/factors/validate-syntax", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				isValid, ok := data["valid"].(bool)
				if !ok {
					t.Error("Response should contain valid field")
					return
				}

				if isValid != tc.shouldPass {
					t.Errorf("Expected validation result %v, got %v", tc.shouldPass, isValid)
				}
			}
		})
	}
}

func TestGetFactorCategories(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/factors/categories", GetFactorCategories)

	req, _ := testutils.CreateJSONRequest("GET", "/factors/categories", nil)
	w := testutils.PerformRequest(router, req)

	testutils.AssertStatusCode(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	if err := testutils.ParseJSONResponse(w, &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	// 验证响应包含分类数据
	data, ok := response["data"]
	if !ok {
		t.Error("Response should contain data field")
	}

	categories, ok := data.(map[string]interface{})
	if !ok {
		t.Error("Categories data should be an object")
		return
	}

	// 验证包含基本分类
	expectedCategories := []string{"price", "technical", "momentum", "volatility", "volume"}
	for _, category := range expectedCategories {
		if _, exists := categories[category]; !exists {
			t.Errorf("Expected category %s not found in response", category)
		}
	}
}

func TestGetQlibFunctions(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/factors/qlib-functions", GetQlibFunctions)

	req, _ := testutils.CreateJSONRequest("GET", "/factors/qlib-functions", nil)
	w := testutils.PerformRequest(router, req)

	testutils.AssertStatusCode(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	if err := testutils.ParseJSONResponse(w, &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return
	}

	// 验证响应包含函数列表
	data, ok := response["data"]
	if !ok {
		t.Error("Response should contain data field")
	}

	functions, ok := data.([]interface{})
	if !ok {
		t.Error("Functions data should be an array")
		return
	}

	// 验证包含基本函数
	if len(functions) == 0 {
		t.Error("Functions list should not be empty")
	}
}