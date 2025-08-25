package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"qlib-backend/internal/models"
	"qlib-backend/internal/qlib"
	"qlib-backend/internal/testutils"
)

// MockBacktestEngine 模拟回测引擎
type MockBacktestEngine struct {
	mock.Mock
}

func (m *MockBacktestEngine) RunBacktest(config qlib.BacktestConfig) (*qlib.BacktestResult, error) {
	args := m.Called(config)
	return args.Get(0).(*qlib.BacktestResult), args.Error(1)
}

func (m *MockBacktestEngine) GetBacktestProgress(taskID uint) (*qlib.BacktestProgress, error) {
	args := m.Called(taskID)
	return args.Get(0).(*qlib.BacktestProgress), args.Error(1)
}

func (m *MockBacktestEngine) StopBacktest(taskID uint) error {
	args := m.Called(taskID)
	return args.Error(0)
}

// MockTaskService 模拟任务服务
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateBacktestTask(name string, userID uint, config interface{}) (uint, error) {
	args := m.Called(name, userID, config)
	return uint(args.Int(0)), args.Error(1)
}

func (m *MockTaskService) UpdateTaskProgress(taskID uint, progress int, status string) error {
	args := m.Called(taskID, progress, status)
	return args.Error(0)
}

type StrategyServiceTestSuite struct {
	suite.Suite
	service         *StrategyService
	testDB          *testutils.TestDB
	mockEngine      *MockBacktestEngine
	mockTaskService *MockTaskService
}

func (suite *StrategyServiceTestSuite) SetupSuite() {
	suite.testDB = testutils.SetupTestDB()
	suite.mockEngine = new(MockBacktestEngine)
	suite.mockTaskService = new(MockTaskService)
	suite.service = NewStrategyService(suite.testDB.DB, suite.mockEngine, suite.mockTaskService)
}

func (suite *StrategyServiceTestSuite) TearDownSuite() {
	suite.testDB.Cleanup()
}

func (suite *StrategyServiceTestSuite) SetupTest() {
	suite.testDB.CleanupTables()
	suite.mockEngine.ExpectedCalls = nil
	suite.mockTaskService.ExpectedCalls = nil
}

func (suite *StrategyServiceTestSuite) TestStartBacktest() {
	userID := uint(1)
	taskID := uint(123)

	// 创建测试模型
	model := models.Model{
		Name:   "测试模型",
		Type:   "lightgbm",
		Status: "completed",
		UserID: userID,
		TestIC: 0.045,
	}
	suite.testDB.DB.Create(&model)

	req := StrategyBacktestRequest{
		Name:        "测试策略回测",
		Type:        "top_k",
		ModelID:     model.ID,
		StartDate:   "2023-01-01",
		EndDate:     "2023-12-31",
		Benchmark:   "HS300",
		TopK:        50,
		Rebalance:   "M", // 月度调仓
		Commission:  0.003,
		Parameters: map[string]interface{}{
			"universe": "CSI300",
			"factors":  []string{"momentum", "value"},
		},
	}

	// 设置模拟期望
	suite.mockTaskService.On("CreateBacktestTask", req.Name, userID, mock.Anything).Return(int(taskID), nil)
	suite.mockEngine.On("RunBacktest", mock.Anything).Return(&qlib.BacktestResult{
		TaskID:      taskID,
		Status:      "running",
		Progress:    0,
		TotalReturn: 0,
	}, nil)

	response, err := suite.service.StartBacktest(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), taskID, response.TaskID)
	assert.Equal(suite.T(), "submitted", response.Status)
	assert.Equal(suite.T(), req.Name, response.StrategyName)
	assert.Greater(suite.T(), response.StrategyID, uint(0))

	// 验证策略记录已创建
	var strategy models.Strategy
	err = suite.testDB.DB.First(&strategy, response.StrategyID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), req.Name, strategy.Name)
	assert.Equal(suite.T(), req.Type, strategy.Type)
	assert.Equal(suite.T(), userID, strategy.UserID)

	suite.mockTaskService.AssertExpectations(suite.T())
	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *StrategyServiceTestSuite) TestStartBacktestWithInvalidModel() {
	userID := uint(1)

	req := StrategyBacktestRequest{
		Name:      "测试策略回测",
		Type:      "top_k",
		ModelID:   999, // 不存在的模型ID
		StartDate: "2023-01-01",
		EndDate:   "2023-12-31",
	}

	response, err := suite.service.StartBacktest(req, userID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "指定的模型不存在或无权限访问")
}

