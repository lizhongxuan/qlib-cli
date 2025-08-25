package handlers

import (
	"net/http"
	"testing"

	"qlib-backend/internal/testutils"
)

func TestModelHandlers(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())

	// 添加模型路由
	router.POST("/models/train", TrainModel)
	router.GET("/models", GetModels)
	router.GET("/models/:id/progress", GetTrainingProgress)
	router.POST("/models/:id/stop", StopTraining)
	router.GET("/models/:id/evaluate", EvaluateModel)
	router.POST("/models/compare", CompareModels)
	router.POST("/models/:id/deploy", DeployModel)
	router.GET("/models/:id/logs", GetTrainingLogs)

	testCases := []testutils.TestCase{
		{
			Name:   "启动模型训练",
			Method: "POST",
			URL:    "/models/train",
			Body: map[string]interface{}{
				"name":       "test_model",
				"model_type": "lgb",
				"dataset_id": 1,
				"parameters": map[string]interface{}{
					"num_leaves":     31,
					"learning_rate":  0.05,
					"feature_fraction": 0.9,
				},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取模型列表",
			Method:         "GET",
			URL:            "/models",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取训练进度",
			Method:         "GET", 
			URL:            "/models/1/progress",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "停止训练",
			Method:         "POST",
			URL:            "/models/1/stop",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "模型评估",
			Method:         "GET",
			URL:            "/models/1/evaluate",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "模型对比",
			Method: "POST",
			URL:    "/models/compare",
			Body: map[string]interface{}{
				"model_ids": []int{1, 2},
				"metrics":   []string{"ic", "rank_ic", "sharpe"},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "部署模型",
			Method:         "POST",
			URL:            "/models/1/deploy",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "获取训练日志",
			Method:         "GET",
			URL:            "/models/1/logs",
			ExpectedStatus: http.StatusOK,
		},
	}

	testutils.RunTestCases(t, router, testCases)
}

func TestTrainModel(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/models/train", TrainModel)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "LightGBM模型训练",
			body: map[string]interface{}{
				"name":       "lgb_test_model",
				"model_type": "lgb",
				"dataset_id": 1,
				"parameters": map[string]interface{}{
					"num_leaves":       31,
					"learning_rate":    0.05,
					"feature_fraction": 0.9,
					"bagging_fraction": 0.8,
					"bagging_freq":     5,
				},
				"dataset_config": map[string]interface{}{
					"train_start": "2020-01-01",
					"train_end":   "2022-12-31", 
					"valid_start": "2023-01-01",
					"valid_end":   "2023-06-30",
					"test_start":  "2023-07-01",
					"test_end":    "2023-12-31",
				},
			},
			status: http.StatusOK,
		},
		{
			name: "XGBoost模型训练",
			body: map[string]interface{}{
				"name":       "xgb_test_model",
				"model_type": "xgb",
				"dataset_id": 1,
				"parameters": map[string]interface{}{
					"max_depth":        6,
					"learning_rate":    0.1,
					"n_estimators":     100,
					"subsample":        0.8,
					"colsample_bytree": 0.8,
				},
			},
			status: http.StatusOK,
		},
		{
			name: "线性模型训练",
			body: map[string]interface{}{
				"name":       "linear_test_model",
				"model_type": "linear",
				"dataset_id": 1,
				"parameters": map[string]interface{}{
					"estimator": "ridge",
					"alpha":     1.0,
				},
			},
			status: http.StatusOK,
		},
		{
			name: "缺少模型名称",
			body: map[string]interface{}{
				"model_type": "lgb",
				"dataset_id": 1,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "不支持的模型类型",
			body: map[string]interface{}{
				"name":       "unsupported_model",
				"model_type": "unsupported",
				"dataset_id": 1,
			},
			status: http.StatusBadRequest,
		},
		{
			name: "缺少数据集ID",
			body: map[string]interface{}{
				"name":       "no_dataset_model",
				"model_type": "lgb",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的参数",
			body: map[string]interface{}{
				"name":       "invalid_params_model",
				"model_type": "lgb",
				"dataset_id": 1,
				"parameters": map[string]interface{}{
					"learning_rate": "invalid", // 应该是数字
				},
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/models/train", tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			if tc.status == http.StatusOK {
				var response map[string]interface{}
				if err := testutils.ParseJSONResponse(w, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
					return
				}

				// 验证响应包含任务ID
				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Response data should be an object")
					return
				}

				if _, exists := data["task_id"]; !exists {
					t.Error("Response should contain task_id")
				}
			}
		})
	}
}

func TestGetModels(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/models", GetModels)

	// 测试分页和过滤参数
	testCases := []struct {
		name   string
		params string
		status int
	}{
		{"默认参数", "", http.StatusOK},
		{"分页参数", "?page=1&limit=10", http.StatusOK},
		{"按状态过滤", "?status=trained", http.StatusOK},
		{"按类型过滤", "?model_type=lgb", http.StatusOK},
		{"组合过滤", "?status=training&model_type=xgb&page=1", http.StatusOK},
		{"排序", "?sort=created_at&order=desc", http.StatusOK},
		{"无效分页", "?page=0&limit=-1", http.StatusOK}, // 应该使用默认值
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/models" + tc.params
			req, _ := testutils.CreateJSONRequest("GET", url, nil)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)

			var response map[string]interface{}
			if err := testutils.ParseJSONResponse(w, &response); err != nil {
				t.Errorf("Failed to parse response: %v", err)
				return
			}

			// 验证响应结构
			data, ok := response["data"].(map[string]interface{})
			if !ok {
				t.Error("Response data should be an object")
				return
			}

			if _, exists := data["models"]; !exists {
				t.Error("Response should contain models field")
			}

			if _, exists := data["total"]; !exists {
				t.Error("Response should contain total field")
			}
		})
	}
}

