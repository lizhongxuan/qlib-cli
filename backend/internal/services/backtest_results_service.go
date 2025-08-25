package services

import (
	"fmt"
	"math"
	"time"

	"qlib-backend/internal/models"
	"qlib-backend/internal/utils"

	"gorm.io/gorm"
)

// BacktestResultsService 回测结果服务
type BacktestResultsService struct {
	db             *gorm.DB
	chartGenerator *utils.ChartGenerator
}

// DetailedBacktestResult 详细回测结果
type DetailedBacktestResult struct {
	ResultID           uint                         `json:"result_id"`
	StrategyID         uint                         `json:"strategy_id"`
	StrategyName       string                       `json:"strategy_name"`
	BacktestPeriod     DateRange                    `json:"backtest_period"`
	PerformanceMetrics *PerformanceMetrics          `json:"performance_metrics"`
	RiskMetrics        *RiskMetrics                 `json:"risk_metrics"`
	TradeAnalysis      *TradeAnalysis               `json:"trade_analysis"`
	TimeSeriesData     *TimeSeriesAnalysis          `json:"time_series_data"`
	PositionAnalysis   *PositionAnalysis            `json:"position_analysis"`
	SectorAnalysis     *SectorAnalysis              `json:"sector_analysis"`
	PeriodAnalysis     *PeriodAnalysis              `json:"period_analysis"`
	Benchmarks         []models.BenchmarkComparison `json:"benchmarks"`
	Charts             []ChartData                  `json:"charts"`
}

// DateRange 日期范围
type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Days      int    `json:"days"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TotalReturn      float64 `json:"total_return"`
	AnnualizedReturn float64 `json:"annualized_return"`
	Volatility       float64 `json:"volatility"`
	SharpeRatio      float64 `json:"sharpe_ratio"`
	SortinoRatio     float64 `json:"sortino_ratio"`
	CalmarRatio      float64 `json:"calmar_ratio"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	WinRate          float64 `json:"win_rate"`
	ProfitLossRatio  float64 `json:"profit_loss_ratio"`
	ExpectedReturn   float64 `json:"expected_return"`
	ReturnStdev      float64 `json:"return_stdev"`
	Beta             float64 `json:"beta"`
	Alpha            float64 `json:"alpha"`
	InformationRatio float64 `json:"information_ratio"`
	TrackingError    float64 `json:"tracking_error"`
}

// RiskMetrics 风险指标
type RiskMetrics struct {
	VaR95               float64       `json:"var_95"`
	VaR99               float64       `json:"var_99"`
	CVaR95              float64       `json:"cvar_95"`
	CVaR99              float64       `json:"cvar_99"`
	MaxDrawdown         float64       `json:"max_drawdown"`
	MaxDrawdownDuration int           `json:"max_drawdown_duration"`
	DownsideDeviation   float64       `json:"downside_deviation"`
	UpsideRatio         float64       `json:"upside_ratio"`
	DownsideRatio       float64       `json:"downside_ratio"`
	SkewkurtosisRisk    *SkewKurtosis `json:"skew_kurtosis_risk"`
}

// SkewKurtosis 偏度和峰度
type SkewKurtosis struct {
	Skewness float64 `json:"skewness"`
	Kurtosis float64 `json:"kurtosis"`
}

// TradeAnalysis 交易分析
type TradeAnalysis struct {
	TotalTrades        int     `json:"total_trades"`
	WinningTrades      int     `json:"winning_trades"`
	LosingTrades       int     `json:"losing_trades"`
	WinRate            float64 `json:"win_rate"`
	AverageWin         float64 `json:"average_win"`
	AverageLoss        float64 `json:"average_loss"`
	ProfitFactor       float64 `json:"profit_factor"`
	LargestWin         float64 `json:"largest_win"`
	LargestLoss        float64 `json:"largest_loss"`
	AverageTradeReturn float64 `json:"average_trade_return"`
	TradingFrequency   float64 `json:"trading_frequency"`
	Turnover           float64 `json:"turnover"`
}

// TimeSeriesAnalysis 时间序列分析
type TimeSeriesAnalysis struct {
	Dates               []string  `json:"dates"`
	PortfolioReturns    []float64 `json:"portfolio_returns"`
	CumulativeReturns   []float64 `json:"cumulative_returns"`
	BenchmarkReturns    []float64 `json:"benchmark_returns"`
	BenchmarkCumulative []float64 `json:"benchmark_cumulative"`
	ExcessReturns       []float64 `json:"excess_returns"`
	Drawdowns           []float64 `json:"drawdowns"`
	RollingVolatility   []float64 `json:"rolling_volatility"`
	RollingSharpe       []float64 `json:"rolling_sharpe"`
	PortfolioValue      []float64 `json:"portfolio_value"`
}

