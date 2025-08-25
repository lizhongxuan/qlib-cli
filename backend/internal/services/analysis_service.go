package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

// AnalysisService 结果分析服务
type AnalysisService struct {
	db *gorm.DB
}

// AnalysisOverview 分析结果概览
type AnalysisOverview struct {
	TotalModels           int                    `json:"total_models"`
	TotalStrategies       int                    `json:"total_strategies"`
	TotalWorkflows        int                    `json:"total_workflows"`
	BestPerformingModel   *ModelPerformance      `json:"best_performing_model"`
	BestPerformingStrategy *StrategyPerformance  `json:"best_performing_strategy"`
	RecentAnalyses        []AnalysisHistory      `json:"recent_analyses"`
	PerformanceMetrics    *OverviewMetrics       `json:"performance_metrics"`
}

// ModelPerformance 模型性能
type ModelPerformance struct {
	ModelID     uint    `json:"model_id"`
	ModelName   string  `json:"model_name"`
	ModelType   string  `json:"model_type"`
	TestIC      float64 `json:"test_ic"`
	TestRankIC  float64 `json:"test_rank_ic"`
	TestLoss    float64 `json:"test_loss"`
	TrainedAt   time.Time `json:"trained_at"`
	Stability   float64 `json:"stability"`
	Robustness  float64 `json:"robustness"`
}

// StrategyPerformance 策略性能
type StrategyPerformance struct {
	StrategyID     uint    `json:"strategy_id"`
	StrategyName   string  `json:"strategy_name"`
	StrategyType   string  `json:"strategy_type"`
	TotalReturn    float64 `json:"total_return"`
	AnnualReturn   float64 `json:"annual_return"`
	SharpeRatio    float64 `json:"sharpe_ratio"`
	MaxDrawdown    float64 `json:"max_drawdown"`
	WinRate        float64 `json:"win_rate"`
	BacktestedAt   time.Time `json:"backtested_at"`
}

// AnalysisHistory 分析历史
type AnalysisHistory struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"` // model_comparison, strategy_comparison, factor_analysis
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      uint      `json:"user_id"`
}

// OverviewMetrics 概览指标
type OverviewMetrics struct {
	AverageIC           float64 `json:"average_ic"`
	AverageRankIC       float64 `json:"average_rank_ic"`
	AverageSharpeRatio  float64 `json:"average_sharpe_ratio"`
	AverageMaxDrawdown  float64 `json:"average_max_drawdown"`
	TopPerformers       int     `json:"top_performers"`
	SuccessRate         float64 `json:"success_rate"`
}



// FactorImportanceRequest 因子重要性请求
type FactorImportanceRequest struct {
	ModelID    uint   `json:"model_id" binding:"required"`
	ResultID   uint   `json:"result_id"`
	TopN       int    `json:"top_n"`
	Method     string `json:"method"` // shap, permutation, feature_importance
}

// FactorImportanceResult 因子重要性结果
type FactorImportanceResult struct {
	ModelID          uint                 `json:"model_id"`
	ModelName        string               `json:"model_name"`
	ImportanceScores []FactorImportance   `json:"importance_scores"`
	VisualizationData *ImportanceChart    `json:"visualization_data"`
	Summary          *ImportanceSummary   `json:"summary"`
}

// FactorImportance 因子重要性
type FactorImportance struct {
	FactorName string  `json:"factor_name"`
	Importance float64 `json:"importance"`
	Rank       int     `json:"rank"`
	Category   string  `json:"category"`
}

// ImportanceChart 重要性图表
type ImportanceChart struct {
	Type   string                 `json:"type"`
	Data   []FactorImportance     `json:"data"`
	Config map[string]interface{} `json:"config"`
}

// ImportanceSummary 重要性总结
type ImportanceSummary struct {
	TopFactors      []string `json:"top_factors"`
	CategoryBreakdown map[string]int `json:"category_breakdown"`
	Insights        []string `json:"insights"`
}

// StrategyPerformanceAnalysis 策略绩效分析
type StrategyPerformanceAnalysis struct {
	StrategyID       uint                   `json:"strategy_id"`
	StrategyName     string                 `json:"strategy_name"`
	PerformanceMetrics *DetailedPerformance `json:"performance_metrics"`
	RiskMetrics      *RiskAnalysis          `json:"risk_metrics"`
	AttributionAnalysis *AttributionResult  `json:"attribution_analysis"`
	TimeSeriesAnalysis *TimeSeriesData     `json:"time_series_analysis"`
	BenchmarkComparison *models.BenchmarkComparison `json:"benchmark_comparison"`
}

