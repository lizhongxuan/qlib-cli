package services

import (
	"fmt"
	"time"

	"qlib-backend/internal/qlib"

	"gorm.io/gorm"
)

type FactorResearchService struct {
	db             *gorm.DB
	factorEngine   *qlib.FactorEngine
	syntaxValidator *qlib.SyntaxValidator
	aiChatService  *AiChatService
}

func NewFactorResearchService(db *gorm.DB, factorEngine *qlib.FactorEngine, syntaxValidator *qlib.SyntaxValidator, aiChatService *AiChatService) *FactorResearchService {
	return &FactorResearchService{
		db:             db,
		factorEngine:   factorEngine,
		syntaxValidator: syntaxValidator,
		aiChatService:  aiChatService,
	}
}

// GetQlibCategories 获取Qlib内置因子分类
func (s *FactorResearchService) GetQlibCategories() ([]QlibFactorCategory, error) {
	categories, err := s.factorEngine.GetBuiltinFactorCategories()
	if err != nil {
		return nil, fmt.Errorf("获取Qlib因子分类失败: %v", err)
	}

	result := make([]QlibFactorCategory, len(categories))
	for i, cat := range categories {
		factors, err := s.factorEngine.GetBuiltinFactorsByCategory(cat.Name)
		if err != nil {
			return nil, fmt.Errorf("获取分类 %s 下的因子失败: %v", cat.Name, err)
		}

		result[i] = QlibFactorCategory{
			Name:        cat.Name,
			Description: cat.Description,
			Count:       len(factors),
			Factors:     factors,
		}
	}

	return result, nil
}

// ValidateFactorSyntax 验证因子表达式语法
func (s *FactorResearchService) ValidateFactorSyntax(req ValidateFactorSyntaxRequest) (*ValidateFactorSyntaxResult, error) {
	result, err := s.syntaxValidator.Validate(req.Expression)
	if err != nil {
		return nil, fmt.Errorf("语法验证失败: %v", err)
	}

	return &ValidateFactorSyntaxResult{
		IsValid:     result.IsValid,
		ErrorMsg:    result.ErrorMsg,
		Suggestions: result.Suggestions,
		ParsedAST:   result.ParsedAST,
		UsedFields:  result.UsedFields,
		UsedFunctions: result.UsedFunctions,
	}, nil
}

// GetQlibFunctions 获取Qlib函数列表
func (s *FactorResearchService) GetQlibFunctions(category string) ([]QlibFunction, error) {
	functions, err := s.factorEngine.GetQlibFunctions()
	if err != nil {
		return nil, fmt.Errorf("获取Qlib函数列表失败: %v", err)
	}

	// 根据分类过滤
	if category != "" {
		filteredFunctions := make([]QlibFunction, 0)
		for _, fn := range functions {
			if fn.Category == category {
				filteredFunctions = append(filteredFunctions, fn)
			}
		}
		return filteredFunctions, nil
	}

	return functions, nil
}