// PositionAnalysis 持仓分析
type PositionAnalysis struct {
	AveragePositions  int                   `json:"average_positions"`
	MaxPositions      int                   `json:"max_positions"`
	MinPositions      int                   `json:"min_positions"`
	PositionSizing    *PositionSizing       `json:"position_sizing"`
	TopHoldings       []HoldingInfo         `json:"top_holdings"`
	SectorExposure    map[string]float64    `json:"sector_exposure"`
	ConcentrationRisk *ConcentrationMetrics `json:"concentration_risk"`
}

// PositionSizing 仓位大小分析
type PositionSizing struct {
	AverageWeight float64 `json:"average_weight"`
	MaxWeight     float64 `json:"max_weight"`
	MinWeight     float64 `json:"min_weight"`
	WeightStdDev  float64 `json:"weight_std_dev"`
}

// HoldingInfo 持仓信息
type HoldingInfo struct {
	Symbol       string  `json:"symbol"`
	Weight       float64 `json:"weight"`
	Return       float64 `json:"return"`
	Contribution float64 `json:"contribution"`
	HoldingDays  int     `json:"holding_days"`
}

// ConcentrationMetrics 集中度指标
type ConcentrationMetrics struct {
	HerfindahlIndex    float64 `json:"herfindahl_index"`
	Top5Concentration  float64 `json:"top_5_concentration"`
	Top10Concentration float64 `json:"top_10_concentration"`
	EffectiveStocks    float64 `json:"effective_stocks"`
}

// SectorAnalysis 行业分析
type SectorAnalysis struct {
	SectorReturns         map[string]float64 `json:"sector_returns"`
	SectorWeights         map[string]float64 `json:"sector_weights"`
	SectorContribution    map[string]float64 `json:"sector_contribution"`
	BestPerformingSector  string             `json:"best_performing_sector"`
	WorstPerformingSector string             `json:"worst_performing_sector"`
}

// PeriodAnalysis 分期间分析
type PeriodAnalysis struct {
	MonthlyReturns     map[string]float64  `json:"monthly_returns"`
	QuarterlyReturns   map[string]float64  `json:"quarterly_returns"`
	YearlyReturns      map[string]float64  `json:"yearly_returns"`
	BestMonth          PeriodInfo          `json:"best_month"`
	WorstMonth         PeriodInfo          `json:"worst_month"`
	BestQuarter        PeriodInfo          `json:"best_quarter"`
	WorstQuarter       PeriodInfo          `json:"worst_quarter"`
	ConsistencyMetrics *ConsistencyMetrics `json:"consistency_metrics"`
}

// PeriodInfo 期间信息
type PeriodInfo struct {
	Period string  `json:"period"`
	Return float64 `json:"return"`
}

// ConsistencyMetrics 一致性指标
type ConsistencyMetrics struct {
	MonthlyWinRate   float64 `json:"monthly_win_rate"`
	QuarterlyWinRate float64 `json:"quarterly_win_rate"`
	YearlyWinRate    float64 `json:"yearly_win_rate"`
	ConsistencyScore float64 `json:"consistency_score"`
}

// ChartData 图表数据
type ChartData struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"` // line, bar, pie, heatmap, scatter
	Title  string                 `json:"title"`
	Data   map[string]interface{} `json:"data"`
	Config map[string]interface{} `json:"config"`
}

// ChartType 图表类型
type ChartType string

const (
	ChartTypeCumulativeReturns  ChartType = "cumulative_returns"
	ChartTypeDrawdowns          ChartType = "drawdowns"
	ChartTypeRollingMetrics     ChartType = "rolling_metrics"
	ChartTypePositionWeights    ChartType = "position_weights"
	ChartTypeSectorExposure     ChartType = "sector_exposure"
	ChartTypeMonthlyReturns     ChartType = "monthly_returns"
	ChartTypeReturnDistribution ChartType = "return_distribution"
	ChartTypeRiskReturn         ChartType = "risk_return"
)

// NewBacktestResultsService 创建新的回测结果服务
func NewBacktestResultsService(db *gorm.DB) *BacktestResultsService {
	chartGenerator := utils.NewChartGenerator()

	return &BacktestResultsService{
		db:             db,
		chartGenerator: chartGenerator,
	}
}

// GetDetailedResultsOptions 详细结果查询选项
type GetDetailedResultsOptions struct {
	IncludeTradeDetails    bool   `json:"include_trade_details"`
	IncludePositionDetails bool   `json:"include_position_details"`
	IncludeRiskMetrics     bool   `json:"include_risk_metrics"`
	TimeRange              string `json:"time_range"`
}

// GetDetailedResults 获取详细回测结果
func (brs *BacktestResultsService) GetDetailedResults(resultID uint, userID uint) (*DetailedBacktestResult, error) {
	// 使用默认选项调用扩展方法
	options := GetDetailedResultsOptions{
		IncludeTradeDetails:    true,
		IncludePositionDetails: true,
		IncludeRiskMetrics:     true,
		TimeRange:              "",
	}
	return brs.GetDetailedResultsWithOptions(resultID, userID, options)
}