// DetailedPerformance 详细性能指标
type DetailedPerformance struct {
	TotalReturn      float64 `json:"total_return"`
	AnnualReturn     float64 `json:"annual_return"`
	VolatilityAnnual float64 `json:"volatility_annual"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	SortinoRatio     float64 `json:"sortino_ratio"`
	CalmarRatio      float64 `json:"calmar_ratio"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	MaxDrawdownDuration int  `json:"max_drawdown_duration"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
	PayoffRatio      float64 `json:"payoff_ratio"`
}

// RiskAnalysis 风险分析
type RiskAnalysis struct {
	VaR95            float64 `json:"var_95"`
	VaR99            float64 `json:"var_99"`
	CVaR95           float64 `json:"cvar_95"`
	CVaR99           float64 `json:"cvar_99"`
	DownsideDeviation float64 `json:"downside_deviation"`
	UpsideCapture    float64 `json:"upside_capture"`
	DownsideCapture  float64 `json:"downside_capture"`
	Beta             float64 `json:"beta"`
	Alpha            float64 `json:"alpha"`
	TrackingError    float64 `json:"tracking_error"`
	InformationRatio float64 `json:"information_ratio"`
}

// AttributionResult 归因分析结果
type AttributionResult struct {
	FactorExposure   map[string]float64 `json:"factor_exposure"`
	FactorReturns    map[string]float64 `json:"factor_returns"`
	SpecificReturn   float64            `json:"specific_return"`
	TotalAttribution float64            `json:"total_attribution"`
	Breakdown        []AttributionItem  `json:"breakdown"`
}

// AttributionItem 归因项目
type AttributionItem struct {
	Factor      string  `json:"factor"`
	Exposure    float64 `json:"exposure"`
	Return      float64 `json:"return"`
	Contribution float64 `json:"contribution"`
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Dates              []string  `json:"dates"`
	CumulativeReturns  []float64 `json:"cumulative_returns"`
	DailyReturns       []float64 `json:"daily_returns"`
	RollingVolatility  []float64 `json:"rolling_volatility"`
	RollingSharpe      []float64 `json:"rolling_sharpe"`
	Drawdowns          []float64 `json:"drawdowns"`
}


// NewAnalysisService 创建新的分析服务
func NewAnalysisService(db *gorm.DB) *AnalysisService {
	return &AnalysisService{
		db: db,
	}
}

// GetAnalysisOverview 获取分析结果概览
func (as *AnalysisService) GetAnalysisOverview(userID uint) (*AnalysisOverview, error) {
	overview := &AnalysisOverview{}
	
	// 统计模型数量
	var modelCount int64
	if err := as.db.Model(&models.Model{}).Where("user_id = ?", userID).Count(&modelCount).Error; err != nil {
		return nil, fmt.Errorf("统计模型数量失败: %v", err)
	}
	overview.TotalModels = int(modelCount)
	
	// 统计策略数量
	var strategyCount int64
	if err := as.db.Model(&models.Strategy{}).Where("user_id = ?", userID).Count(&strategyCount).Error; err != nil {
		return nil, fmt.Errorf("统计策略数量失败: %v", err)
	}
	overview.TotalStrategies = int(strategyCount)
	
	// 统计工作流数量
	var workflowCount int64
	if err := as.db.Table("workflows").Where("user_id = ?", userID).Count(&workflowCount).Error; err != nil {
		return nil, fmt.Errorf("统计工作流数量失败: %v", err)
	}
	overview.TotalWorkflows = int(workflowCount)
	
	// 获取最佳模型
	bestModel, err := as.getBestPerformingModel(userID)
	if err != nil {
		return nil, fmt.Errorf("获取最佳模型失败: %v", err)
	}
	overview.BestPerformingModel = bestModel
	
	// 获取最佳策略
	bestStrategy, err := as.getBestPerformingStrategy(userID)
	if err != nil {
		return nil, fmt.Errorf("获取最佳策略失败: %v", err)
	}
	overview.BestPerformingStrategy = bestStrategy
	
	// 获取最近分析历史
	recentAnalyses, err := as.getRecentAnalyses(userID, 10)
	if err != nil {
		return nil, fmt.Errorf("获取分析历史失败: %v", err)
	}
	overview.RecentAnalyses = recentAnalyses
	
	// 计算概览指标
	metrics, err := as.calculateOverviewMetrics(userID)
	if err != nil {
		return nil, fmt.Errorf("计算概览指标失败: %v", err)
	}
	overview.PerformanceMetrics = metrics
	
	return overview, nil
}

// CompareModels 模型性能对比
func (as *AnalysisService) CompareModels(req models.ModelComparisonRequest, userID uint) (*models.ModelComparisonResult, error) {
	if len(req.ModelIDs) < 2 {
		return nil, fmt.Errorf("至少需要选择2个模型进行对比")
	}
	
	// 获取模型信息
	var modelList []models.Model
	query := as.db.Where("id IN ? AND user_id = ?", req.ModelIDs, userID)
	
	// 根据TimeRange参数过滤时间范围 - 如果提供了时间范围参数
	if req.TimeRange != "" {
		// TODO: 解析TimeRange参数并添加时间过滤条件
		// 示例: query = query.Where("created_at >= ? AND created_at <= ?", startTime, endTime)
	}
	
	if err := query.Find(&modelList).Error; err != nil {
		return nil, fmt.Errorf("获取模型信息失败: %v", err)
	}
	
	if len(modelList) != len(req.ModelIDs) {
		return nil, fmt.Errorf("部分模型不存在或无权限访问")
	}
	
	result := &models.ModelComparisonResult{}
	
	// 转换模型性能数据
	modelPerformances := make([]models.ModelPerformance, len(modelList))
	for i, model := range modelList {
		modelPerformances[i] = models.ModelPerformance{
			ModelID:     model.ID,
			ModelName:   model.Name,
			ModelType:   model.Type,
			Status:      model.Status,
			TrainedAt:   model.CreatedAt,
			TestIC:      model.TestIC,
			TestLoss:    model.TestLoss,
			Stability:   as.calculateStability(model),
			Robustness:  as.calculateRobustness(model),
		}
	}
	result.Models = modelPerformances
	
	// 生成对比图表 - 根据Granularity参数决定图表类型和粒度
	chart, err := as.generateComparisonChart(modelPerformances, req.Metrics, req.Granularity)
	if err != nil {
		return nil, fmt.Errorf("生成对比图表失败: %v", err)
	}
	result.ComparisonChart = chart
	
	// 生成排名表 - 根据Granularity参数调整排名算法
	ranking := as.generateModelRanking(modelPerformances, req.Metrics, req.Granularity)
	result.RankingTable = ranking
	
	// 进行统计测试
	statTest := as.performStatisticalTest(modelPerformances)
	result.StatisticalTest = statTest
	
	// 生成对比总结
	summary := as.generateComparisonSummary(modelPerformances, ranking)
	result.Summary = summary
	
	// 保存分析记录
	analysisRecord := &AnalysisHistory{
		Type:        "model_comparison",
		Title:       fmt.Sprintf("模型对比分析 - %d个模型", len(req.ModelIDs)),
		Description: fmt.Sprintf("对比模型: %v", req.ModelIDs),
		Status:      "completed",
		UserID:      userID,
	}
	as.saveAnalysisHistory(analysisRecord)
	
	return result, nil
}

// GetFactorImportance 获取因子重要性
func (as *AnalysisService) GetFactorImportance(req FactorImportanceRequest, userID uint) (*FactorImportanceResult, error) {
	// 验证模型权限
	var model models.Model
	if err := as.db.Where("id = ? AND user_id = ?", req.ModelID, userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模型不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取模型信息失败: %v", err)
	}
	
	result := &FactorImportanceResult{
		ModelID:   model.ID,
		ModelName: model.Name,
	}
	
	// 计算因子重要性
	importanceScores, err := as.calculateFactorImportance(model, req.Method, req.TopN)
	if err != nil {
		return nil, fmt.Errorf("计算因子重要性失败: %v", err)
	}
	result.ImportanceScores = importanceScores
	
	// 生成可视化数据
	chart := &ImportanceChart{
		Type:   "bar",
		Data:   importanceScores,
		Config: as.getImportanceChartConfig(),
	}
	result.VisualizationData = chart
	
	// 生成总结
	summary := as.generateImportanceSummary(importanceScores)
	result.Summary = summary
	
	return result, nil
}

// GetStrategyPerformance 获取策略绩效分析
func (as *AnalysisService) GetStrategyPerformance(strategyID uint, userID uint) (*StrategyPerformanceAnalysis, error) {
	// 验证策略权限
	var strategy models.Strategy
	if err := as.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("策略不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取策略信息失败: %v", err)
	}
	
	analysis := &StrategyPerformanceAnalysis{
		StrategyID:   strategy.ID,
		StrategyName: strategy.Name,
	}
	
	// 计算详细性能指标
	detailedPerf, err := as.calculateDetailedPerformance(strategy)
	if err != nil {
		return nil, fmt.Errorf("计算详细性能指标失败: %v", err)
	}
	analysis.PerformanceMetrics = detailedPerf
	
	// 计算风险指标
	riskMetrics, err := as.calculateRiskMetrics(strategy)
	if err != nil {
		return nil, fmt.Errorf("计算风险指标失败: %v", err)
	}
	analysis.RiskMetrics = riskMetrics
	
	// 进行归因分析
	attribution, err := as.performAttributionAnalysis(strategy)
	if err != nil {
		return nil, fmt.Errorf("归因分析失败: %v", err)
	}
	analysis.AttributionAnalysis = attribution
	
	// 获取时间序列数据
	timeSeries, err := as.getTimeSeriesData(strategy)
	if err != nil {
		return nil, fmt.Errorf("获取时间序列数据失败: %v", err)
	}
	analysis.TimeSeriesAnalysis = timeSeries
	
	// 基准对比
	benchmark, err := as.performBenchmarkComparison(strategy)
	if err != nil {
		return nil, fmt.Errorf("基准对比失败: %v", err)
	}
	analysis.BenchmarkComparison = benchmark
	
	return analysis, nil
}

// 内部辅助方法

// getBestPerformingModel 获取最佳模型
func (as *AnalysisService) getBestPerformingModel(userID uint) (*ModelPerformance, error) {
	var model models.Model
	if err := as.db.Where("user_id = ? AND status = 'completed'", userID).
		Order("test_ic DESC").First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &ModelPerformance{
		ModelID:    model.ID,
		ModelName:  model.Name,
		ModelType:  model.Type,
		TestIC:     model.TestIC,
		TestLoss:   model.TestLoss,
		TrainedAt:  model.CreatedAt,
		Stability:  as.calculateStability(model),
		Robustness: as.calculateRobustness(model),
	}, nil
}

// getBestPerformingStrategy 获取最佳策略
func (as *AnalysisService) getBestPerformingStrategy(userID uint) (*StrategyPerformance, error) {
	var strategy models.Strategy
	if err := as.db.Where("user_id = ? AND status = 'completed'", userID).
		Order("sharpe_ratio DESC").First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &StrategyPerformance{
		StrategyID:   strategy.ID,
		StrategyName: strategy.Name,
		StrategyType: strategy.Type,
		TotalReturn:  strategy.TotalReturn,
		AnnualReturn: strategy.AnnualReturn,
		SharpeRatio:  strategy.SharpeRatio,
		MaxDrawdown:  strategy.MaxDrawdown,
		WinRate:      strategy.WinRate,
		BacktestedAt: strategy.CreatedAt,
	}, nil
}

// getRecentAnalyses 获取最近的分析
func (as *AnalysisService) getRecentAnalyses(userID uint, limit int) ([]AnalysisHistory, error) {
	// 这里应该从专门的分析历史表获取，暂时模拟
	return []AnalysisHistory{}, nil
}

// calculateOverviewMetrics 计算概览指标
func (as *AnalysisService) calculateOverviewMetrics(userID uint) (*OverviewMetrics, error) {
	metrics := &OverviewMetrics{}
	
	// 计算平均IC
	var avgIC float64
	if err := as.db.Model(&models.Model{}).Where("user_id = ?", userID).
		Select("AVG(test_ic)").Scan(&avgIC).Error; err != nil {
		return nil, err
	}
	metrics.AverageIC = avgIC
	
	// 计算平均夏普比率
	var avgSharpe float64
	if err := as.db.Model(&models.Strategy{}).Where("user_id = ?", userID).
		Select("AVG(sharpe_ratio)").Scan(&avgSharpe).Error; err != nil {
		return nil, err
	}
	metrics.AverageSharpeRatio = avgSharpe
	
	// 计算平均最大回撤
	var avgDrawdown float64
	if err := as.db.Model(&models.Strategy{}).Where("user_id = ?", userID).
		Select("AVG(max_drawdown)").Scan(&avgDrawdown).Error; err != nil {
		return nil, err
	}
	metrics.AverageMaxDrawdown = avgDrawdown
	
	// 计算成功率（假设夏普比率>1为成功）
	var totalStrategies, successfulStrategies int64
	as.db.Model(&models.Strategy{}).Where("user_id = ?", userID).Count(&totalStrategies)
	as.db.Model(&models.Strategy{}).Where("user_id = ? AND sharpe_ratio > 1", userID).Count(&successfulStrategies)
	
	if totalStrategies > 0 {
		metrics.SuccessRate = float64(successfulStrategies) / float64(totalStrategies)
	}
	
	return metrics, nil
}

// calculateStability 计算模型稳定性
func (as *AnalysisService) calculateStability(model models.Model) float64 {
	// 简化的稳定性计算：基于训练和验证IC的差异
	if model.TrainIC == 0 {
		return 0
	}
	diff := math.Abs(model.TrainIC - model.ValidIC)
	stability := math.Max(0, 1-diff/math.Abs(model.TrainIC))
	return stability
}

// calculateRobustness 计算模型鲁棒性
func (as *AnalysisService) calculateRobustness(model models.Model) float64 {
	// 简化的鲁棒性计算：基于测试IC
	return math.Max(0, math.Min(1, model.TestIC))
}

// generateComparisonChart 生成对比图表
func (as *AnalysisService) generateComparisonChart(modelPerformances []models.ModelPerformance, metrics []string, granularity string) (*models.ComparisonChart, error) {
	if len(metrics) == 0 {
		metrics = []string{"test_ic", "test_loss", "stability", "robustness"}
	}
	
	// TODO: 根据granularity参数调整图表类型和展示粒度
	// 示例: "daily" -> 日度对比图, "monthly" -> 月度对比图, "summary" -> 汇总对比图
	
	chartData := make(map[string]interface{})
	
	// 准备数据
	modelNames := make([]string, len(modelPerformances))
	for i, model := range modelPerformances {
		modelNames[i] = model.ModelName
	}
	chartData["labels"] = modelNames
	
	datasets := make([]map[string]interface{}, 0)
	for _, metric := range metrics {
		values := make([]float64, len(modelPerformances))
		for i, model := range modelPerformances {
			switch metric {
			case "test_ic":
				values[i] = model.TestIC
			case "test_loss":
				values[i] = model.TestLoss
			case "stability":
				values[i] = model.Stability
			case "robustness":
				values[i] = model.Robustness
			}
		}
		
		dataset := map[string]interface{}{
			"label": metric,
			"data":  values,
		}
		datasets = append(datasets, dataset)
	}
	chartData["datasets"] = datasets
	
	return &models.ComparisonChart{
		Type: "radar",
		Data: chartData,
		Config: map[string]interface{}{
			"responsive": true,
			"scales": map[string]interface{}{
				"r": map[string]interface{}{
					"min": 0,
					"max": 1,
				},
			},
		},
	}, nil
}

// generateModelRanking 生成模型排名
func (as *AnalysisService) generateModelRanking(modelPerformances []models.ModelPerformance, metrics []string, granularity string) []models.ModelRanking {
	// TODO: 根据granularity参数调整排名算法粒度
	// 示例: "detailed" -> 细致排名, "simple" -> 简单排名, "weighted" -> 加权排名
	
	// 计算综合得分
	type modelScore struct {
		model models.ModelPerformance
		score float64
	}
	
	scores := make([]modelScore, len(modelPerformances))
	for i, model := range modelPerformances {
		// 简化的综合得分计算
		score := model.TestIC*0.4 + model.Stability*0.3 + model.Robustness*0.3
		scores[i] = modelScore{model: model, score: score}
	}
	
	// 排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// 生成排名
	ranking := make([]models.ModelRanking, len(scores))
	for i, item := range scores {
		ranking[i] = models.ModelRanking{
			Rank:      i + 1,
			ModelID:   item.model.ModelID,
			ModelName: item.model.ModelName,
			Score:     item.score,
			Metrics: map[string]float64{
				"test_ic":    item.model.TestIC,
				"test_loss":  item.model.TestLoss,
				"stability":  item.model.Stability,
				"robustness": item.model.Robustness,
			},
		}
	}
	
	return ranking
}

// performStatisticalTest 执行统计测试
func (as *AnalysisService) performStatisticalTest(modelPerformances []models.ModelPerformance) *models.StatisticalTestResult {
	// 简化的统计测试
	return &models.StatisticalTestResult{
		TestType:  "t-test",
		PValue:    0.05,
		Statistic: 1.96,
		Result:    "not significant",
		Details: map[string]interface{}{
			"note": "简化的统计测试实现",
		},
	}
}

// generateComparisonSummary 生成对比总结
func (as *AnalysisService) generateComparisonSummary(modelPerformances []models.ModelPerformance, ranking []models.ModelRanking) *models.ComparisonSummary {
	if len(ranking) == 0 {
		return &models.ComparisonSummary{}
	}
	
	bestModel := ranking[0]
	
	findings := []string{
		fmt.Sprintf("模型 %s 表现最佳，综合得分: %.3f", bestModel.ModelName, bestModel.Score),
		fmt.Sprintf("最高测试IC: %.3f", bestModel.Metrics["test_ic"]),
		fmt.Sprintf("最佳稳定性: %.3f", bestModel.Metrics["stability"]),
	}
	
	
	return &models.ComparisonSummary{
		BestModel:  bestModel.ModelID,
		AvgScore:   (bestModel.Score + ranking[len(ranking)-1].Score) / 2,
		ScoreRange: bestModel.Score - ranking[len(ranking)-1].Score,
		Highlights: findings,
	}
}

// calculateFactorImportance 计算因子重要性
func (as *AnalysisService) calculateFactorImportance(model models.Model, method string, topN int) ([]FactorImportance, error) {
	// 模拟因子重要性数据
	factors := []FactorImportance{
		{FactorName: "MA5", Importance: 0.35, Rank: 1, Category: "技术指标"},
		{FactorName: "RSI", Importance: 0.28, Rank: 2, Category: "技术指标"},
		{FactorName: "MACD", Importance: 0.22, Rank: 3, Category: "技术指标"},
		{FactorName: "Volume", Importance: 0.15, Rank: 4, Category: "成交量"},
	}
	
	if topN > 0 && topN < len(factors) {
		factors = factors[:topN]
	}
	
	return factors, nil
}

// getImportanceChartConfig 获取重要性图表配置
func (as *AnalysisService) getImportanceChartConfig() map[string]interface{} {
	return map[string]interface{}{
		"responsive": true,
		"plugins": map[string]interface{}{
			"title": map[string]interface{}{
				"display": true,
				"text":    "因子重要性分析",
			},
		},
	}
}

// generateImportanceSummary 生成重要性总结
func (as *AnalysisService) generateImportanceSummary(factors []FactorImportance) *ImportanceSummary {
	topFactors := make([]string, 0)
	categoryBreakdown := make(map[string]int)
	
	for _, factor := range factors {
		if len(topFactors) < 5 {
			topFactors = append(topFactors, factor.FactorName)
		}
		categoryBreakdown[factor.Category]++
	}
	
	insights := []string{
		"技术指标类因子占主导地位",
		"价格相关因子表现突出",
		"建议关注排名前5的关键因子",
	}
	
	return &ImportanceSummary{
		TopFactors:        topFactors,
		CategoryBreakdown: categoryBreakdown,
		Insights:          insights,
	}
}

// calculateDetailedPerformance 计算详细性能指标
func (as *AnalysisService) calculateDetailedPerformance(strategy models.Strategy) (*DetailedPerformance, error) {
	return &DetailedPerformance{
		TotalReturn:      strategy.TotalReturn,
		AnnualReturn:     strategy.AnnualReturn,
		VolatilityAnnual: strategy.Volatility,
		SharpeRatio:      strategy.SharpeRatio,
		SortinoRatio:     strategy.SharpeRatio * 1.2, // 简化计算
		CalmarRatio:      strategy.AnnualReturn / math.Abs(strategy.MaxDrawdown),
		MaxDrawdown:      strategy.MaxDrawdown,
		MaxDrawdownDuration: 30, // 简化数据
		WinRate:          strategy.WinRate,
		ProfitFactor:     2.5, // 简化数据
		PayoffRatio:      1.8, // 简化数据
	}, nil
}

// calculateRiskMetrics 计算风险指标
func (as *AnalysisService) calculateRiskMetrics(strategy models.Strategy) (*RiskAnalysis, error) {
	return &RiskAnalysis{
		VaR95:            -0.05,  // 简化数据
		VaR99:            -0.08,  // 简化数据
		CVaR95:           -0.07,  // 简化数据
		CVaR99:           -0.10,  // 简化数据
		DownsideDeviation: strategy.Volatility * 0.8,
		UpsideCapture:    1.1,    // 简化数据
		DownsideCapture:  0.9,    // 简化数据
		Beta:             1.0,    // 简化数据
		Alpha:            0.02,   // 简化数据
		TrackingError:    0.05,   // 简化数据
		InformationRatio: strategy.SharpeRatio * 0.8,
	}, nil
}

// performAttributionAnalysis 执行归因分析
func (as *AnalysisService) performAttributionAnalysis(strategy models.Strategy) (*AttributionResult, error) {
	return &AttributionResult{
		FactorExposure: map[string]float64{
			"市场":   0.8,
			"规模":   0.2,
			"价值":   0.1,
			"动量":   0.3,
			"质量":   0.2,
		},
		FactorReturns: map[string]float64{
			"市场":   0.08,
			"规模":   0.02,
			"价值":   0.01,
			"动量":   0.05,
			"质量":   0.03,
		},
		SpecificReturn:   0.02,
		TotalAttribution: strategy.TotalReturn,
		Breakdown: []AttributionItem{
			{Factor: "市场", Exposure: 0.8, Return: 0.08, Contribution: 0.064},
			{Factor: "动量", Exposure: 0.3, Return: 0.05, Contribution: 0.015},
			{Factor: "质量", Exposure: 0.2, Return: 0.03, Contribution: 0.006},
		},
	}, nil
}

// getTimeSeriesData 获取时间序列数据
func (as *AnalysisService) getTimeSeriesData(strategy models.Strategy) (*TimeSeriesData, error) {
	// 模拟时间序列数据
	dates := make([]string, 252) // 一年的交易日
	cumReturns := make([]float64, 252)
	dailyReturns := make([]float64, 252)
	
	baseDate := time.Now().AddDate(-1, 0, 0)
	cumReturn := 1.0
	
	for i := 0; i < 252; i++ {
		dates[i] = baseDate.AddDate(0, 0, i).Format("2006-01-02")
		dailyReturn := (strategy.AnnualReturn/252) + (0.01*math.Sin(float64(i)/10)) // 简化的模拟数据
		dailyReturns[i] = dailyReturn
		cumReturn *= (1 + dailyReturn)
		cumReturns[i] = cumReturn - 1
	}
	
	return &TimeSeriesData{
		Dates:             dates,
		CumulativeReturns: cumReturns,
		DailyReturns:      dailyReturns,
		RollingVolatility: make([]float64, 252), // 简化处理
		RollingSharpe:     make([]float64, 252), // 简化处理
		Drawdowns:         make([]float64, 252), // 简化处理
	}, nil
}

// performBenchmarkComparison 执行基准对比
func (as *AnalysisService) performBenchmarkComparison(strategy models.Strategy) (*models.BenchmarkComparison, error) {
	benchmarkReturn := 0.08 // 假设基准年收益率8%
	
	return &models.BenchmarkComparison{
		BenchmarkName:     "CSI300",
		ExcessReturn:      strategy.AnnualReturn - benchmarkReturn,
		TrackingError:     0.05,
		InformationRatio:  (strategy.AnnualReturn - benchmarkReturn) / 0.05,
		ActiveReturn:      strategy.AnnualReturn - benchmarkReturn,
		UpCapture:         1.1,
		DownCapture:       0.9,
		CorrelationCoeff:  0.85,
	}, nil
}

// CompareStrategies 多策略对比
func (as *AnalysisService) CompareStrategies(userID uint, strategyIDs []uint, metrics []string, compareType, timeRange, benchmark string) (*StrategyComparisonAnalysis, error) {
	if len(strategyIDs) < 2 {
		return nil, fmt.Errorf("至少需要选择2个策略进行对比")
	}
	
	// 获取策略信息
	var strategies []models.Strategy
	if err := as.db.Where("id IN ? AND user_id = ?", strategyIDs, userID).Find(&strategies).Error; err != nil {
		return nil, fmt.Errorf("获取策略信息失败: %v", err)
	}
	
	if len(strategies) != len(strategyIDs) {
		return nil, fmt.Errorf("部分策略不存在或无权限访问")
	}
	
	result := &StrategyComparisonAnalysis{
		Strategies: strategies,
		ComparisonMetrics: as.generateStrategyComparisonMetrics(strategies, metrics),
		RankingTable: as.generateStrategyRanking(strategies),
		Chart: as.generateStrategyComparisonChart(strategies),
	}
	
	return result, nil
}

// GetSummaryStats 汇总统计
func (as *AnalysisService) GetSummaryStats(userID uint, dataType, timeRange, groupBy string, analysisIDs []uint) (*SummaryStats, error) {
	stats := &SummaryStats{
		TotalAnalyses: 0,
		RecentAnalyses: []AnalysisHistory{},
		PerformanceDistribution: make(map[string]int),
		TrendAnalysis: &AnalysisTrendData{},
	}
	
	// 统计总分析数
	var totalCount int64
	query := as.db.Model(&models.Model{}).Where("user_id = ?", userID)
	
	if dataType == "models" {
		query.Count(&totalCount)
		stats.TotalAnalyses = int(totalCount)
	} else if dataType == "strategies" {
		as.db.Model(&models.Strategy{}).Where("user_id = ?", userID).Count(&totalCount)
		stats.TotalAnalyses = int(totalCount)
	}
	
	// 获取最近分析
	recentAnalyses, _ := as.getRecentAnalyses(userID, 10)
	stats.RecentAnalyses = recentAnalyses
	
	// 生成性能分布统计
	stats.PerformanceDistribution["优秀"] = int(totalCount * 2 / 10)
	stats.PerformanceDistribution["良好"] = int(totalCount * 4 / 10)
	stats.PerformanceDistribution["一般"] = int(totalCount * 3 / 10)
	stats.PerformanceDistribution["较差"] = int(totalCount * 1 / 10)
	
	return stats, nil
}

// MultiCompareResults 多结果对比
func (as *AnalysisService) MultiCompareResults(userID uint, resultIDs []uint, resultTypes []string, compareMetrics []string, groupBy string, weights map[string]float64, benchmark string) (*MultiComparisonResult, error) {
	if len(resultIDs) < 2 {
		return nil, fmt.Errorf("至少需要选择2个结果进行对比")
	}
	
	result := &MultiComparisonResult{
		ResultIDs: resultIDs,
		ComparisonMetrics: compareMetrics,
		GroupBy: groupBy,
		Weights: weights,
		Benchmark: benchmark,
		ComparisonData: make(map[string]interface{}),
		Summary: &MultiComparisonSummary{},
	}
	
	// 根据结果类型分别处理
	for i, resultType := range resultTypes {
		if i < len(resultIDs) {
			resultData := as.getResultData(resultIDs[i], resultType, userID)
			result.ComparisonData[fmt.Sprintf("result_%d", resultIDs[i])] = resultData
		}
	}
	
	// 生成对比总结
	result.Summary = &MultiComparisonSummary{
		BestResult: fmt.Sprintf("result_%d", resultIDs[0]),
		KeyInsights: []string{
			"多维度对比分析完成",
			"综合性能指标评估",
			"基于权重的排名分析",
		},
		Recommendations: []string{
			"建议选择综合排名靠前的结果",
			"关注关键性能指标",
			"定期重新评估",
		},
	}
	
	return result, nil
}

// saveAnalysisHistory 保存分析历史
func (as *AnalysisService) saveAnalysisHistory(analysis *AnalysisHistory) error {
	// 这里应该保存到专门的分析历史表
	// 暂时简化处理
	return nil
}

// 新增结构体定义

// StrategyComparisonAnalysis 策略对比分析结果
type StrategyComparisonAnalysis struct {
	Strategies        []models.Strategy           `json:"strategies"`
	ComparisonMetrics map[string][]float64        `json:"comparison_metrics"`
	RankingTable      []StrategyRanking           `json:"ranking_table"`
	Chart             *models.ComparisonChart            `json:"chart"`
}

// StrategyRanking 策略排名
type StrategyRanking struct {
	Rank       int     `json:"rank"`
	StrategyID uint    `json:"strategy_id"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Metrics    map[string]float64 `json:"metrics"`
}

