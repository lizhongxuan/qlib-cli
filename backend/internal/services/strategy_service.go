package services

import (
	"fmt"
	"time"

	"qlib-backend/internal/models"
	"qlib-backend/internal/qlib"

	"gorm.io/gorm"
)

type StrategyService struct {
	db              *gorm.DB
	backtestEngine  *qlib.BacktestEngine
	taskService     *TaskService
}

func NewStrategyService(db *gorm.DB, backtestEngine *qlib.BacktestEngine, taskService *TaskService) *StrategyService {
	return &StrategyService{
		db:              db,
		backtestEngine:  backtestEngine,
		taskService:     taskService,
	}
}

// StartBacktest 启动策略回测
func (s *StrategyService) StartBacktest(req StrategyBacktestRequest, userID uint) (*StrategyBacktestResponse, error) {
	// 验证回测参数
	if err := s.validateBacktestParams(req); err != nil {
		return nil, fmt.Errorf("回测参数验证失败: %v", err)
	}

	// 如果指定了模型ID，验证模型是否存在且属于用户
	if req.ModelID != 0 {
		var model models.Model
		if err := s.db.Where("id = ? AND user_id = ?", req.ModelID, userID).First(&model).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("指定的模型不存在或无权限访问")
			}
			return nil, fmt.Errorf("验证模型失败: %v", err)
		}
		if model.Status != "completed" && model.Status != "deployed" {
			return nil, fmt.Errorf("模型尚未训练完成")
		}
	}

	// 创建策略记录
	strategy := &models.Strategy{
		Name:           req.Name,
		Type:           req.StrategyType,
		Description:    req.Description,
		Status:         "backtesting",
		Progress:       0,
		ConfigJSON:     req.ConfigJSON,
		ModelID:        req.ModelID,
		BacktestStart:  req.BacktestStart,
		BacktestEnd:    req.BacktestEnd,
		UserID:         userID,
	}

	if err := s.db.Create(strategy).Error; err != nil {
		return nil, fmt.Errorf("创建策略记录失败: %v", err)
	}

	// 创建回测任务
	task := &models.Task{
		Name:        fmt.Sprintf("策略回测: %s", req.Name),
		Type:        "strategy_backtest",
		Status:      "pending",
		Description: fmt.Sprintf("回测%s策略", req.StrategyType),
		ConfigJSON:  req.ConfigJSON,
		UserID:      userID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建回测任务失败: %v", err)
	}

	// 启动异步回测任务
	go s.executeBacktest(strategy.ID, task.ID, req)

	return &StrategyBacktestResponse{
		StrategyID: strategy.ID,
		TaskID:     task.ID,
		Status:     "started",
		Message:    "策略回测已启动",
	}, nil
}

// GetStrategies 获取策略列表
func (s *StrategyService) GetStrategies(page, pageSize int, status, strategyType string, userID uint) (*PaginatedStrategies, error) {
	var strategies []models.Strategy
	var total int64

	query := s.db.Model(&models.Strategy{}).Where("user_id = ?", userID).Preload("Model")

	// 添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if strategyType != "" {
		query = query.Where("type = ?", strategyType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取策略总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&strategies).Error; err != nil {
		return nil, fmt.Errorf("获取策略列表失败: %v", err)
	}

	return &PaginatedStrategies{
		Data:       strategies,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// GetBacktestResults 获取回测结果
func (s *StrategyService) GetBacktestResults(strategyID uint, userID uint) (*BacktestResultsResponse, error) {
	var strategy models.Strategy
	if err := s.db.Where("id = ? AND user_id = ?", strategyID, userID).Preload("Model").First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("策略不存在")
		}
		return nil, fmt.Errorf("获取策略失败: %v", err)
	}

	if strategy.Status != "completed" {
		return nil, fmt.Errorf("策略回测尚未完成")
	}

	// 调用回测引擎获取详细结果
	results, err := s.backtestEngine.GetBacktestResults(qlib.BacktestResultsParams{
		StrategyID: strategyID,
	})
	if err != nil {
		return nil, fmt.Errorf("获取回测结果失败: %v", err)
	}

	return &BacktestResultsResponse{
		StrategyID:       strategy.ID,
		StrategyName:     strategy.Name,
		BasicMetrics:     s.buildBasicMetrics(strategy),
		PerformanceData:  results.PerformanceData,
		RiskMetrics:      results.RiskMetrics,
		PositionData:     results.PositionData,
		TradeDetails:     results.TradeDetails,
		BenchmarkData:    results.BenchmarkData,
		GeneratedAt:      time.Now(),
	}, nil
}