// GetDetailedResultsWithOptions 获取详细回测结果（带选项）
func (brs *BacktestResultsService) GetDetailedResultsWithOptions(resultID uint, userID uint, options GetDetailedResultsOptions) (*DetailedBacktestResult, error) {
	// 验证权限并获取策略信息
	var strategy models.Strategy
	if err := brs.db.Where("id = ? AND user_id = ?", resultID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("回测结果不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取回测结果失败: %v", err)
	}

	result := &DetailedBacktestResult{
		ResultID:     resultID,
		StrategyID:   strategy.ID,
		StrategyName: strategy.Name,
	}

	// 解析回测时间范围
	result.BacktestPeriod = DateRange{
		StartDate: strategy.BacktestStart,
		EndDate:   strategy.BacktestEnd,
		Days:      brs.calculateTradingDays(strategy.BacktestStart, strategy.BacktestEnd),
	}

	// 计算性能指标
	performance, err := brs.calculatePerformanceMetrics(strategy)
	if err != nil {
		return nil, fmt.Errorf("计算性能指标失败: %v", err)
	}
	result.PerformanceMetrics = performance

	// 根据选项计算风险指标
	if options.IncludeRiskMetrics {
		risk, err := brs.calculateRiskMetrics(strategy)
		if err != nil {
			return nil, fmt.Errorf("计算风险指标失败: %v", err)
		}
		result.RiskMetrics = risk
	}

	// 根据选项进行交易分析
	if options.IncludeTradeDetails {
		tradeAnalysis, err := brs.calculateTradeAnalysis(strategy)
		if err != nil {
			return nil, fmt.Errorf("计算交易分析失败: %v", err)
		}
		result.TradeAnalysis = tradeAnalysis
	}

	// 时间序列数据 - 根据TimeRange过滤
	timeSeries, err := brs.generateTimeSeriesDataWithRange(strategy, options.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("生成时间序列数据失败: %v", err)
	}
	result.TimeSeriesData = timeSeries

	// 根据选项进行持仓分析
	if options.IncludePositionDetails {
		positionAnalysis, err := brs.calculatePositionAnalysis(strategy)
		if err != nil {
			return nil, fmt.Errorf("计算持仓分析失败: %v", err)
		}
		result.PositionAnalysis = positionAnalysis
	}

	// 行业分析
	sectorAnalysis, err := brs.calculateSectorAnalysis(strategy)
	if err != nil {
		return nil, fmt.Errorf("计算行业分析失败: %v", err)
	}
	result.SectorAnalysis = sectorAnalysis

	// 分期间分析
	periodAnalysis, err := brs.calculatePeriodAnalysis(strategy)
	if err != nil {
		return nil, fmt.Errorf("计算分期间分析失败: %v", err)
	}
	result.PeriodAnalysis = periodAnalysis

	// 基准对比
	benchmarks, err := brs.calculateBenchmarkComparison(strategy)
	if err != nil {
		return nil, fmt.Errorf("计算基准对比失败: %v", err)
	}
	result.Benchmarks = benchmarks

	// 生成图表
	charts, err := brs.generateCharts(strategy, result)
	if err != nil {
		return nil, fmt.Errorf("生成图表失败: %v", err)
	}
	result.Charts = charts

	return result, nil
}

// GetChartDataOptions 图表数据查询选项
type GetChartDataOptions struct {
	TimeRange  string   `json:"time_range"`
	Resolution string   `json:"resolution"`
	Benchmark  string   `json:"benchmark"`
	Indicators []string `json:"indicators"`
}

// GetChartData 获取图表数据
func (brs *BacktestResultsService) GetChartData(resultID uint, chartType ChartType, userID uint) (*ChartData, error) {
	// 使用默认选项调用扩展方法
	options := GetChartDataOptions{
		Resolution: "daily",
		Indicators: []string{},
	}
	return brs.GetChartDataWithOptions(resultID, chartType, userID, options)
}

// GetChartDataWithOptions 获取图表数据（带选项）
func (brs *BacktestResultsService) GetChartDataWithOptions(resultID uint, chartType ChartType, userID uint, options GetChartDataOptions) (*ChartData, error) {
	// 验证权限
	var strategy models.Strategy
	if err := brs.db.Where("id = ? AND user_id = ?", resultID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("回测结果不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取回测结果失败: %v", err)
	}

	// TODO: 根据选项参数生成不同的图表
	// 可以基于TimeRange, Resolution, Benchmark, Indicators调整图表数据

	switch chartType {
	case ChartTypeCumulativeReturns:
		return brs.generateCumulativeReturnsChartWithOptions(strategy, options)
	case ChartTypeDrawdowns:
		return brs.generateDrawdownsChartWithOptions(strategy, options)
	case ChartTypeRollingMetrics:
		return brs.generateRollingMetricsChartWithOptions(strategy, options)
	case ChartTypePositionWeights:
		return brs.generatePositionWeightsChart(strategy)
	case ChartTypeSectorExposure:
		return brs.generateSectorExposureChart(strategy)
	case ChartTypeMonthlyReturns:
		return brs.generateMonthlyReturnsChart(strategy)
	case ChartTypeReturnDistribution:
		return brs.generateReturnDistributionChart(strategy)
	case ChartTypeRiskReturn:
		return brs.generateRiskReturnChart(strategy)
	default:
		return nil, fmt.Errorf("不支持的图表类型: %s", chartType)
	}
}