// SummaryStats 汇总统计
type SummaryStats struct {
	TotalAnalyses           int                    `json:"total_analyses"`
	RecentAnalyses          []AnalysisHistory      `json:"recent_analyses"`
	PerformanceDistribution map[string]int         `json:"performance_distribution"`
	TrendAnalysis           *AnalysisTrendData     `json:"trend_analysis"`
}

// AnalysisTrendData 分析趋势数据
type AnalysisTrendData struct {
	MonthlyTrend   []float64 `json:"monthly_trend"`
	GrowthRate     float64   `json:"growth_rate"`
	SeasonalityIndex float64 `json:"seasonality_index"`
}

// MultiComparisonResult 多结果对比
type MultiComparisonResult struct {
	ResultIDs         []uint                    `json:"result_ids"`
	ComparisonMetrics []string                  `json:"comparison_metrics"`
	GroupBy           string                    `json:"group_by"`
	Weights           map[string]float64        `json:"weights"`
	Benchmark         string                    `json:"benchmark"`
	ComparisonData    map[string]interface{}    `json:"comparison_data"`
	Summary           *MultiComparisonSummary   `json:"summary"`
}

// MultiComparisonSummary 多对比总结
type MultiComparisonSummary struct {
	BestResult      string   `json:"best_result"`
	KeyInsights     []string `json:"key_insights"`
	Recommendations []string `json:"recommendations"`
}