func TestGetTrainingProgress(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/models/:id/progress", GetTrainingProgress)

	testCases := []struct {
		name   string
		id     string
		status int
	}{
		{"获取存在模型的进度", "1", http.StatusOK},
		{"获取不存在模型的进度", "999999", http.StatusNotFound},
		{"无效的模型ID", "invalid", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/models/" + tc.id + "/progress"
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

				// 验证进度字段
				requiredFields := []string{"progress", "status", "current_epoch", "total_epochs"}
				for _, field := range requiredFields {
					if _, exists := data[field]; !exists {
						t.Errorf("Progress response should contain %s field", field)
					}
				}
			}
		})
	}
}

func TestCompareModels(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/models/compare", CompareModels)

	testCases := []struct {
		name   string
		body   map[string]interface{}
		status int
	}{
		{
			name: "对比两个模型",
			body: map[string]interface{}{
				"model_ids": []int{1, 2},
				"metrics":   []string{"ic", "rank_ic", "sharpe_ratio"},
			},
			status: http.StatusOK,
		},
		{
			name: "对比多个模型",
			body: map[string]interface{}{
				"model_ids": []int{1, 2, 3, 4},
				"metrics":   []string{"ic", "rank_ic", "mse", "mae"},
			},
			status: http.StatusOK,
		},
		{
			name: "默认指标对比",
			body: map[string]interface{}{
				"model_ids": []int{1, 2},
			},
			status: http.StatusOK,
		},
		{
			name: "缺少模型ID",
			body: map[string]interface{}{
				"metrics": []string{"ic", "rank_ic"},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "空的模型ID列表",
			body: map[string]interface{}{
				"model_ids": []int{},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "只有一个模型ID",
			body: map[string]interface{}{
				"model_ids": []int{1},
			},
			status: http.StatusBadRequest,
		},
		{
			name: "无效的指标名称",
			body: map[string]interface{}{
				"model_ids": []int{1, 2},
				"metrics":   []string{"invalid_metric"},
			},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := testutils.CreateJSONRequest("POST", "/models/compare", tc.body)
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

				// 验证对比结果结构
				if _, exists := data["comparison"]; !exists {
					t.Error("Response should contain comparison field")
				}
			}
		})
	}
}

func TestEvaluateModel(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.GET("/models/:id/evaluate", EvaluateModel)

	testCases := []struct {
		name   string
		id     string
		params string
		status int
	}{
		{"评估存在的模型", "1", "", http.StatusOK},
		{"带测试数据集的评估", "1", "?test_dataset_id=2", http.StatusOK},
		{"自定义时间范围评估", "1", "?start_date=2023-01-01&end_date=2023-12-31", http.StatusOK},
		{"评估不存在的模型", "999999", "", http.StatusNotFound},
		{"无效的模型ID", "invalid", "", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/models/" + tc.id + "/evaluate" + tc.params
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

				// 验证评估结果结构
				requiredFields := []string{"metrics", "ic_analysis", "predictions"}
				for _, field := range requiredFields {
					if _, exists := data[field]; !exists {
						t.Errorf("Evaluation response should contain %s field", field)
					}
				}
			}
		})
	}
}

func TestDeployModel(t *testing.T) {
	router := testutils.SetupTestRouter()
	router.Use(testutils.MockAuthMiddleware())
	router.POST("/models/:id/deploy", DeployModel)

	testCases := []struct {
		name   string
		id     string
		body   map[string]interface{}
		status int
	}{
		{
			name: "部署到生产环境",
			id:   "1",
			body: map[string]interface{}{
				"environment": "production",
				"config": map[string]interface{}{
					"cpu_limit":    "2",
					"memory_limit": "4Gi",
					"replicas":     2,
				},
			},
			status: http.StatusOK,
		},
		{
			name: "部署到测试环境",
			id:   "1",
			body: map[string]interface{}{
				"environment": "test",
				"config": map[string]interface{}{
					"cpu_limit":    "1",
					"memory_limit": "2Gi",
					"replicas":     1,
				},
			},
			status: http.StatusOK,
		},
		{
			name: "默认配置部署",
			id:   "1",
			body: map[string]interface{}{
				"environment": "test",
			},
			status: http.StatusOK,
		},
		{
			name: "缺少环境参数",
			id:   "1",
			body: map[string]interface{}{},
			status: http.StatusBadRequest,
		},
		{
			name:   "部署不存在的模型",
			id:     "999999",
			body:   map[string]interface{}{"environment": "test"},
			status: http.StatusNotFound,
		},
		{
			name:   "无效的模型ID",
			id:     "invalid",
			body:   map[string]interface{}{"environment": "test"},
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/models/" + tc.id + "/deploy"
			req, _ := testutils.CreateJSONRequest("POST", url, tc.body)
			w := testutils.PerformRequest(router, req)
			testutils.AssertStatusCode(t, tc.status, w.Code)
		})
	}
}