// generateCumulativeReturnsChartWithOptions 生成带选项的累积收益图表
func (brs *BacktestResultsService) generateCumulativeReturnsChartWithOptions(strategy models.Strategy, options GetChartDataOptions) (*ChartData, error) {
	// 生成时间序列数据，根据TimeRange过滤
	timeSeries, err := brs.generateTimeSeriesDataWithRange(strategy, options.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("生成时间序列数据失败: %v", err)
	}

	// 根据Resolution调整数据粒度
	dates, cumulativeReturns, benchmarkCumulative := brs.resampleData(
		timeSeries.Dates,
		timeSeries.CumulativeReturns,
		timeSeries.BenchmarkCumulative,
		options.Resolution,
	)

	// 构建图表数据
	chartData := &ChartData{
		ID:    "cumulative_returns",
		Type:  "line",
		Title: "累积收益曲线",
		Data: map[string]interface{}{
			"dates":    dates,
			"strategy": cumulativeReturns,
		},
		Config: map[string]interface{}{
			"yAxis": map[string]interface{}{
				"title":  "累积收益率",
				"format": "percentage",
			},
			"xAxis": map[string]interface{}{
				"title":  "时间",
				"format": "date",
			},
		},
	}

	// 如果指定了基准，添加基准数据
	if options.Benchmark != "" {
		chartData.Data["benchmark"] = benchmarkCumulative
		chartData.Config["legend"] = []string{"策略", options.Benchmark}
	}

	// 添加额外指标
	if len(options.Indicators) > 0 {
		for _, indicator := range options.Indicators {
			switch indicator {
			case "drawdown":
				chartData.Data["drawdown"] = timeSeries.Drawdowns
			case "volatility":
				chartData.Data["volatility"] = timeSeries.RollingVolatility
			}
		}
	}

	return chartData, nil
}

// generateDrawdownsChartWithOptions 生成带选项的回撤图表
func (brs *BacktestResultsService) generateDrawdownsChartWithOptions(strategy models.Strategy, options GetChartDataOptions) (*ChartData, error) {
	// 生成时间序列数据，根据TimeRange过滤
	timeSeries, err := brs.generateTimeSeriesDataWithRange(strategy, options.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("生成时间序列数据失败: %v", err)
	}

	// 根据Resolution调整数据粒度
	dates, drawdowns, _ := brs.resampleData(
		timeSeries.Dates,
		timeSeries.Drawdowns,
		nil,
		options.Resolution,
	)

	// 构建图表数据
	chartData := &ChartData{
		ID:    "drawdowns",
		Type:  "area",
		Title: "回撤分析",
		Data: map[string]interface{}{
			"dates":     dates,
			"drawdowns": drawdowns,
		},
		Config: map[string]interface{}{
			"yAxis": map[string]interface{}{
				"title":  "回撤幅度",
				"format": "percentage",
				"min":    -1,
				"max":    0,
			},
			"xAxis": map[string]interface{}{
				"title":  "时间",
				"format": "date",
			},
			"fill":  true,
			"color": "#ff4d4f",
		},
	}

	return chartData, nil
}

// generateRollingMetricsChartWithOptions 生成带选项的滚动指标图表
func (brs *BacktestResultsService) generateRollingMetricsChartWithOptions(strategy models.Strategy, options GetChartDataOptions) (*ChartData, error) {
	// 生成时间序列数据，根据TimeRange过滤
	timeSeries, err := brs.generateTimeSeriesDataWithRange(strategy, options.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("生成时间序列数据失败: %v", err)
	}

	// 根据Resolution调整数据粒度
	dates, volatility, sharpe := brs.resampleData(
		timeSeries.Dates,
		timeSeries.RollingVolatility,
		timeSeries.RollingSharpe,
		options.Resolution,
	)

	// 构建图表数据
	chartData := &ChartData{
		ID:    "rolling_metrics",
		Type:  "line",
		Title: "滚动指标",
		Data: map[string]interface{}{
			"dates": dates,
		},
		Config: map[string]interface{}{
			"yAxis": map[string]interface{}{
				"title": "指标值",
			},
			"xAxis": map[string]interface{}{
				"title":  "时间",
				"format": "date",
			},
		},
	}

	// 根据指定的indicators添加不同指标
	if len(options.Indicators) > 0 {
		for _, indicator := range options.Indicators {
			switch indicator {
			case "volatility":
				chartData.Data["volatility"] = volatility
			case "sharpe":
				chartData.Data["sharpe"] = sharpe
			}
		}
	} else {
		// 默认显示波动率和夏普比率
		chartData.Data["volatility"] = volatility
		chartData.Data["sharpe"] = sharpe
	}

	return chartData, nil
}

