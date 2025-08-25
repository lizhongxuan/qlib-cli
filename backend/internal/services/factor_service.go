package services

import (
	"fmt"
	"time"

	"qlib-backend/internal/models"
	"qlib-backend/internal/qlib"

	"gorm.io/gorm"
)

type FactorService struct {
	db           *gorm.DB
	factorEngine *qlib.FactorEngine
}

func NewFactorService(db *gorm.DB, factorEngine *qlib.FactorEngine) *FactorService {
	return &FactorService{
		db:           db,
		factorEngine: factorEngine,
	}
}

// CreateFactor 创建新因子
func (s *FactorService) CreateFactor(req FactorCreateRequest, userID uint) (*models.Factor, error) {
	// 验证因子表达式语法
	if err := s.factorEngine.ValidateExpression(req.Expression); err != nil {
		return nil, fmt.Errorf("因子表达式语法错误: %v", err)
	}

	factor := &models.Factor{
		Name:        req.Name,
		Expression:  req.Expression,
		Description: req.Description,
		Category:    req.Category,
		Status:      "active",
		UserID:      userID,
		IsPublic:    req.IsPublic,
	}

	if err := s.db.Create(factor).Error; err != nil {
		return nil, fmt.Errorf("创建因子失败: %v", err)
	}

	return factor, nil
}

// GetFactors 获取因子列表
func (s *FactorService) GetFactors(page, pageSize int, category, status string, userID uint, isPublic *bool) (*PaginatedFactors, error) {
	var factors []models.Factor
	var total int64

	query := s.db.Model(&models.Factor{})

	// 权限过滤：只能看到自己的私有因子和所有公开因子
	if isPublic != nil && *isPublic {
		query = query.Where("is_public = ?", true)
	} else {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	}

	// 添加过滤条件
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取因子总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&factors).Error; err != nil {
		return nil, fmt.Errorf("获取因子列表失败: %v", err)
	}

	return &PaginatedFactors{
		Data:       factors,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// GetFactorByID 根据ID获取因子
func (s *FactorService) GetFactorByID(id uint, userID uint) (*models.Factor, error) {
	var factor models.Factor
	query := s.db.Where("id = ?", id)
	
	// 权限检查：只能访问自己的私有因子或公开因子
	query = query.Where("user_id = ? OR is_public = ?", userID, true)
	
	if err := query.First(&factor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("因子不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取因子失败: %v", err)
	}
	return &factor, nil
}

// UpdateFactor 更新因子
func (s *FactorService) UpdateFactor(id uint, req FactorUpdateRequest, userID uint) (*models.Factor, error) {
	var factor models.Factor
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&factor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("因子不存在或无权限修改")
		}
		return nil, fmt.Errorf("获取因子失败: %v", err)
	}

	// 验证新的表达式语法（如果有更新）
	if req.Expression != "" && req.Expression != factor.Expression {
		if err := s.factorEngine.ValidateExpression(req.Expression); err != nil {
			return nil, fmt.Errorf("因子表达式语法错误: %v", err)
		}
	}

	// 更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Expression != "" {
		updates["expression"] = req.Expression
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if err := s.db.Model(&factor).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新因子失败: %v", err)
	}

	return &factor, nil
}

// DeleteFactor 删除因子
func (s *FactorService) DeleteFactor(id uint, userID uint) error {
	var factor models.Factor
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&factor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("因子不存在或无权限删除")
		}
		return fmt.Errorf("获取因子失败: %v", err)
	}

	if err := s.db.Delete(&factor).Error; err != nil {
		return fmt.Errorf("删除因子失败: %v", err)
	}

	return nil
}

// TestFactor 测试因子性能
func (s *FactorService) TestFactor(req FactorTestRequest, userID uint) (*FactorTestResult, error) {
	// 验证因子表达式
	if err := s.factorEngine.ValidateExpression(req.Expression); err != nil {
		return nil, fmt.Errorf("因子表达式语法错误: %v", err)
	}

	// 调用因子引擎进行测试
	result, err := s.factorEngine.TestFactor(qlib.FactorTestParams{
		Expression:  req.Expression,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Universe:    req.Universe,
		Benchmark:   req.Benchmark,
		Freq:        req.Freq,
	})
	if err != nil {
		return nil, fmt.Errorf("因子测试失败: %v", err)
	}

	testResult := &FactorTestResult{
		IC:         result.IC,
		IR:         result.IR,
		RankIC:     result.RankIC,
		Turnover:   result.Turnover,
		Coverage:   result.Coverage,
		Sharpe:     result.Sharpe,
		Return:     result.Return,
		Volatility: result.Volatility,
		MaxDD:      result.MaxDrawdown,
		TestDate:   time.Now(),
		Details:    result.Details,
	}

	return testResult, nil
}