// GetBacktestProgress 获取回测进度
func (s *StrategyService) GetBacktestProgress(strategyID uint, userID uint) (*BacktestProgressResponse, error) {
	var strategy models.Strategy
	if err := s.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("策略不存在")
		}
		return nil, fmt.Errorf("获取策略失败: %v", err)
	}

	// 获取相关任务信息
	var task models.Task
	s.db.Where("type = ? AND config_json LIKE ?", "strategy_backtest", "%\"strategy_id\":"+fmt.Sprint(strategyID)+"%").First(&task)

	return &BacktestProgressResponse{
		StrategyID:  strategy.ID,
		Progress:    strategy.Progress,
		Status:      strategy.Status,
		TaskID:      task.ID,
		StartTime:   task.StartTime,
		ElapsedTime: s.calculateElapsedTime(task.StartTime),
		CurrentStep: s.getCurrentStep(strategy.Progress),
		Logs:        s.getBacktestLogs(strategyID),
	}, nil
}

// StopBacktest 停止回测
func (s *StrategyService) StopBacktest(strategyID uint, userID uint) error {
	var strategy models.Strategy
	if err := s.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("策略不存在")
		}
		return fmt.Errorf("获取策略失败: %v", err)
	}

	if strategy.Status != "backtesting" {
		return fmt.Errorf("策略当前状态不支持停止操作")
	}

	// 更新策略状态
	if err := s.db.Model(&strategy).Updates(map[string]interface{}{
		"status": "cancelled",
	}).Error; err != nil {
		return fmt.Errorf("停止回测失败: %v", err)
	}

	// 通知回测引擎停止回测
	if err := s.backtestEngine.StopBacktest(strategyID); err != nil {
		return fmt.Errorf("停止回测引擎失败: %v", err)
	}

	return nil
}

// GetAttributionAnalysis 获取策略归因分析
func (s *StrategyService) GetAttributionAnalysis(strategyID uint, userID uint) (*AttributionAnalysisResult, error) {
	var strategy models.Strategy
	if err := s.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("策略不存在")
		}
		return nil, fmt.Errorf("获取策略失败: %v", err)
	}

	if strategy.Status != "completed" {
		return nil, fmt.Errorf("策略回测尚未完成")
	}

	// 调用回测引擎进行归因分析
	attribution, err := s.backtestEngine.GetAttributionAnalysis(qlib.AttributionAnalysisParams{
		StrategyID: strategyID,
	})
	if err != nil {
		return nil, fmt.Errorf("归因分析失败: %v", err)
	}

	return &AttributionAnalysisResult{
		StrategyID:        strategyID,
		FactorAttribution: attribution.FactorAttribution,
		SectorAttribution: attribution.SectorAttribution,
		StyleAttribution:  attribution.StyleAttribution,
		SecuritySelection: attribution.SecuritySelection,
		TimingEffect:      attribution.TimingEffect,
		InteractionEffect: attribution.InteractionEffect,
		GeneratedAt:       time.Now(),
	}, nil
}

// CompareStrategies 策略对比
func (s *StrategyService) CompareStrategies(req StrategyComparisonRequest, userID uint) (*StrategyComparisonResult, error) {
	if len(req.StrategyIDs) < 2 {
		return nil, fmt.Errorf("至少需要选择2个策略进行对比")
	}

	var strategies []models.Strategy
	if err := s.db.Where("id IN ? AND user_id = ?", req.StrategyIDs, userID).Preload("Model").Find(&strategies).Error; err != nil {
		return nil, fmt.Errorf("获取策略信息失败: %v", err)
	}

	if len(strategies) != len(req.StrategyIDs) {
		return nil, fmt.Errorf("部分策略不存在或无权限访问")
	}

	// 验证所有策略都已完成回测
	for _, strategy := range strategies {
		if strategy.Status != "completed" {
			return nil, fmt.Errorf("策略 %s 回测尚未完成", strategy.Name)
		}
	}

	// 调用回测引擎进行策略对比
	comparison, err := s.backtestEngine.CompareStrategies(qlib.StrategyComparisonParams{
		StrategyIDs: req.StrategyIDs,
		Metrics:     req.Metrics,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	})
	if err != nil {
		return nil, fmt.Errorf("策略对比失败: %v", err)
	}

	return &StrategyComparisonResult{
		Strategies:       strategies,
		ComparisonMatrix: comparison.ComparisonMatrix,
		RankingResults:   comparison.RankingResults,
		BestStrategy:     comparison.BestStrategy,
		ComparisonDate:   time.Now(),
	}, nil
}