// ExportBacktestReportExtended 导出回测报告（扩展版本）
func (brs *BacktestResultsService) ExportBacktestReportExtended(req models.BacktestReportExportRequestExtended, userID uint) (string, error) {
	// 验证所有结果ID的权限
	for _, resultID := range req.ResultIDs {
		var strategy models.Strategy
		if err := brs.db.Where("id = ? AND user_id = ?", resultID, userID).First(&strategy).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return "", fmt.Errorf("回测结果 %d 不存在或无权限访问", resultID)
			}
			return "", fmt.Errorf("验证回测结果权限失败: %v", err)
		}
	}

	// 根据报告类型生成不同的任务ID
	taskID := fmt.Sprintf("export_%s_%s_%d", req.ReportType, req.Format, time.Now().Unix())

	// TODO: 实现真实的导出逻辑
	// 1. 根据ReportType决定导出内容 (summary, detailed, comparison)
	// 2. 根据Format决定导出格式 (pdf, excel, html)
	// 3. 使用Template应用模板
	// 4. 根据Sections选择包含的部分
	// 5. 根据IncludeCharts决定是否包含图表
	// 6. 使用Benchmark作为对比基准
	// 7. 根据Language设置语言

	return taskID, nil
}

// ExportBacktestReport 导出回测报告
func (brs *BacktestResultsService) ExportBacktestReport(req models.BacktestReportExportRequest, userID uint) (string, error) {
	// 获取详细结果
	result, err := brs.GetDetailedResults(req.ResultID, userID)
	if err != nil {
		return "", fmt.Errorf("获取回测结果失败: %v", err)
	}

	// 根据格式导出
	switch req.Format {
	case "pdf":
		return brs.exportToPDF(result, req)
	case "excel":
		return brs.exportToExcel(result, req)
	case "json":
		return brs.exportToJSON(result, req)
	default:
		return "", fmt.Errorf("不支持的导出格式: %s", req.Format)
	}
}

// 内部计算方法

// calculatePerformanceMetrics 计算性能指标
func (brs *BacktestResultsService) calculatePerformanceMetrics(strategy models.Strategy) (*PerformanceMetrics, error) {
	return &PerformanceMetrics{
		TotalReturn:      strategy.TotalReturn,
		AnnualizedReturn: strategy.AnnualReturn,
		Volatility:       strategy.Volatility,
		SharpeRatio:      strategy.SharpeRatio,
		SortinoRatio:     strategy.SharpeRatio * 1.2, // 简化计算
		CalmarRatio:      strategy.AnnualReturn / math.Abs(strategy.MaxDrawdown),
		MaxDrawdown:      strategy.MaxDrawdown,
		WinRate:          strategy.WinRate,
		ProfitLossRatio:  2.1, // 模拟数据
		ExpectedReturn:   strategy.AnnualReturn / 252,
		ReturnStdev:      strategy.Volatility / math.Sqrt(252),
		Beta:             1.0,
		Alpha:            strategy.ExcessReturn,
		InformationRatio: strategy.SharpeRatio * 0.8,
		TrackingError:    0.05,
	}, nil
}

// calculateRiskMetrics 计算风险指标
func (brs *BacktestResultsService) calculateRiskMetrics(strategy models.Strategy) (*RiskMetrics, error) {
	return &RiskMetrics{
		VaR95:               -0.05,
		VaR99:               -0.08,
		CVaR95:              -0.07,
		CVaR99:              -0.10,
		MaxDrawdown:         strategy.MaxDrawdown,
		MaxDrawdownDuration: 45, // 模拟数据
		DownsideDeviation:   strategy.Volatility * 0.8,
		UpsideRatio:         1.1,
		DownsideRatio:       0.9,
		SkewkurtosisRisk: &SkewKurtosis{
			Skewness: -0.2,
			Kurtosis: 3.5,
		},
	}, nil
}

// calculateTradeAnalysis 计算交易分析
func (brs *BacktestResultsService) calculateTradeAnalysis(strategy models.Strategy) (*TradeAnalysis, error) {
	// 模拟交易数据
	totalTrades := 1200
	winningTrades := int(float64(totalTrades) * strategy.WinRate)
	losingTrades := totalTrades - winningTrades

	return &TradeAnalysis{
		TotalTrades:        totalTrades,
		WinningTrades:      winningTrades,
		LosingTrades:       losingTrades,
		WinRate:            strategy.WinRate,
		AverageWin:         0.035,
		AverageLoss:        -0.018,
		ProfitFactor:       2.3,
		LargestWin:         0.152,
		LargestLoss:        -0.087,
		AverageTradeReturn: strategy.TotalReturn / float64(totalTrades),
		TradingFrequency:   252.0 / float64(totalTrades), // 每年交易频率
		Turnover:           3.2,
	}, nil
}

// generateTimeSeriesData 生成时间序列数据
// generateTimeSeriesDataWithRange 生成带时间范围的时间序列数据
func (brs *BacktestResultsService) generateTimeSeriesDataWithRange(strategy models.Strategy, timeRange string) (*TimeSeriesAnalysis, error) {
	// 如果没有指定时间范围，使用完整范围
	if timeRange == "" {
		return brs.generateTimeSeriesData(strategy)
	}

	// TODO: 解析timeRange参数，如"3m", "6m", "1y"等，并据此过滤数据
	// 当前简化处理，返回完整数据
	return brs.generateTimeSeriesData(strategy)
}

