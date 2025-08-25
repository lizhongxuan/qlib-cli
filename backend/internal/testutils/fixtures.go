package testutils

import (
	"time"
	"qlib-backend/internal/models"
)

// CreateTestUser 创建测试用户
func CreateTestUser() *models.User {
	return &models.User{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
	}
}

// CreateTestDataset 创建测试数据集
func CreateTestDataset() *models.Dataset {
	return &models.Dataset{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_dataset",
		Description: "Test dataset description",
		DataPath:    "/test/path/dataset.csv",
		Status:      "active",
		Market:      "csi300",
		StartDate:   "2020-01-01",
		EndDate:     "2023-12-31",
		FileSize:    1024,
		RecordCount: 10000,
		UserID:      1,
	}
}

// CreateTestFactor 创建测试因子
func CreateTestFactor() *models.Factor {
	return &models.Factor{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_factor",
		Description: "Test factor description",
		Expression:  "$close / Ref($close, 1) - 1",
		Category:    "momentum",
		Status:      "active",
		IC:          0.05,
		IR:          0.6,
		RankIC:      0.08,
		UserID:      1,
		IsPublic:    false,
	}
}

// CreateTestModel 创建测试模型
func CreateTestModel() *models.Model {
	return &models.Model{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_model",
		Description: "Test model description",
		Type:        "lgb",
		Status:      "completed",
		Progress:    100,
		ModelPath:   "/test/path/model.pkl",
		ConfigJSON:  `{"num_leaves": 31, "learning_rate": 0.05}`,
		TrainIC:     0.05,
		ValidIC:     0.048,
		TestIC:      0.052,
		TrainStart:  "2020-01-01",
		TrainEnd:    "2022-12-31",
		ValidStart:  "2023-01-01",
		ValidEnd:    "2023-06-30",
		TestStart:   "2023-07-01",
		TestEnd:     "2023-12-31",
		UserID:      1,
	}
}

// CreateTestStrategy 创建测试策略
func CreateTestStrategy() *models.Strategy {
	return &models.Strategy{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_strategy",
		Description: "Test strategy description",
		Type:        "TopkDropoutStrategy",
		Status:      "completed",
		Progress:    100,
		ConfigJSON:  `{"topk": 30, "n_drop": 3}`,
		ModelID:     1,
		BacktestStart: "2022-01-01",
		BacktestEnd:   "2023-12-31",
		TotalReturn:   0.15,
		AnnualReturn:  0.12,
		SharpeRatio:   1.2,
		MaxDrawdown:   -0.08,
		Volatility:    0.18,
		WinRate:       0.55,
		UserID:       1,
	}
}

// CreateTestTask 创建测试任务
func CreateTestTask() *models.Task {
	return &models.Task{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_task",
		Type:        "model_training",
		Status:      "completed",
		Progress:    100,
		Priority:    1,
		Description: "Test model training task",
		ConfigJSON:  `{"model_type": "lgb"}`,
		ResultJSON:  `{"success": true, "model_id": 1}`,
		LogPath:     "/logs/task_1.log",
		ErrorMsg:    "",
		EstimatedTime: 3600,
		UserID:      1,
	}
}

// CreateTestWorkflow 创建测试工作流
func CreateTestWorkflow() *models.Workflow {
	return &models.Workflow{
		ID:         1,
		Name:       "test_workflow",
		TemplateID: 1,
		Status:     "completed",
		Progress:   100,
		ConfigJSON: `{"model": "lgb", "strategy": "TopkDropout"}`,
		ResultJSON: `{"success": true, "model_id": 1, "strategy_id": 1}`,
		UserID:     1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestNotification 创建测试通知
func CreateTestNotification() *models.Notification {
	return &models.Notification{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Type:      "info",
		Category:  "task_completion",
		Title:     "Test Notification",
		Message:   "This is a test notification",
		IsRead:    false,
		ActionURL: "/tasks/1",
		Priority:  "normal",
		UserID:    1,
	}
}

// SeedTestData 填充测试数据
func SeedTestData(db interface{}) error {
	// 这里可以添加批量数据创建逻辑
	return nil
}

// MockAPIResponse 模拟API响应
type MockAPIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CreateSuccessResponse 创建成功响应
func CreateSuccessResponse(data interface{}) MockAPIResponse {
	return MockAPIResponse{
		Success: true,
		Data:    data,
		Message: "操作成功",
	}
}

// CreateErrorResponse 创建错误响应
func CreateErrorResponse(error string) MockAPIResponse {
	return MockAPIResponse{
		Success: false,
		Error:   error,
	}
}

// MockFactorData 模拟因子数据
type MockFactorData struct {
	Instrument string    `json:"instrument"`
	Date       time.Time `json:"date"`
	Value      float64   `json:"value"`
	IsValid    bool      `json:"is_valid"`
}

// CreateMockFactorData 创建模拟因子数据
func CreateMockFactorData() []MockFactorData {
	return []MockFactorData{
		{
			Instrument: "000001.SZ",
			Date:       time.Now().AddDate(0, 0, -1),
			Value:      0.02,
			IsValid:    true,
		},
		{
			Instrument: "000002.SZ", 
			Date:       time.Now().AddDate(0, 0, -1),
			Value:      -0.01,
			IsValid:    true,
		},
	}
}

// MockModelMetrics 模拟模型指标
type MockModelMetrics struct {
	IC     float64 `json:"ic"`
	RankIC float64 `json:"rank_ic"`
	ICIR   float64 `json:"icir"`
	MSE    float64 `json:"mse"`
}

// CreateMockModelMetrics 创建模拟模型指标
func CreateMockModelMetrics() MockModelMetrics {
	return MockModelMetrics{
		IC:     0.05,
		RankIC: 0.08,
		ICIR:   0.6,
		MSE:    0.001,
	}
}

// MockBacktestResult 模拟回测结果
type MockBacktestResult struct {
	TotalReturn  float64 `json:"total_return"`
	AnnualReturn float64 `json:"annual_return"`
	SharpeRatio  float64 `json:"sharpe_ratio"`
	MaxDrawdown  float64 `json:"max_drawdown"`
	WinRate      float64 `json:"win_rate"`
}

// CreateMockBacktestResult 创建模拟回测结果
func CreateMockBacktestResult() MockBacktestResult {
	return MockBacktestResult{
		TotalReturn:  0.15,
		AnnualReturn: 0.12,
		SharpeRatio:  1.2,
		MaxDrawdown:  -0.08,
		WinRate:      0.55,
	}
}