// OptimizeStrategy 参数优化
func (s *StrategyService) OptimizeStrategy(strategyID uint, req StrategyOptimizationRequest, userID uint) (*StrategyOptimizationResponse, error) {
	var strategy models.Strategy
	if err := s.db.Where("id = ? AND user_id = ?", strategyID, userID).First(&strategy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("策略不存在")
		}
		return nil, fmt.Errorf("获取策略失败: %v", err)
	}

	// 创建优化任务
	task := &models.Task{
		Name:        fmt.Sprintf("策略优化: %s", strategy.Name),
		Type:        "strategy_optimization",
		Status:      "pending",
		Description: "参数优化任务",
		ConfigJSON:  req.ConfigJSON,
		UserID:      userID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建优化任务失败: %v", err)
	}

	// 启动异步优化任务
	go s.executeOptimization(strategyID, task.ID, req)

	return &StrategyOptimizationResponse{
		StrategyID: strategyID,
		TaskID:     task.ID,
		Status:     "started",
		Message:    "参数优化已启动",
	}, nil
}

// ExportBacktestReport 导出回测报告
func (s *StrategyService) ExportBacktestReport(req models.BacktestReportExportRequestExtended, userID uint) (*BacktestReportExportResponse, error) {
	// 验证所有回测结果都属于用户
	var strategies []models.Strategy
	if err := s.db.Where("id IN ? AND user_id = ?", req.ResultIDs, userID).Find(&strategies).Error; err != nil {
		return nil, fmt.Errorf("获取策略信息失败: %v", err)
	}

	if len(strategies) != len(req.ResultIDs) {
		return nil, fmt.Errorf("部分策略不存在或无权限访问")
	}

	// 调用回测引擎生成报告
	report, err := s.backtestEngine.ExportReport(qlib.ReportExportParams{
		StrategyIDs: req.ResultIDs, // 使用ResultIDs作为StrategyIDs
		Format:      req.Format,
		Sections:    req.Sections,
	})
	if err != nil {
		return nil, fmt.Errorf("生成报告失败: %v", err)
	}

	return &BacktestReportExportResponse{
		ReportID:    report.ReportID,
		DownloadURL: report.DownloadURL,
		Format:      req.Format,
		ExportedAt:  time.Now(),
	}, nil
}

// 内部方法

// validateBacktestParams 验证回测参数
func (s *StrategyService) validateBacktestParams(req StrategyBacktestRequest) error {
	if req.Name == "" {
		return fmt.Errorf("策略名称不能为空")
	}
	if req.StrategyType == "" {
		return fmt.Errorf("策略类型不能为空")
	}
	if req.BacktestStart == "" || req.BacktestEnd == "" {
		return fmt.Errorf("回测时间范围不能为空")
	}

	// 验证支持的策略类型
	supportedTypes := map[string]bool{
		"TopkDropoutStrategy": true,
		"WeightStrategyBase":  true,
		"BuyAndHoldStrategy":  true,
		"FixedWeightStrategy": true,
	}

	if !supportedTypes[req.StrategyType] {
		return fmt.Errorf("不支持的策略类型: %s", req.StrategyType)
	}

	return nil
}