func (brs *BacktestResultsService) generateTimeSeriesData(strategy models.Strategy) (*TimeSeriesAnalysis, error) {
	// 模拟时间序列数据
	days := brs.calculateTradingDays(strategy.BacktestStart, strategy.BacktestEnd)

	dates := make([]string, days)
	portfolioReturns := make([]float64, days)
	cumulativeReturns := make([]float64, days)
	benchmarkReturns := make([]float64, days)
	benchmarkCumulative := make([]float64, days)
	excessReturns := make([]float64, days)
	drawdowns := make([]float64, days)
	portfolioValue := make([]float64, days)

	startDate, _ := time.Parse("2006-01-02", strategy.BacktestStart)
	dailyReturn := strategy.AnnualReturn / float64(days)
	benchmarkDailyReturn := 0.08 / float64(days) // 假设基准年收益8%

	portfolioVal := 100000.0 // 初始资金
	cumReturn := 0.0
	benchmarkCumReturn := 0.0
	maxValue := portfolioVal

	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		dates[i] = date.Format("2006-01-02")

		// 模拟日收益率（添加一些波动）
		dailyRet := dailyReturn + 0.01*math.Sin(float64(i)/20) + 0.005*(math.Mod(float64(i), 2)-0.5)
		benchmarkRet := benchmarkDailyReturn + 0.005*math.Sin(float64(i)/25)

		portfolioReturns[i] = dailyRet
		benchmarkReturns[i] = benchmarkRet
		excessReturns[i] = dailyRet - benchmarkRet

		cumReturn += dailyRet
		benchmarkCumReturn += benchmarkRet

		cumulativeReturns[i] = cumReturn
		benchmarkCumulative[i] = benchmarkCumReturn

		portfolioVal *= (1 + dailyRet)
		portfolioValue[i] = portfolioVal

		if portfolioVal > maxValue {
			maxValue = portfolioVal
		}
		drawdowns[i] = (portfolioVal - maxValue) / maxValue
	}

	// 计算滚动指标
	rollingVolatility := brs.calculateRollingVolatility(portfolioReturns, 20)
	rollingSharpe := brs.calculateRollingSharpe(portfolioReturns, 60)

	return &TimeSeriesAnalysis{
		Dates:               dates,
		PortfolioReturns:    portfolioReturns,
		CumulativeReturns:   cumulativeReturns,
		BenchmarkReturns:    benchmarkReturns,
		BenchmarkCumulative: benchmarkCumulative,
		ExcessReturns:       excessReturns,
		Drawdowns:           drawdowns,
		RollingVolatility:   rollingVolatility,
		RollingSharpe:       rollingSharpe,
		PortfolioValue:      portfolioValue,
	}, nil
}

// calculatePositionAnalysis 计算持仓分析
func (brs *BacktestResultsService) calculatePositionAnalysis(strategy models.Strategy) (*PositionAnalysis, error) {
	// 模拟持仓数据
	topHoldings := []HoldingInfo{
		{Symbol: "000001.SZ", Weight: 0.05, Return: 0.12, Contribution: 0.006, HoldingDays: 180},
		{Symbol: "000002.SZ", Weight: 0.04, Return: 0.08, Contribution: 0.0032, HoldingDays: 150},
		{Symbol: "600000.SH", Weight: 0.045, Return: 0.15, Contribution: 0.0067, HoldingDays: 200},
		{Symbol: "600036.SH", Weight: 0.038, Return: 0.10, Contribution: 0.0038, HoldingDays: 120},
		{Symbol: "000858.SZ", Weight: 0.042, Return: 0.09, Contribution: 0.0037, HoldingDays: 165},
	}

	sectorExposure := map[string]float64{
		"金融":   0.35,
		"科技":   0.25,
		"消费":   0.20,
		"医药":   0.12,
		"制造业": 0.08,
	}

	return &PositionAnalysis{
		AveragePositions: 45,
		MaxPositions:     60,
		MinPositions:     30,
		PositionSizing: &PositionSizing{
			AverageWeight: 0.022,
			MaxWeight:     0.05,
			MinWeight:     0.01,
			WeightStdDev:  0.008,
		},
		TopHoldings:    topHoldings,
		SectorExposure: sectorExposure,
		ConcentrationRisk: &ConcentrationMetrics{
			HerfindahlIndex:    0.12,
			Top5Concentration:  0.215,
			Top10Concentration: 0.38,
			EffectiveStocks:    28.5,
		},
	}, nil
}