// BatchTestFactors 批量测试因子
func (s *FactorService) BatchTestFactors(req BatchFactorTestRequest, userID uint) (*BatchFactorTestResult, error) {
	results := make([]FactorTestSummary, 0, len(req.FactorIDs))
	
	for _, factorID := range req.FactorIDs {
		factor, err := s.GetFactorByID(factorID, userID)
		if err != nil {
			// 记录错误但继续处理其他因子
			results = append(results, FactorTestSummary{
				FactorID:    factorID,
				FactorName:  "Unknown",
				Status:      "failed",
				ErrorMsg:    err.Error(),
			})
			continue
		}

		testReq := FactorTestRequest{
			Expression: factor.Expression,
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			Universe:   req.Universe,
			Benchmark:  req.Benchmark,
			Freq:       req.Freq,
		}

		testResult, err := s.TestFactor(testReq, userID)
		if err != nil {
			results = append(results, FactorTestSummary{
				FactorID:    factorID,
				FactorName:  factor.Name,
				Status:      "failed",
				ErrorMsg:    err.Error(),
			})
			continue
		}

		results = append(results, FactorTestSummary{
			FactorID:    factorID,
			FactorName:  factor.Name,
			Status:      "completed",
			IC:          testResult.IC,
			IR:          testResult.IR,
			RankIC:      testResult.RankIC,
			Turnover:    testResult.Turnover,
			Coverage:    testResult.Coverage,
		})

		// 更新因子的性能指标
		s.db.Model(factor).Updates(map[string]interface{}{
			"ic":       testResult.IC,
			"ir":       testResult.IR,
			"rank_ic":  testResult.RankIC,
			"turnover": testResult.Turnover,
			"coverage": testResult.Coverage,
		})
	}

	return &BatchFactorTestResult{
		Results:    results,
		TotalCount: len(req.FactorIDs),
		SuccessCount: func() int {
			count := 0
			for _, r := range results {
				if r.Status == "completed" {
					count++
				}
			}
			return count
		}(),
		FailedCount: func() int {
			count := 0
			for _, r := range results {
				if r.Status == "failed" {
					count++
				}
			}
			return count
		}(),
	}, nil
}

// GetFactorCategories 获取因子分类
func (s *FactorService) GetFactorCategories() ([]FactorCategory, error) {
	// 从数据库获取所有已使用的分类
	var categories []string
	if err := s.db.Model(&models.Factor{}).
		Select("DISTINCT category").
		Where("category != ''").
		Pluck("category", &categories).Error; err != nil {
		return nil, fmt.Errorf("获取因子分类失败: %v", err)
	}

	// 添加预定义分类
	predefinedCategories := []string{
		"price", "volume", "momentum", "reversal", "volatility",
		"quality", "growth", "value", "technical", "fundamental",
	}

	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}
	for _, cat := range predefinedCategories {
		categoryMap[cat] = true
	}

	result := make([]FactorCategory, 0, len(categoryMap))
	for category := range categoryMap {
		// 获取该分类下的因子数量
		var count int64
		s.db.Model(&models.Factor{}).Where("category = ?", category).Count(&count)
		
		result = append(result, FactorCategory{
			Name:        category,
			Count:       int(count),
			Description: getFactorCategoryDescription(category),
		})
	}

	return result, nil
}

// ImportFactors 导入因子库
func (s *FactorService) ImportFactors(req ImportFactorsRequest, userID uint) (*ImportFactorsResult, error) {
	imported := 0
	failed := 0
	errors := make([]string, 0)

	for _, factorData := range req.Factors {
		// 验证因子表达式
		if err := s.factorEngine.ValidateExpression(factorData.Expression); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("因子 %s 表达式语法错误: %v", factorData.Name, err))
			continue
		}

		// 检查是否已存在相同名称的因子
		var existingFactor models.Factor
		if err := s.db.Where("name = ? AND user_id = ?", factorData.Name, userID).First(&existingFactor).Error; err == nil {
			if !req.OverwriteExisting {
				failed++
				errors = append(errors, fmt.Sprintf("因子 %s 已存在", factorData.Name))
				continue
			}
			// 更新现有因子
			updates := map[string]interface{}{
				"expression":  factorData.Expression,
				"description": factorData.Description,
				"category":    factorData.Category,
			}
			if err := s.db.Model(&existingFactor).Updates(updates).Error; err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("更新因子 %s 失败: %v", factorData.Name, err))
				continue
			}
		} else {
			// 创建新因子
			factor := &models.Factor{
				Name:        factorData.Name,
				Expression:  factorData.Expression,
				Description: factorData.Description,
				Category:    factorData.Category,
				Status:      "active",
				UserID:      userID,
				IsPublic:    factorData.IsPublic,
			}

			if err := s.db.Create(factor).Error; err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("创建因子 %s 失败: %v", factorData.Name, err))
				continue
			}
		}

		imported++
	}

	return &ImportFactorsResult{
		ImportedCount: imported,
		FailedCount:   failed,
		Errors:        errors,
	}, nil
}