// executeBacktest 执行回测任务
func (s *StrategyService) executeBacktest(strategyID, taskID uint, req StrategyBacktestRequest) {
	// 更新任务状态
	s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": time.Now(),
	})

	// 更新策略状态
	s.db.Model(&models.Strategy{}).Where("id = ?", strategyID).Updates(map[string]interface{}{
		"status": "backtesting",
	})

	// 调用回测引擎
	backtestParams := qlib.BacktestParams{
		StrategyID:    strategyID,
		StrategyType:  req.StrategyType,
		ModelID:       req.ModelID,
		ConfigJSON:    req.ConfigJSON,
		BacktestStart: req.BacktestStart,
		BacktestEnd:   req.BacktestEnd,
		Universe:      req.Universe,
		Benchmark:     req.Benchmark,
	}

	// 设置进度回调
	progressCallback := func(progress int, metrics map[string]float64) {
		updates := map[string]interface{}{
			"progress": progress,
		}

		// 更新性能指标
		if totalReturn, ok := metrics["total_return"]; ok {
			updates["total_return"] = totalReturn
		}
		if annualReturn, ok := metrics["annual_return"]; ok {
			updates["annual_return"] = annualReturn
		}
		if sharpeRatio, ok := metrics["sharpe_ratio"]; ok {
			updates["sharpe_ratio"] = sharpeRatio
		}
		if maxDrawdown, ok := metrics["max_drawdown"]; ok {
			updates["max_drawdown"] = maxDrawdown
		}
		if volatility, ok := metrics["volatility"]; ok {
			updates["volatility"] = volatility
		}

		s.db.Model(&models.Strategy{}).Where("id = ?", strategyID).Updates(updates)
	}

	// 执行回测
	result, err := s.backtestEngine.RunBacktest(backtestParams, progressCallback)

	// 更新最终状态
	if err != nil {
		// 回测失败
		s.db.Model(&models.Strategy{}).Where("id = ?", strategyID).Updates(map[string]interface{}{
			"status": "failed",
		})
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": err.Error(),
			"end_time":  time.Now(),
		})
	} else {
		// 回测成功
		s.db.Model(&models.Strategy{}).Where("id = ?", strategyID).Updates(map[string]interface{}{
			"status":          "completed",
			"progress":        100,
			"total_return":    result.TotalReturn,
			"annual_return":   result.AnnualReturn,
			"excess_return":   result.ExcessReturn,
			"sharpe_ratio":    result.SharpeRatio,
			"max_drawdown":    result.MaxDrawdown,
			"volatility":      result.Volatility,
			"win_rate":        result.WinRate,
		})
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":   "completed",
			"end_time": time.Now(),
		})
	}
}

// executeOptimization 执行参数优化
func (s *StrategyService) executeOptimization(strategyID, taskID uint, req StrategyOptimizationRequest) {
	// 更新任务状态
	s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": time.Now(),
	})

	// 调用回测引擎进行参数优化
	optimizationParams := qlib.OptimizationParams{
		StrategyID:      strategyID,
		ParameterRanges: req.ParameterRanges,
		OptimizationMethod: req.OptimizationMethod,
		TargetMetric:    req.TargetMetric,
		MaxIterations:   req.MaxIterations,
	}

	result, err := s.backtestEngine.OptimizeParameters(optimizationParams)

	// 更新任务状态
	if err != nil {
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": err.Error(),
			"end_time":  time.Now(),
		})
	} else {
		resultJSON, _ := result.ToJSON()
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":      "completed",
			"result_json": resultJSON,
			"end_time":    time.Now(),
		})
	}
}

// buildBasicMetrics 构建基础指标
func (s *StrategyService) buildBasicMetrics(strategy models.Strategy) map[string]interface{} {
	return map[string]interface{}{
		"total_return":    strategy.TotalReturn,
		"annual_return":   strategy.AnnualReturn,
		"excess_return":   strategy.ExcessReturn,
		"sharpe_ratio":    strategy.SharpeRatio,
		"max_drawdown":    strategy.MaxDrawdown,
		"volatility":      strategy.Volatility,
		"win_rate":        strategy.WinRate,
		"backtest_start":  strategy.BacktestStart,
		"backtest_end":    strategy.BacktestEnd,
	}
}

// getBacktestLogs 获取回测日志
func (s *StrategyService) getBacktestLogs(strategyID uint) []string {
	// 这里应该从日志文件或数据库中读取实际的回测日志
	return []string{
		fmt.Sprintf("[%s] 开始策略回测 ID: %d", time.Now().Format("2006-01-02 15:04:05"), strategyID),
		fmt.Sprintf("[%s] 数据加载完成", time.Now().Add(-5*time.Minute).Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] 策略初始化完成", time.Now().Add(-4*time.Minute).Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] 回测进行中...", time.Now().Add(-3*time.Minute).Format("2006-01-02 15:04:05")),
	}
}