func (suite *StrategyServiceTestSuite) TestGetStrategies() {
	userID := uint(1)

	// 创建测试策略
	strategies := []models.Strategy{
		{
			Name:           "策略A",
			Type:           "top_k",
			Status:         "completed",
			UserID:         userID,
			TotalReturn:    0.156,
			SharpeRatio:    1.45,
			MaxDrawdown:    -0.08,
			BacktestAt:     time.Now().AddDate(0, 0, -1),
		},
		{
			Name:           "策略B",
			Type:           "long_short",
			Status:         "running",
			UserID:         userID,
			TotalReturn:    0,
			SharpeRatio:    0,
			MaxDrawdown:    0,
			BacktestAt:     time.Now(),
		},
	}

	for i := range strategies {
		suite.testDB.DB.Create(&strategies[i])
	}

	// 测试获取所有策略
	result, err := suite.service.GetStrategies(userID, 1, 10, "", "")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(2), result.Total)
	assert.Len(suite.T(), result.Strategies, 2)

	// 测试按状态筛选
	result, err = suite.service.GetStrategies(userID, 1, 10, "completed", "")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Strategies, 1)
	assert.Equal(suite.T(), "策略A", result.Strategies[0].Name)

	// 测试按类型筛选
	result, err = suite.service.GetStrategies(userID, 1, 10, "", "top_k")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Strategies, 1)
	assert.Equal(suite.T(), "策略A", result.Strategies[0].Name)
}

func (suite *StrategyServiceTestSuite) TestGetBacktestResults() {
	userID := uint(1)

	// 创建测试策略
	strategy := models.Strategy{
		Name:           "测试策略",
		Type:           "top_k",
		Status:         "completed",
		UserID:         userID,
		TotalReturn:    0.156,
		AnnualReturn:   0.123,
		SharpeRatio:    1.45,
		MaxDrawdown:    -0.08,
		VolatilityAnnual: 0.15,
		BacktestAt:     time.Now(),
	}
	suite.testDB.DB.Create(&strategy)

	results, err := suite.service.GetBacktestResults(strategy.ID, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), results)
	assert.Equal(suite.T(), strategy.ID, results.StrategyID)
	assert.Equal(suite.T(), strategy.Name, results.StrategyName)
	assert.Equal(suite.T(), strategy.TotalReturn, results.TotalReturn)
	assert.Equal(suite.T(), strategy.SharpeRatio, results.SharpeRatio)
	assert.NotNil(suite.T(), results.PerformanceMetrics)
	assert.NotNil(suite.T(), results.RiskMetrics)
}

func (suite *StrategyServiceTestSuite) TestGetBacktestProgress() {
	taskID := uint(123)

	suite.mockEngine.On("GetBacktestProgress", taskID).Return(&qlib.BacktestProgress{
		TaskID:      taskID,
		Status:      "running",
		Progress:    65,
		CurrentStep: "计算组合权重中...",
		ETA:         time.Now().Add(5 * time.Minute),
	}, nil)

	progress, err := suite.service.GetBacktestProgress(taskID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), progress)
	assert.Equal(suite.T(), taskID, progress.TaskID)
	assert.Equal(suite.T(), "running", progress.Status)
	assert.Equal(suite.T(), 65, progress.Progress)
	assert.Equal(suite.T(), "计算组合权重中...", progress.CurrentStep)

	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *StrategyServiceTestSuite) TestStopBacktest() {
	taskID := uint(123)

	suite.mockEngine.On("StopBacktest", taskID).Return(nil)

	err := suite.service.StopBacktest(taskID)

	assert.NoError(suite.T(), err)
	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *StrategyServiceTestSuite) TestGetStrategyAttribution() {
	userID := uint(1)
	strategyID := uint(456)

	// 创建测试策略
	strategy := models.Strategy{
		Name:        "归因测试策略",
		Type:        "top_k",
		Status:      "completed",
		UserID:      userID,
		TotalReturn: 0.156,
		SharpeRatio: 1.45,
	}
	strategy.ID = strategyID
	suite.testDB.DB.Create(&strategy)

	attribution, err := suite.service.GetStrategyAttribution(strategyID, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), attribution)
	assert.Equal(suite.T(), strategyID, attribution.StrategyID)
	assert.NotNil(suite.T(), attribution.FactorAttribution)
	assert.NotNil(suite.T(), attribution.SectorAttribution)
	assert.NotNil(suite.T(), attribution.StyleAttribution)
	assert.Greater(suite.T(), len(attribution.TopHoldings), 0)
}