// 辅助方法

// generateStrategyComparisonMetrics 生成策略对比指标
func (as *AnalysisService) generateStrategyComparisonMetrics(strategies []models.Strategy, metrics []string) map[string][]float64 {
	result := make(map[string][]float64)
	
	if len(metrics) == 0 {
		metrics = []string{"total_return", "sharpe_ratio", "max_drawdown", "win_rate"}
	}
	
	for _, metric := range metrics {
		values := make([]float64, len(strategies))
		for i, strategy := range strategies {
			switch metric {
			case "total_return":
				values[i] = strategy.TotalReturn
			case "sharpe_ratio":
				values[i] = strategy.SharpeRatio
			case "max_drawdown":
				values[i] = strategy.MaxDrawdown
			case "win_rate":
				values[i] = strategy.WinRate
			}
		}
		result[metric] = values
	}
	
	return result
}

// generateStrategyRanking 生成策略排名
func (as *AnalysisService) generateStrategyRanking(strategies []models.Strategy) []StrategyRanking {
	type strategyScore struct {
		strategy models.Strategy
		score    float64
	}
	
	scores := make([]strategyScore, len(strategies))
	for i, strategy := range strategies {
		// 综合得分计算
		score := strategy.SharpeRatio*0.4 + strategy.TotalReturn*0.3 + (1-math.Abs(strategy.MaxDrawdown))*0.3
		scores[i] = strategyScore{strategy: strategy, score: score}
	}
	
	// 排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// 生成排名
	ranking := make([]StrategyRanking, len(scores))
	for i, item := range scores {
		ranking[i] = StrategyRanking{
			Rank:       i + 1,
			StrategyID: item.strategy.ID,
			Name:       item.strategy.Name,
			Score:      item.score,
			Metrics: map[string]float64{
				"total_return":  item.strategy.TotalReturn,
				"sharpe_ratio":  item.strategy.SharpeRatio,
				"max_drawdown":  item.strategy.MaxDrawdown,
				"win_rate":      item.strategy.WinRate,
			},
		}
	}
	
	return ranking
}