// GetSyntaxReference 获取语法参考
func (s *FactorResearchService) GetSyntaxReference() (*SyntaxReference, error) {
	return &SyntaxReference{
		Operators: []OperatorInfo{
			{Name: "+", Description: "加法运算", Example: "$close + $open", Category: "算术运算符"},
			{Name: "-", Description: "减法运算", Example: "$close - $open", Category: "算术运算符"},
			{Name: "*", Description: "乘法运算", Example: "$close * $volume", Category: "算术运算符"},
			{Name: "/", Description: "除法运算", Example: "$close / $open", Category: "算术运算符"},
			{Name: ">", Description: "大于比较", Example: "$close > $open", Category: "比较运算符"},
			{Name: "<", Description: "小于比较", Example: "$close < $open", Category: "比较运算符"},
			{Name: ">=", Description: "大于等于", Example: "$close >= $open", Category: "比较运算符"},
			{Name: "<=", Description: "小于等于", Example: "$close <= $open", Category: "比较运算符"},
			{Name: "==", Description: "等于", Example: "$close == $open", Category: "比较运算符"},
			{Name: "!=", Description: "不等于", Example: "$close != $open", Category: "比较运算符"},
		},
		Fields: []FieldInfo{
			{Name: "$open", Description: "开盘价", DataType: "float", Example: "$open"},
			{Name: "$high", Description: "最高价", DataType: "float", Example: "$high"},
			{Name: "$low", Description: "最低价", DataType: "float", Example: "$low"},
			{Name: "$close", Description: "收盘价", DataType: "float", Example: "$close"},
			{Name: "$volume", Description: "成交量", DataType: "int", Example: "$volume"},
			{Name: "$factor", Description: "复权因子", DataType: "float", Example: "$factor"},
			{Name: "$vwap", Description: "成交量加权平均价", DataType: "float", Example: "$vwap"},
		},
		Functions: []FunctionInfo{
			{
				Name:        "Mean",
				Description: "计算移动平均值",
				Signature:   "Mean(data, window)",
				Parameters: []ParameterInfo{
					{Name: "data", Type: "Series", Description: "输入数据序列"},
					{Name: "window", Type: "int", Description: "窗口大小"},
				},
				ReturnType: "Series",
				Examples:   []string{"Mean($close, 20)", "Mean($volume, 10)"},
				Category:   "统计函数",
			},
			{
				Name:        "Std",
				Description: "计算标准差",
				Signature:   "Std(data, window)",
				Parameters: []ParameterInfo{
					{Name: "data", Type: "Series", Description: "输入数据序列"},
					{Name: "window", Type: "int", Description: "窗口大小"},
				},
				ReturnType: "Series",
				Examples:   []string{"Std($close, 20)", "Std($return, 60)"},
				Category:   "统计函数",
			},
			{
				Name:        "Corr",
				Description: "计算相关系数",
				Signature:   "Corr(data1, data2, window)",
				Parameters: []ParameterInfo{
					{Name: "data1", Type: "Series", Description: "第一个数据序列"},
					{Name: "data2", Type: "Series", Description: "第二个数据序列"},
					{Name: "window", Type: "int", Description: "窗口大小"},
				},
				ReturnType: "Series",
				Examples:   []string{"Corr($close, $volume, 20)"},
				Category:   "统计函数",
			},
			{
				Name:        "Rank",
				Description: "计算排名",
				Signature:   "Rank(data)",
				Parameters: []ParameterInfo{
					{Name: "data", Type: "Series", Description: "输入数据序列"},
				},
				ReturnType: "Series",
				Examples:   []string{"Rank($close)", "Rank($volume)"},
				Category:   "排序函数",
			},
		},
		Examples: []ExampleInfo{
			{
				Name:        "价格动量因子",
				Expression:  "Ref($close, 20) / $close - 1",
				Description: "20日价格变化率",
				Category:    "动量因子",
			},
			{
				Name:        "成交量比率",
				Expression:  "$volume / Mean($volume, 20)",
				Description: "当日成交量相对20日均量的比率",
				Category:    "成交量因子",
			},
			{
				Name:        "价格相对强度",
				Expression:  "($close - Mean($close, 20)) / Std($close, 20)",
				Description: "价格相对于均值的标准化偏离度",
				Category:    "技术因子",
			},
		},
	}, nil
}

// SaveFactorWorkspace 保存工作区因子
func (s *FactorResearchService) SaveFactorWorkspace(req SaveFactorWorkspaceRequest, userID uint) (*SaveFactorWorkspaceResult, error) {
	// 验证因子表达式
	if err := s.factorEngine.ValidateExpression(req.Expression); err != nil {
		return nil, fmt.Errorf("因子表达式无效: %v", err)
	}

	// 创建工作区因子记录
	workspace := &FactorWorkspace{
		UserID:      userID,
		Name:        req.Name,
		Expression:  req.Expression,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		IsPublic:    req.IsPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(workspace).Error; err != nil {
		return nil, fmt.Errorf("保存工作区因子失败: %v", err)
	}

	return &SaveFactorWorkspaceResult{
		WorkspaceID: workspace.ID,
		Message:     "工作区因子保存成功",
	}, nil
}

// GetFactorWorkspaces 获取用户的工作区因子
func (s *FactorResearchService) GetFactorWorkspaces(userID uint, page, pageSize int) (*PaginatedFactorWorkspaces, error) {
	var workspaces []FactorWorkspace
	var total int64

	query := s.db.Model(&FactorWorkspace{}).Where("user_id = ? OR is_public = ?", userID, true)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取工作区因子总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&workspaces).Error; err != nil {
		return nil, fmt.Errorf("获取工作区因子列表失败: %v", err)
	}

	return &PaginatedFactorWorkspaces{
		Data:       workspaces,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// TestFactorInWorkspace 在工作区测试因子
func (s *FactorResearchService) TestFactorInWorkspace(req WorkspaceFactorTestRequest, userID uint) (*WorkspaceFactorTestResult, error) {
	// 验证表达式
	if err := s.factorEngine.ValidateExpression(req.Expression); err != nil {
		return nil, fmt.Errorf("因子表达式无效: %v", err)
	}

	// 执行因子测试
	testParams := qlib.FactorTestParams{
		Expression: req.Expression,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Universe:   req.Universe,
		Benchmark:  req.Benchmark,
		Freq:       req.Freq,
	}

	result, err := s.factorEngine.TestFactor(testParams)
	if err != nil {
		return nil, fmt.Errorf("因子测试失败: %v", err)
	}

	// 保存测试历史记录
	testHistory := &FactorTestHistory{
		UserID:     userID,
		Expression: req.Expression,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Universe:   req.Universe,
		IC:         result.IC,
		IR:         result.IR,
		RankIC:     result.RankIC,
		Turnover:   result.Turnover,
		Coverage:   result.Coverage,
		TestDate:   time.Now(),
	}

	if err := s.db.Create(testHistory).Error; err != nil {
		// 记录日志但不阻止返回结果
		fmt.Printf("Warning: 保存测试历史失败: %v\n", err)
	}

	return &WorkspaceFactorTestResult{
		IC:         result.IC,
		IR:         result.IR,
		RankIC:     result.RankIC,
		Turnover:   result.Turnover,
		Coverage:   result.Coverage,
		Sharpe:     result.Sharpe,
		Return:     result.Return,
		Volatility: result.Volatility,
		MaxDD:      result.MaxDrawdown,
		Details:    result.Details,
		TestDate:   time.Now(),
	}, nil
}

// GetFactorTestHistory 获取因子测试历史
func (s *FactorResearchService) GetFactorTestHistory(userID uint, limit int) ([]FactorTestHistory, error) {
	var history []FactorTestHistory
	
	query := s.db.Where("user_id = ?", userID).Order("test_date DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&history).Error; err != nil {
		return nil, fmt.Errorf("获取测试历史失败: %v", err)
	}

	return history, nil
}

// 数据结构定义
type QlibFactorCategory struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Count       int                      `json:"count"`
	Factors     []qlib.BuiltinFactor    `json:"factors"`
}

type QlibFunction = qlib.QlibFunction

type ValidateFactorSyntaxRequest struct {
	Expression string `json:"expression" binding:"required"`
}

type ValidateFactorSyntaxResult struct {
	IsValid       bool     `json:"is_valid"`
	ErrorMsg      string   `json:"error_msg,omitempty"`
	Suggestions   []string `json:"suggestions,omitempty"`
	ParsedAST     string   `json:"parsed_ast,omitempty"`
	UsedFields    []string `json:"used_fields,omitempty"`
	UsedFunctions []string `json:"used_functions,omitempty"`
}

type SyntaxReference struct {
	Operators []OperatorInfo `json:"operators"`
	Fields    []FieldInfo    `json:"fields"`
	Functions []FunctionInfo `json:"functions"`
	Examples  []ExampleInfo  `json:"examples"`
}

type OperatorInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Category    string `json:"category"`
}

type FieldInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DataType    string `json:"data_type"`
	Example     string `json:"example"`
}

type FunctionInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Signature   string          `json:"signature"`
	Parameters  []ParameterInfo `json:"parameters"`
	ReturnType  string          `json:"return_type"`
	Examples    []string        `json:"examples"`
	Category    string          `json:"category"`
}

type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ExampleInfo struct {
	Name        string `json:"name"`
	Expression  string `json:"expression"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type SaveFactorWorkspaceRequest struct {
	Name        string   `json:"name" binding:"required"`
	Expression  string   `json:"expression" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	IsPublic    bool     `json:"is_public"`
}

type SaveFactorWorkspaceResult struct {
	WorkspaceID uint   `json:"workspace_id"`
	Message     string `json:"message"`
}

type WorkspaceFactorTestRequest struct {
	Expression string `json:"expression" binding:"required"`
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
	Universe   string `json:"universe"`
	Benchmark  string `json:"benchmark"`
	Freq       string `json:"freq"`
}

type WorkspaceFactorTestResult struct {
	IC         float64                `json:"ic"`
	IR         float64                `json:"ir"`
	RankIC     float64                `json:"rank_ic"`
	Turnover   float64                `json:"turnover"`
	Coverage   float64                `json:"coverage"`
	Sharpe     float64                `json:"sharpe"`
	Return     float64                `json:"return"`
	Volatility float64                `json:"volatility"`
	MaxDD      float64                `json:"max_drawdown"`
	Details    map[string]interface{} `json:"details"`
	TestDate   time.Time              `json:"test_date"`
}

type PaginatedFactorWorkspaces struct {
	Data       []FactorWorkspace `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int64             `json:"total_pages"`
}

// 数据库模型
type FactorWorkspace struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Expression  string    `json:"expression" gorm:"type:text;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Category    string    `json:"category" gorm:"size:50"`
	Tags        []string  `json:"tags" gorm:"type:json"`
	IsPublic    bool      `json:"is_public" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FactorTestHistory struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null"`
	Expression string    `json:"expression" gorm:"type:text;not null"`
	StartDate  string    `json:"start_date" gorm:"size:10"`
	EndDate    string    `json:"end_date" gorm:"size:10"`
	Universe   string    `json:"universe" gorm:"size:50"`
	IC         float64   `json:"ic"`
	IR         float64   `json:"ir"`
	RankIC     float64   `json:"rank_ic"`
	Turnover   float64   `json:"turnover"`
	Coverage   float64   `json:"coverage"`
	TestDate   time.Time `json:"test_date"`
}