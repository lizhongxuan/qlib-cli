package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"qlib-backend/internal/models"
	"qlib-backend/internal/testutils"
)

type AnalysisServiceTestSuite struct {
	suite.Suite
	service *AnalysisService
	testDB  *testutils.TestDB
}

func (suite *AnalysisServiceTestSuite) SetupSuite() {
	suite.testDB = testutils.SetupTestDB()
	suite.service = NewAnalysisService(suite.testDB.DB)
}

func (suite *AnalysisServiceTestSuite) TearDownSuite() {
	suite.testDB.Cleanup()
}

func (suite *AnalysisServiceTestSuite) SetupTest() {
	suite.testDB.CleanupTables()
}

func (suite *AnalysisServiceTestSuite) TestGetAnalysisOverview() {
	userID := uint(1)

	// 创建测试数据 - 模型
	model := models.Model{
		Name:        "测试模型",
		Type:        "lightgbm",
		Status:      "completed",
		UserID:      userID,
		TrainedAt:   time.Now(),
		TestIC:      0.045,
		TestRankIC:  0.038,
		TestLoss:    0.234,
	}
	suite.testDB.DB.Create(&model)

	// 创建测试数据 - 策略
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

	// 获取分析概览
	overview, err := suite.service.GetAnalysisOverview(userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), overview)
	assert.Equal(suite.T(), 1, overview.TotalModels)
	assert.Equal(suite.T(), 1, overview.TotalStrategies)
	assert.NotNil(suite.T(), overview.BestPerformingModel)
	assert.Equal(suite.T(), model.Name, overview.BestPerformingModel.ModelName)
	assert.Equal(suite.T(), model.TestIC, overview.BestPerformingModel.TestIC)
	assert.NotNil(suite.T(), overview.BestPerformingStrategy)
	assert.Equal(suite.T(), strategy.Name, overview.BestPerformingStrategy.StrategyName)
	assert.Equal(suite.T(), strategy.SharpeRatio, overview.BestPerformingStrategy.SharpeRatio)
}

func (suite *AnalysisServiceTestSuite) TestCompareModels() {
	userID := uint(1)

	// 创建测试模型
	models := []models.Model{
		{
			Name:       "模型A",
			Type:       "lightgbm",
			Status:     "completed",
			UserID:     userID,
			TestIC:     0.045,
			TestRankIC: 0.038,
			TestLoss:   0.234,
			TrainedAt:  time.Now().AddDate(0, 0, -1),
		},
		{
			Name:       "模型B",
			Type:       "xgboost",
			Status:     "completed",
			UserID:     userID,
			TestIC:     0.052,
			TestRankIC: 0.041,
			TestLoss:   0.221,
			TrainedAt:  time.Now(),
		},
	}

	for i := range models {
		suite.testDB.DB.Create(&models[i])
	}

	req := ModelComparisonRequest{
		ModelIDs: []uint{models[0].ID, models[1].ID},
		Metrics:  []string{"ic", "rank_ic", "loss"},
	}

	comparison, err := suite.service.CompareModels(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), comparison)
	assert.Len(suite.T(), comparison.Models, 2)
	assert.Equal(suite.T(), "模型A", comparison.Models[0].ModelName)
	assert.Equal(suite.T(), "模型B", comparison.Models[1].ModelName)
	assert.NotNil(suite.T(), comparison.BestModel)
	assert.Equal(suite.T(), "模型B", comparison.BestModel.ModelName) // 模型B的IC更高
	assert.NotEmpty(suite.T(), comparison.ComparisonSummary)
}

func (suite *AnalysisServiceTestSuite) TestGetFactorImportance() {
	userID := uint(1)
	resultID := uint(123)

	importance, err := suite.service.GetFactorImportance(resultID, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), importance)
	assert.Greater(suite.T(), len(importance.FactorImportance), 0)
	assert.NotEmpty(suite.T(), importance.Method)
	assert.NotNil(suite.T(), importance.TopFactors)

	// 验证因子重要性数据结构
	for _, factor := range importance.FactorImportance {
		assert.NotEmpty(suite.T(), factor.FactorName)
		assert.GreaterOrEqual(suite.T(), factor.Importance, 0.0)
		assert.LessOrEqual(suite.T(), factor.Importance, 1.0)
	}
}

func (suite *AnalysisServiceTestSuite) TestGetStrategyPerformance() {
	userID := uint(1)
	resultID := uint(456)

	performance, err := suite.service.GetStrategyPerformance(resultID, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), performance)
	assert.NotNil(suite.T(), performance.BasicMetrics)
	assert.NotNil(suite.T(), performance.RiskMetrics)
	assert.Greater(suite.T(), len(performance.Returns), 0)
	assert.Greater(suite.T(), len(performance.Positions), 0)

	// 验证基本指标
	assert.NotZero(suite.T(), performance.BasicMetrics.TotalReturn)
	assert.NotZero(suite.T(), performance.BasicMetrics.AnnualReturn)
	assert.NotZero(suite.T(), performance.BasicMetrics.SharpeRatio)

	// 验证风险指标
	assert.NotZero(suite.T(), performance.RiskMetrics.MaxDrawdown)
	assert.Greater(suite.T(), performance.RiskMetrics.Volatility, 0.0)
}