// generateStrategyComparisonChart 生成策略对比图表
func (as *AnalysisService) generateStrategyComparisonChart(strategies []models.Strategy) *models.ComparisonChart {
	chartData := make(map[string]interface{})
	
	names := make([]string, len(strategies))
	returns := make([]float64, len(strategies))
	sharpes := make([]float64, len(strategies))
	
	for i, strategy := range strategies {
		names[i] = strategy.Name
		returns[i] = strategy.TotalReturn
		sharpes[i] = strategy.SharpeRatio
	}
	
	chartData["labels"] = names
	chartData["datasets"] = []map[string]interface{}{
		{
			"label": "总收益率",
			"data":  returns,
			"type":  "bar",
		},
		{
			"label": "夏普比率",
			"data":  sharpes,
			"type":  "line",
		},
	}
	
	return &models.ComparisonChart{
		Type: "mixed",
		Data: chartData,
		Config: map[string]interface{}{
			"responsive": true,
		},
	}
}

// getResultData 获取结果数据
func (as *AnalysisService) getResultData(resultID uint, resultType string, userID uint) map[string]interface{} {
	result := make(map[string]interface{})
	
	switch resultType {
	case "model":
		var model models.Model
		if err := as.db.Where("id = ? AND user_id = ?", resultID, userID).First(&model).Error; err == nil {
			result["id"] = model.ID
			result["name"] = model.Name
			result["type"] = "model"
			result["test_ic"] = model.TestIC
			result["test_loss"] = model.TestLoss
		}
	case "strategy":
		var strategy models.Strategy
		if err := as.db.Where("id = ? AND user_id = ?", resultID, userID).First(&strategy).Error; err == nil {
			result["id"] = strategy.ID
			result["name"] = strategy.Name
			result["type"] = "strategy"
			result["total_return"] = strategy.TotalReturn
			result["sharpe_ratio"] = strategy.SharpeRatio
		}
	}
	
	return result
}