// calculateSectorAnalysis 计算行业分析
func (brs *BacktestResultsService) calculateSectorAnalysis(strategy models.Strategy) (*SectorAnalysis, error) {
	sectorReturns := map[string]float64{
		"科技":   0.18,
		"金融":   0.12,
		"消费":   0.15,
		"医药":   0.22,
		"制造业": 0.08,
	}

	sectorWeights := map[string]float64{
		"科技":   0.25,
		"金融":   0.35,
		"消费":   0.20,
		"医药":   0.12,
		"制造业": 0.08,
	}

	sectorContribution := make(map[string]float64)
	for sector, weight := range sectorWeights {
		if ret, exists := sectorReturns[sector]; exists {
			sectorContribution[sector] = weight * ret
		}
	}

	// 找出最佳和最差行业
	bestSector, worstSector := "", ""
	bestReturn, worstReturn := -1.0, 2.0

	for sector, ret := range sectorReturns {
		if ret > bestReturn {
			bestReturn = ret
			bestSector = sector
		}
		if ret < worstReturn {
			worstReturn = ret
			worstSector = sector
		}
	}

	return &SectorAnalysis{
		SectorReturns:         sectorReturns,
		SectorWeights:         sectorWeights,
		SectorContribution:    sectorContribution,
		BestPerformingSector:  bestSector,
		WorstPerformingSector: worstSector,
	}, nil
}

// calculatePeriodAnalysis 计算分期间分析
func (brs *BacktestResultsService) calculatePeriodAnalysis(strategy models.Strategy) (*PeriodAnalysis, error) {
	// 模拟月度收益
	monthlyReturns := map[string]float64{
		"2023-01": 0.035,
		"2023-02": 0.028,
		"2023-03": 0.042,
		"2023-04": -0.015,
		"2023-05": 0.038,
		"2023-06": 0.022,
		"2023-07": 0.031,
		"2023-08": -0.008,
		"2023-09": 0.025,
		"2023-10": 0.045,
		"2023-11": 0.018,
		"2023-12": 0.033,
	}

	// 计算季度收益
	quarterlyReturns := map[string]float64{
		"2023-Q1": 0.108,
		"2023-Q2": 0.045,
		"2023-Q3": 0.048,
		"2023-Q4": 0.099,
	}

	yearlyReturns := map[string]float64{
		"2023": strategy.AnnualReturn,
	}

	// 找出最佳和最差月份
	bestMonth, worstMonth := PeriodInfo{}, PeriodInfo{}
	bestReturn, worstReturn := -1.0, 1.0

	for month, ret := range monthlyReturns {
		if ret > bestReturn {
			bestReturn = ret
			bestMonth = PeriodInfo{Period: month, Return: ret}
		}
		if ret < worstReturn {
			worstReturn = ret
			worstMonth = PeriodInfo{Period: month, Return: ret}
		}
	}

	// 计算一致性指标
	positiveMonths := 0
	for _, ret := range monthlyReturns {
		if ret > 0 {
			positiveMonths++
		}
	}

	return &PeriodAnalysis{
		MonthlyReturns:   monthlyReturns,
		QuarterlyReturns: quarterlyReturns,
		YearlyReturns:    yearlyReturns,
		BestMonth:        bestMonth,
		WorstMonth:       worstMonth,
		BestQuarter:      PeriodInfo{Period: "2023-Q1", Return: 0.108},
		WorstQuarter:     PeriodInfo{Period: "2023-Q2", Return: 0.045},
		ConsistencyMetrics: &ConsistencyMetrics{
			MonthlyWinRate:   float64(positiveMonths) / float64(len(monthlyReturns)),
			QuarterlyWinRate: 1.0,
			YearlyWinRate:    1.0,
			ConsistencyScore: 0.85,
		},
	}, nil
}

// calculateBenchmarkComparison 计算基准对比
func (brs *BacktestResultsService) calculateBenchmarkComparison(strategy models.Strategy) ([]models.BenchmarkComparison, error) {
	benchmarks := []models.BenchmarkComparison{
		{
			BenchmarkName:    "CSI300",
			ExcessReturn:     strategy.ExcessReturn,
			TrackingError:    0.05,
			InformationRatio: strategy.ExcessReturn / 0.05,
			ActiveReturn:     strategy.ExcessReturn,
			UpCapture:        1.1,
			DownCapture:      0.9,
			CorrelationCoeff: 0.85,
		},
		{
			BenchmarkName:    "CSI500",
			ExcessReturn:     strategy.AnnualReturn - 0.12,
			TrackingError:    0.08,
			InformationRatio: (strategy.AnnualReturn - 0.12) / 0.08,
			ActiveReturn:     strategy.AnnualReturn - 0.12,
			UpCapture:        1.05,
			DownCapture:      0.95,
			CorrelationCoeff: 0.75,
		},
	}

	return benchmarks, nil
}

// 辅助方法