func (suite *AnalysisServiceTestSuite) TestCompareStrategies() {
	userID := uint(1)

	// 创建测试策略
	strategies := []models.Strategy{
		{
			Name:             "策略A",
			Type:             "top_k",
			Status:           "completed",
			UserID:           userID,
			TotalReturn:      0.156,
			AnnualReturn:     0.123,
			SharpeRatio:      1.45,
			MaxDrawdown:      -0.08,
			VolatilityAnnual: 0.15,
			BacktestAt:       time.Now().AddDate(0, 0, -1),
		},
		{
			Name:             "策略B",
			Type:             "long_short",
			Status:           "completed",
			UserID:           userID,
			TotalReturn:      0.189,
			AnnualReturn:     0.145,
			SharpeRatio:      1.62,
			MaxDrawdown:      -0.12,
			VolatilityAnnual: 0.18,
			BacktestAt:       time.Now(),
		},
	}

	for i := range strategies {
		suite.testDB.DB.Create(&strategies[i])
	}

	req := StrategyComparisonRequest{
		StrategyIDs: []uint{strategies[0].ID, strategies[1].ID},
		Metrics:     []string{"return", "sharpe", "drawdown"},
		Benchmark:   "HS300",
	}

	comparison, err := suite.service.CompareStrategies(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), comparison)
	assert.Len(suite.T(), comparison.Strategies, 2)
	assert.Equal(suite.T(), "策略A", comparison.Strategies[0].StrategyName)
	assert.Equal(suite.T(), "策略B", comparison.Strategies[1].StrategyName)
	assert.NotNil(suite.T(), comparison.BestStrategy)
	assert.Equal(suite.T(), "策略B", comparison.BestStrategy.StrategyName) // 策略B的夏普比率更高
	assert.NotEmpty(suite.T(), comparison.ComparisonSummary)
}

func (suite *AnalysisServiceTestSuite) TestGenerateAnalysisReport() {
	userID := uint(1)

	req := AnalysisReportRequest{
		ReportType:    "comprehensive",
		ModelIDs:      []uint{1, 2},
		StrategyIDs:   []uint{3, 4},
		IncludeCharts: true,
		Format:        "pdf",
		Language:      "zh-CN",
		Sections: []string{
			"model_performance",
			"strategy_performance",
			"factor_analysis",
			"risk_analysis",
		},
	}

	taskID, err := suite.service.GenerateAnalysisReport(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), taskID)
	assert.Contains(suite.T(), taskID, "analysis_report_")
}

func (suite *AnalysisServiceTestSuite) TestGetReportStatus() {
	taskID := "analysis_report_test_123"
	userID := uint(1)

	status, err := suite.service.GetReportStatus(taskID, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), taskID, status.TaskID)
	assert.Contains(suite.T(), []string{"pending", "running", "completed", "failed"}, status.Status)
	assert.GreaterOrEqual(suite.T(), status.Progress, 0)
	assert.LessOrEqual(suite.T(), status.Progress, 100)
}

func (suite *AnalysisServiceTestSuite) TestGetSummaryStats() {
	userID := uint(1)

	// 创建测试数据
	model := models.Model{
		Name:       "统计测试模型",
		Type:       "lightgbm",
		Status:     "completed",
		UserID:     userID,
		TestIC:     0.045,
		TestRankIC: 0.038,
		TrainedAt:  time.Now(),
	}
	suite.testDB.DB.Create(&model)

	strategy := models.Strategy{
		Name:           "统计测试策略",
		Type:           "top_k",
		Status:         "completed",
		UserID:         userID,
		TotalReturn:    0.156,
		SharpeRatio:    1.45,
		MaxDrawdown:    -0.08,
		BacktestAt:     time.Now(),
	}
	suite.testDB.DB.Create(&strategy)

	stats, err := suite.service.GetSummaryStats(userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.NotNil(suite.T(), stats.ModelStats)
	assert.NotNil(suite.T(), stats.StrategyStats)
	assert.NotNil(suite.T(), stats.OverallPerformance)

	// 验证模型统计
	assert.Equal(suite.T(), 1, stats.ModelStats.TotalModels)
	assert.Greater(suite.T(), stats.ModelStats.AvgIC, 0.0)

	// 验证策略统计
	assert.Equal(suite.T(), 1, stats.StrategyStats.TotalStrategies)
	assert.Greater(suite.T(), stats.StrategyStats.AvgSharpe, 0.0)
}

func (suite *AnalysisServiceTestSuite) TestMultiResultComparison() {
	userID := uint(1)

	req := MultiResultComparisonRequest{
		ModelResults:    []uint{1, 2, 3},
		StrategyResults: []uint{4, 5, 6},
		ComparisonType:  "comprehensive",
		Metrics: []string{
			"return",
			"sharpe",
			"ic",
			"drawdown",
		},
		GroupBy: "type",
	}

	comparison, err := suite.service.MultiResultComparison(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), comparison)
	assert.NotNil(suite.T(), comparison.ModelComparison)
	assert.NotNil(suite.T(), comparison.StrategyComparison)
	assert.NotNil(suite.T(), comparison.CrossComparison)
	assert.NotEmpty(suite.T(), comparison.ComparisonSummary)
	assert.NotNil(suite.T(), comparison.BestOverall)
}

func TestAnalysisServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AnalysisServiceTestSuite))
}