// getCurrentStep 获取当前步骤
func (s *StrategyService) getCurrentStep(progress int) string {
	switch {
	case progress < 10:
		return "初始化"
	case progress < 30:
		return "数据加载"
	case progress < 50:
		return "策略执行"
	case progress < 80:
		return "结果计算"
	case progress < 100:
		return "报告生成"
	default:
		return "已完成"
	}
}

// calculateElapsedTime 计算运行时间
func (s *StrategyService) calculateElapsedTime(startTime *time.Time) int64 {
	if startTime == nil {
		return 0
	}
	return int64(time.Since(*startTime).Seconds())
}

// 请求和响应结构体
type StrategyBacktestRequest struct {
	Name          string `json:"name" binding:"required"`
	StrategyType  string `json:"strategy_type" binding:"required"`
	Description   string `json:"description"`
	ModelID       uint   `json:"model_id"`
	ConfigJSON    string `json:"config_json" binding:"required"`
	BacktestStart string `json:"backtest_start" binding:"required"`
	BacktestEnd   string `json:"backtest_end" binding:"required"`
	Universe      string `json:"universe"`
	Benchmark     string `json:"benchmark"`
}

type StrategyBacktestResponse struct {
	StrategyID uint   `json:"strategy_id"`
	TaskID     uint   `json:"task_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

type BacktestResultsResponse struct {
	StrategyID      uint                   `json:"strategy_id"`
	StrategyName    string                 `json:"strategy_name"`
	BasicMetrics    map[string]interface{} `json:"basic_metrics"`
	PerformanceData map[string]interface{} `json:"performance_data"`
	RiskMetrics     map[string]interface{} `json:"risk_metrics"`
	PositionData    map[string]interface{} `json:"position_data"`
	TradeDetails    map[string]interface{} `json:"trade_details"`
	BenchmarkData   map[string]interface{} `json:"benchmark_data"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

type BacktestProgressResponse struct {
	StrategyID  uint       `json:"strategy_id"`
	Progress    int        `json:"progress"`
	Status      string     `json:"status"`
	TaskID      uint       `json:"task_id"`
	StartTime   *time.Time `json:"start_time"`
	ElapsedTime int64      `json:"elapsed_time"`
	CurrentStep string     `json:"current_step"`
	Logs        []string   `json:"logs"`
}

type AttributionAnalysisResult struct {
	StrategyID        uint                   `json:"strategy_id"`
	FactorAttribution map[string]interface{} `json:"factor_attribution"`
	SectorAttribution map[string]interface{} `json:"sector_attribution"`
	StyleAttribution  map[string]interface{} `json:"style_attribution"`
	SecuritySelection map[string]interface{} `json:"security_selection"`
	TimingEffect      map[string]interface{} `json:"timing_effect"`
	InteractionEffect map[string]interface{} `json:"interaction_effect"`
	GeneratedAt       time.Time              `json:"generated_at"`
}

type StrategyComparisonRequest struct {
	StrategyIDs []uint   `json:"strategy_ids" binding:"required"`
	Metrics     []string `json:"metrics"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
}

type StrategyComparisonResult struct {
	Strategies       []models.Strategy      `json:"strategies"`
	ComparisonMatrix map[string]interface{} `json:"comparison_matrix"`
	RankingResults   map[string]interface{} `json:"ranking_results"`
	BestStrategy     map[string]interface{} `json:"best_strategy"`
	ComparisonDate   time.Time              `json:"comparison_date"`
}

type StrategyOptimizationRequest struct {
	ParameterRanges    map[string]interface{} `json:"parameter_ranges" binding:"required"`
	OptimizationMethod string                 `json:"optimization_method"`
	TargetMetric       string                 `json:"target_metric"`
	MaxIterations      int                    `json:"max_iterations"`
	ConfigJSON         string                 `json:"config_json"`
}

type StrategyOptimizationResponse struct {
	StrategyID uint   `json:"strategy_id"`
	TaskID     uint   `json:"task_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}


type BacktestReportExportResponse struct {
	ReportID    string    `json:"report_id"`
	DownloadURL string    `json:"download_url"`
	Format      string    `json:"format"`
	ExportedAt  time.Time `json:"exported_at"`
}

type PaginatedStrategies struct {
	Data       []models.Strategy `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int64             `json:"total_pages"`
}