// GetFactorAnalysis 获取因子分析结果
func (s *FactorService) GetFactorAnalysis(id uint, userID uint) (*FactorAnalysisResult, error) {
	factor, err := s.GetFactorByID(id, userID)
	if err != nil {
		return nil, err
	}

	// 调用因子引擎进行详细分析
	analysis, err := s.factorEngine.AnalyzeFactor(factor.Expression)
	if err != nil {
		return nil, fmt.Errorf("因子分析失败: %v", err)
	}

	return &FactorAnalysisResult{
		FactorID:         factor.ID,
		FactorName:       factor.Name,
		BasicMetrics:     analysis.BasicMetrics,
		TimeSeriesData:   analysis.TimeSeriesData,
		DistributionData: analysis.DistributionData,
		CorrelationData:  analysis.CorrelationData,
		SectorAnalysis:   analysis.SectorAnalysis,
		GeneratedAt:      time.Now(),
	}, nil
}

// getFactorCategoryDescription 获取因子分类描述
func getFactorCategoryDescription(category string) string {
	descriptions := map[string]string{
		"price":       "价格类因子，基于股票价格数据",
		"volume":      "成交量类因子，基于成交量数据",
		"momentum":    "动量类因子，反映价格趋势",
		"reversal":    "反转类因子，基于均值回归理论",
		"volatility":  "波动率类因子，衡量价格波动",
		"quality":     "质量类因子，基于财务质量指标",
		"growth":      "成长类因子，基于成长性指标",
		"value":       "价值类因子，基于估值指标",
		"technical":   "技术类因子，基于技术分析",
		"fundamental": "基本面因子，基于财务数据",
	}
	
	if desc, ok := descriptions[category]; ok {
		return desc
	}
	return "自定义因子分类"
}

// 请求和响应结构体
type FactorCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Expression  string `json:"expression" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsPublic    bool   `json:"is_public"`
}

type FactorUpdateRequest struct {
	Name        string `json:"name"`
	Expression  string `json:"expression"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	IsPublic    *bool  `json:"is_public"`
}

type FactorTestRequest struct {
	Expression string `json:"expression" binding:"required"`
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
	Universe   string `json:"universe"`
	Benchmark  string `json:"benchmark"`
	Freq       string `json:"freq"`
}

type FactorTestResult struct {
	IC         float64                `json:"ic"`
	IR         float64                `json:"ir"`
	RankIC     float64                `json:"rank_ic"`
	Turnover   float64                `json:"turnover"`
	Coverage   float64                `json:"coverage"`
	Sharpe     float64                `json:"sharpe"`
	Return     float64                `json:"return"`
	Volatility float64                `json:"volatility"`
	MaxDD      float64                `json:"max_drawdown"`
	TestDate   time.Time              `json:"test_date"`
	Details    map[string]interface{} `json:"details"`
}

type BatchFactorTestRequest struct {
	FactorIDs []uint `json:"factor_ids" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Universe  string `json:"universe"`
	Benchmark string `json:"benchmark"`
	Freq      string `json:"freq"`
}

type FactorTestSummary struct {
	FactorID   uint    `json:"factor_id"`
	FactorName string  `json:"factor_name"`
	Status     string  `json:"status"`
	IC         float64 `json:"ic"`
	IR         float64 `json:"ir"`
	RankIC     float64 `json:"rank_ic"`
	Turnover   float64 `json:"turnover"`
	Coverage   float64 `json:"coverage"`
	ErrorMsg   string  `json:"error_msg,omitempty"`
}

type BatchFactorTestResult struct {
	Results      []FactorTestSummary `json:"results"`
	TotalCount   int                 `json:"total_count"`
	SuccessCount int                 `json:"success_count"`
	FailedCount  int                 `json:"failed_count"`
}

type FactorCategory struct {
	Name        string `json:"name"`
	Count       int    `json:"count"`
	Description string `json:"description"`
}

type ImportFactorsRequest struct {
	Factors           []FactorData `json:"factors" binding:"required"`
	OverwriteExisting bool         `json:"overwrite_existing"`
}

type FactorData struct {
	Name        string `json:"name" binding:"required"`
	Expression  string `json:"expression" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsPublic    bool   `json:"is_public"`
}

type ImportFactorsResult struct {
	ImportedCount int      `json:"imported_count"`
	FailedCount   int      `json:"failed_count"`
	Errors        []string `json:"errors"`
}

type FactorAnalysisResult struct {
	FactorID         uint                   `json:"factor_id"`
	FactorName       string                 `json:"factor_name"`
	BasicMetrics     map[string]interface{} `json:"basic_metrics"`
	TimeSeriesData   map[string]interface{} `json:"time_series_data"`
	DistributionData map[string]interface{} `json:"distribution_data"`
	CorrelationData  map[string]interface{} `json:"correlation_data"`
	SectorAnalysis   map[string]interface{} `json:"sector_analysis"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

type PaginatedFactors struct {
	Data       []models.Factor `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int64           `json:"total_pages"`
}