// resampleData 根据分辨率重采样数据
func (brs *BacktestResultsService) resampleData(dates []string, data1, data2 []float64, resolution string) ([]string, []float64, []float64) {
	// 如果分辨率为daily或为空，直接返回原始数据
	if resolution == "" || resolution == "daily" {
		return dates, data1, data2
	}

	// 简化实现：基于resolution采样数据
	step := 1
	switch resolution {
	case "weekly":
		step = 5 // 每5个交易日采样一次
	case "monthly":
		step = 22 // 每22个交易日采样一次
	}

	if step == 1 {
		return dates, data1, data2
	}

	// 重采样
	resampledDates := make([]string, 0)
	resampledData1 := make([]float64, 0)
	resampledData2 := make([]float64, 0)

	for i := 0; i < len(dates); i += step {
		if i < len(dates) {
			resampledDates = append(resampledDates, dates[i])
			if i < len(data1) {
				resampledData1 = append(resampledData1, data1[i])
			}
			if data2 != nil && i < len(data2) {
				resampledData2 = append(resampledData2, data2[i])
			}
		}
	}

	return resampledDates, resampledData1, resampledData2
}

// calculateTradingDays 计算交易日天数
func (brs *BacktestResultsService) calculateTradingDays(startDate, endDate string) int {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	totalDays := int(end.Sub(start).Hours() / 24)
	return int(float64(totalDays) * 5 / 7) // 简化计算，假设5/7是交易日
}

// calculateRollingVolatility 计算滚动波动率
func (brs *BacktestResultsService) calculateRollingVolatility(returns []float64, window int) []float64 {
	volatility := make([]float64, len(returns))

	for i := 0; i < len(returns); i++ {
		if i < window-1 {
			volatility[i] = 0
			continue
		}

		windowReturns := returns[i-window+1 : i+1]
		mean := brs.calculateMean(windowReturns)
		variance := 0.0

		for _, ret := range windowReturns {
			variance += math.Pow(ret-mean, 2)
		}
		variance /= float64(len(windowReturns) - 1)
		volatility[i] = math.Sqrt(variance) * math.Sqrt(252) // 年化
	}

	return volatility
}

// calculateRollingSharpe 计算滚动夏普比率
func (brs *BacktestResultsService) calculateRollingSharpe(returns []float64, window int) []float64 {
	sharpe := make([]float64, len(returns))
	riskFreeRate := 0.03 / 252 // 假设无风险利率3%

	for i := 0; i < len(returns); i++ {
		if i < window-1 {
			sharpe[i] = 0
			continue
		}

		windowReturns := returns[i-window+1 : i+1]
		mean := brs.calculateMean(windowReturns)
		std := brs.calculateStd(windowReturns)

		if std > 0 {
			sharpe[i] = (mean - riskFreeRate) / std * math.Sqrt(252)
		}
	}

	return sharpe
}

// calculateMean 计算均值
func (brs *BacktestResultsService) calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStd 计算标准差
func (brs *BacktestResultsService) calculateStd(values []float64) float64 {
	mean := brs.calculateMean(values)
	variance := 0.0

	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values) - 1)
	return math.Sqrt(variance)
}

// 图表生成方法的占位符（具体实现在chart_generator.go中）

func (brs *BacktestResultsService) generateCharts(strategy models.Strategy, result *DetailedBacktestResult) ([]ChartData, error) {
	// 这里会调用chart_generator生成各种图表
	return []ChartData{}, nil
}

func (brs *BacktestResultsService) generateCumulativeReturnsChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "cumulative_returns",
		Type:  "line",
		Title: "累积收益曲线",
	}, nil
}

func (brs *BacktestResultsService) generateDrawdownsChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "drawdowns",
		Type:  "area",
		Title: "回撤分析",
	}, nil
}

func (brs *BacktestResultsService) generateRollingMetricsChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "rolling_metrics",
		Type:  "line",
		Title: "滚动指标",
	}, nil
}

func (brs *BacktestResultsService) generatePositionWeightsChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "position_weights",
		Type:  "bar",
		Title: "持仓权重",
	}, nil
}

func (brs *BacktestResultsService) generateSectorExposureChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "sector_exposure",
		Type:  "pie",
		Title: "行业暴露",
	}, nil
}

func (brs *BacktestResultsService) generateMonthlyReturnsChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "monthly_returns",
		Type:  "bar",
		Title: "月度收益",
	}, nil
}

func (brs *BacktestResultsService) generateReturnDistributionChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "return_distribution",
		Type:  "histogram",
		Title: "收益分布",
	}, nil
}

func (brs *BacktestResultsService) generateRiskReturnChart(strategy models.Strategy) (*ChartData, error) {
	return &ChartData{
		ID:    "risk_return",
		Type:  "scatter",
		Title: "风险收益散点图",
	}, nil
}

// 导出方法的占位符

func (brs *BacktestResultsService) exportToPDF(result *DetailedBacktestResult, req models.BacktestReportExportRequest) (string, error) {
	return "/tmp/reports/backtest_report.pdf", nil
}

func (brs *BacktestResultsService) exportToExcel(result *DetailedBacktestResult, req models.BacktestReportExportRequest) (string, error) {
	return "/tmp/reports/backtest_report.xlsx", nil
}

func (brs *BacktestResultsService) exportToJSON(result *DetailedBacktestResult, req models.BacktestReportExportRequest) (string, error) {
	return "/tmp/reports/backtest_report.json", nil
}
