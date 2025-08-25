package models

import "time"

// BenchmarkComparison 基准对比
type BenchmarkComparison struct {
	BenchmarkName     string  `json:"benchmark_name"`
	BenchmarkReturn   float64 `json:"benchmark_return"`
	StrategyReturn    float64 `json:"strategy_return"`
	ExcessReturn      float64 `json:"excess_return"`
	TrackingError     float64 `json:"tracking_error"`
	InformationRatio  float64 `json:"information_ratio"`
	ActiveReturn      float64 `json:"active_return"`
	UpCapture         float64 `json:"up_capture"`
	DownCapture       float64 `json:"down_capture"`
	CorrelationCoeff  float64 `json:"correlation_coefficient"`
	BetaCoeff         float64 `json:"beta_coeff"`
	AlphaCoeff        float64 `json:"alpha_coeff"`
}

// ModelComparisonRequest 模型对比请求
type ModelComparisonRequest struct {
	ModelIDs    []uint   `json:"model_ids" binding:"required"`
	Metrics     []string `json:"metrics"`
	TimeRange   string   `json:"time_range,omitempty"`
	Granularity string   `json:"granularity,omitempty"`
}

// ModelPerformance 模型性能
type ModelPerformance struct {
	ModelID    uint                   `json:"model_id"`
	ModelName  string                 `json:"model_name"`
	ModelType  string                 `json:"model_type"`
	Status     string                 `json:"status"`
	TrainedAt  time.Time             `json:"trained_at"`
	Metrics    map[string]float64     `json:"metrics"`
	TestIC     float64                `json:"test_ic"`
	TestLoss   float64                `json:"test_loss"`
	Stability  float64                `json:"stability"`
	Robustness float64                `json:"robustness"`
}

// ModelComparisonResult 模型对比结果
type ModelComparisonResult struct {
	Models          []ModelPerformance     `json:"models"`
	ComparisonChart *ComparisonChart       `json:"comparison_chart"`
	RankingTable    []ModelRanking         `json:"ranking_table"`
	StatisticalTest *StatisticalTestResult `json:"statistical_test"`
	Summary         *ComparisonSummary     `json:"summary"`
	ComparisonDate  time.Time              `json:"comparison_date"`
}


// ComparisonChart 对比图表
type ComparisonChart struct {
	Type   string                 `json:"type"`
	Data   map[string]interface{} `json:"data"`
	Config map[string]interface{} `json:"config"`
}

// ModelRanking 模型排名
type ModelRanking struct {
	Rank      int                `json:"rank"`
	ModelID   uint               `json:"model_id"`
	ModelName string             `json:"model_name"`
	Score     float64            `json:"score"`
	Metrics   map[string]float64 `json:"metrics"`
}

// StatisticalTestResult 统计测试结果
type StatisticalTestResult struct {
	TestType   string                 `json:"test_type"`
	PValue     float64                `json:"p_value"`
	Statistic  float64                `json:"statistic"`
	Result     string                 `json:"result"`
	Details    map[string]interface{} `json:"details"`
}

// ComparisonSummary 对比摘要
type ComparisonSummary struct {
	BestModel    uint                   `json:"best_model"`
	WorstModel   uint                   `json:"worst_model"`
	AvgScore     float64                `json:"avg_score"`
	ScoreRange   float64                `json:"score_range"`
	Highlights   []string               `json:"highlights"`
	Details      map[string]interface{} `json:"details"`
}

// BacktestReportExportRequest 回测报告导出请求
type BacktestReportExportRequest struct {
	ResultID uint     `json:"result_id" binding:"required"` // 兼容性保留
	Format   string   `json:"format" binding:"required"`    // pdf, excel, html
	Sections []string `json:"sections"`                     // 要包含的部分
	Language string   `json:"language"`                     // zh, en
}

// BacktestReportExportRequestExtended 扩展的回测报告导出请求
type BacktestReportExportRequestExtended struct {
	ResultIDs     []uint   `json:"result_ids" binding:"required"`
	ReportType    string   `json:"report_type" binding:"required"` // summary, detailed, comparison
	Format        string   `json:"format" binding:"required"`      // pdf, excel, html
	Template      string   `json:"template"`                       // 模板名称
	Sections      []string `json:"sections"`                       // 包含的部分
	IncludeCharts bool     `json:"include_charts"`                 // 是否包含图表
	Benchmark     string   `json:"benchmark"`                      // 基准指标
	Language      string   `json:"language"`                       // zh, en
}

// ChartData 图表数据 (从utils包移动到这里作为共享类型)
type ChartData struct {
	Type   string                 `json:"type"`   // line, bar, pie, heatmap等
	Title  string                 `json:"title"`
	Data   []interface{}          `json:"data"`
	Config map[string]interface{} `json:"config,omitempty"`
}