func (suite *StrategyServiceTestSuite) TestCompareStrategies() {
	userID := uint(1)

	// 创建测试策略
	strategies := []models.Strategy{
		{
			Name:           "策略A",
			Type:           "top_k",
			Status:         "completed",
			UserID:         userID,
			TotalReturn:    0.156,
			SharpeRatio:    1.45,
			MaxDrawdown:    -0.08,
		},
		{
			Name:           "策略B",
			Type:           "long_short",
			Status:         "completed",
			UserID:         userID,
			TotalReturn:    0.189,
			SharpeRatio:    1.62,
			MaxDrawdown:    -0.12,
		},
	}

	for i := range strategies {
		suite.testDB.DB.Create(&strategies[i])
	}

	req := StrategyComparisonRequest{
		StrategyIDs: []uint{strategies[0].ID, strategies[1].ID},
		Metrics:     []string{"return", "sharpe", "drawdown"},
		Benchmark:   "HS300",
		Period:      "all",
	}

	comparison, err := suite.service.CompareStrategies(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), comparison)
	assert.Len(suite.T(), comparison.Strategies, 2)
	assert.Equal(suite.T(), "策略A", comparison.Strategies[0].StrategyName)
	assert.Equal(suite.T(), "策略B", comparison.Strategies[1].StrategyName)
	assert.NotNil(suite.T(), comparison.ComparisonMetrics)
	assert.NotEmpty(suite.T(), comparison.Summary)
}

func (suite *StrategyServiceTestSuite) TestOptimizeStrategy() {
	userID := uint(1)
	strategyID := uint(456)

	// 创建测试策略
	strategy := models.Strategy{
		Name:   "优化测试策略",
		Type:   "top_k",
		Status: "completed",
		UserID: userID,
	}
	strategy.ID = strategyID
	suite.testDB.DB.Create(&strategy)

	req := StrategyOptimizationRequest{
		StrategyID: strategyID,
		Parameters: []OptimizationParameter{
			{
				Name:   "top_k",
				Type:   "int",
				Min:    20,
				Max:    100,
				Step:   10,
				Current: 50,
			},
			{
				Name:   "rebalance_freq",
				Type:   "string",
				Options: []interface{}{"W", "M", "Q"},
				Current: "M",
			},
		},
		Objective:    "sharpe_ratio",
		MaxTrials:    10,
		Method:       "grid_search",
	}

	taskID, err := suite.service.OptimizeStrategy(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), taskID)
	assert.Contains(suite.T(), taskID, "strategy_optimization_")
}

func (suite *StrategyServiceTestSuite) TestExportBacktestReport() {
	userID := uint(1)

	req := BacktestReportExportRequest{
		StrategyIDs:   []uint{1, 2, 3},
		ReportType:    "comprehensive",
		Format:        "pdf",
		Language:      "zh-CN",
		IncludeCharts: true,
		Sections: []string{
			"summary",
			"performance",
			"risk_analysis",
			"attribution",
		},
		Benchmark: "HS300",
	}

	taskID, err := suite.service.ExportBacktestReport(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), taskID)
	assert.Contains(suite.T(), taskID, "backtest_report_")
}

func TestStrategyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StrategyServiceTestSuite))
}