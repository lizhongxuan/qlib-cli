package services

import (
	"fmt"
	"time"

	"qlib-backend/internal/utils"

	"gorm.io/gorm"
)

// WorkflowConfigService 工作流配置服务
type WorkflowConfigService struct {
	db            *gorm.DB
	yamlGenerator *utils.YAMLGenerator
}

// WorkflowConfigRequest 工作流配置请求
type WorkflowConfigRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Steps       []ConfigStep           `json:"steps" binding:"required"`
	Config      map[string]interface{} `json:"config"`
}

// ConfigStep 配置步骤
type ConfigStep struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
	Required     bool                   `json:"required"`
	Enabled      bool                   `json:"enabled"`
}

// WorkflowValidationResult 工作流验证结果
type WorkflowValidationResult struct {
	IsValid  bool                      `json:"is_valid"`
	Errors   []ValidationError         `json:"errors"`
	Warnings []ValidationWarning       `json:"warnings"`
	Summary  WorkflowValidationSummary `json:"summary"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Step    string `json:"step,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	Field   string `json:"field"`
	Step    string `json:"step,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WorkflowValidationSummary 工作流验证汇总
type WorkflowValidationSummary struct {
	TotalSteps    int              `json:"total_steps"`
	EnabledSteps  int              `json:"enabled_steps"`
	RequiredSteps int              `json:"required_steps"`
	EstimatedTime time.Duration    `json:"estimated_time"`
	ResourceUsage ResourceEstimate `json:"resource_usage"`
	Dependencies  []string         `json:"dependencies"`
	OutputFormats []string         `json:"output_formats"`
}

// ResourceEstimate 资源估算
type ResourceEstimate struct {
	Memory    string `json:"memory"`
	Storage   string `json:"storage"`
	CPU       string `json:"cpu"`
	GPUNeeded bool   `json:"gpu_needed"`
}

// PresetTemplate 预设模板
type PresetTemplate struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Icon          string                 `json:"icon"`
	Tags          []string               `json:"tags"`
	Difficulty    string                 `json:"difficulty"` // beginner, intermediate, advanced
	EstimatedTime string                 `json:"estimated_time"`
	Config        map[string]interface{} `json:"config"`
	Steps         []ConfigStep           `json:"steps"`
}

// NewWorkflowConfigService 创建新的工作流配置服务
func NewWorkflowConfigService(db *gorm.DB) *WorkflowConfigService {
	yamlGenerator := utils.NewYAMLGenerator()

	return &WorkflowConfigService{
		db:            db,
		yamlGenerator: yamlGenerator,
	}
}

// GetPresetTemplates 获取预设工作流模板
func (wcs *WorkflowConfigService) GetPresetTemplates(category string) ([]PresetTemplate, error) {
	templates := wcs.getBuiltinTemplates()

	if category != "" {
		var filtered []PresetTemplate
		for _, template := range templates {
			if template.Category == category {
				filtered = append(filtered, template)
			}
		}
		return filtered, nil
	}

	return templates, nil
}

// ValidateWorkflowConfig 验证工作流配置
func (wcs *WorkflowConfigService) ValidateWorkflowConfig(req WorkflowConfigRequest) (*WorkflowValidationResult, error) {
	result := &WorkflowValidationResult{
		IsValid:  true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationWarning, 0),
		Summary: WorkflowValidationSummary{
			Dependencies:  make([]string, 0),
			OutputFormats: make([]string, 0),
		},
	}

	// 基础验证
	wcs.validateBasicConfig(req, result)

	// 步骤验证
	wcs.validateSteps(req.Steps, result)

	// 依赖关系验证
	wcs.validateDependencies(req.Steps, result)

	// 生成汇总信息
	wcs.generateValidationSummary(req, result)

	return result, nil
}

// validateBasicConfig 验证基础配置
func (wcs *WorkflowConfigService) validateBasicConfig(req WorkflowConfigRequest, result *WorkflowValidationResult) {
	if req.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Code:    "REQUIRED",
			Message: "工作流名称不能为空",
		})
		result.IsValid = false
	}

	if len(req.Steps) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "steps",
			Code:    "REQUIRED",
			Message: "至少需要一个步骤",
		})
		result.IsValid = false
	}
}

// validateSteps 验证步骤配置
func (wcs *WorkflowConfigService) validateSteps(steps []ConfigStep, result *WorkflowValidationResult) {
	stepNames := make(map[string]bool)
	supportedTypes := map[string]bool{
		"data_preparation":  true,
		"factor_generation": true,
		"model_training":    true,
		"strategy_backtest": true,
		"result_analysis":   true,
		"report_generation": true,
	}

	for i, step := range steps {
		stepPrefix := fmt.Sprintf("steps[%d]", i)

		// 验证步骤名称
		if step.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   stepPrefix + ".name",
				Step:    step.Name,
				Code:    "REQUIRED",
				Message: "步骤名称不能为空",
			})
			result.IsValid = false
		} else {
			if stepNames[step.Name] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   stepPrefix + ".name",
					Step:    step.Name,
					Code:    "DUPLICATE",
					Message: fmt.Sprintf("步骤名称 '%s' 重复", step.Name),
				})
				result.IsValid = false
			}
			stepNames[step.Name] = true
		}

		// 验证步骤类型
		if !supportedTypes[step.Type] {
			result.Errors = append(result.Errors, ValidationError{
				Field:   stepPrefix + ".type",
				Step:    step.Name,
				Code:    "UNSUPPORTED_TYPE",
				Message: fmt.Sprintf("不支持的步骤类型: %s", step.Type),
			})
			result.IsValid = false
		}
	}
}

// validateDependencies 验证依赖关系
func (wcs *WorkflowConfigService) validateDependencies(steps []ConfigStep, result *WorkflowValidationResult) {
	stepNames := make(map[string]bool)
	for _, step := range steps {
		if step.Enabled {
			stepNames[step.Name] = true
		}
	}

	for i, step := range steps {
		if !step.Enabled {
			continue
		}

		stepPrefix := fmt.Sprintf("steps[%d]", i)

		for _, dep := range step.Dependencies {
			if !stepNames[dep] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   stepPrefix + ".dependencies",
					Step:    step.Name,
					Code:    "MISSING_DEPENDENCY",
					Message: fmt.Sprintf("依赖步骤 '%s' 不存在或未启用", dep),
				})
				result.IsValid = false
			}
		}
	}
}

// GenerateYAMLConfig 生成YAML配置文件
func (wcs *WorkflowConfigService) GenerateYAMLConfig(req WorkflowConfigRequest) (string, error) {
	// 验证配置
	validation, err := wcs.ValidateWorkflowConfig(req)
	if err != nil {
		return "", fmt.Errorf("配置验证失败: %v", err)
	}

	if !validation.IsValid {
		return "", fmt.Errorf("配置包含错误，无法生成YAML")
	}

	// 生成YAML
	config := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"category":    req.Category,
		"version":     "1.0",
		"created_at":  time.Now().Format(time.RFC3339),
		"config":      req.Config,
		"steps":       req.Steps,
		"metadata": map[string]interface{}{
			"total_steps":    len(req.Steps),
			"enabled_steps":  wcs.countEnabledSteps(req.Steps),
			"required_steps": wcs.countRequiredSteps(req.Steps),
		},
	}

	return wcs.yamlGenerator.Generate(config)
}

// 辅助方法

// countEnabledSteps 统计启用的步骤数
func (wcs *WorkflowConfigService) countEnabledSteps(steps []ConfigStep) int {
	count := 0
	for _, step := range steps {
		if step.Enabled {
			count++
		}
	}
	return count
}

// countRequiredSteps 统计必需的步骤数
func (wcs *WorkflowConfigService) countRequiredSteps(steps []ConfigStep) int {
	count := 0
	for _, step := range steps {
		if step.Required {
			count++
		}
	}
	return count
}

// generateValidationSummary 生成验证汇总
func (wcs *WorkflowConfigService) generateValidationSummary(req WorkflowConfigRequest, result *WorkflowValidationResult) {
	result.Summary.TotalSteps = len(req.Steps)
	result.Summary.EnabledSteps = wcs.countEnabledSteps(req.Steps)
	result.Summary.RequiredSteps = wcs.countRequiredSteps(req.Steps)

	// 估算执行时间
	estimatedMinutes := 0
	for _, step := range req.Steps {
		if !step.Enabled {
			continue
		}
		switch step.Type {
		case "data_preparation":
			estimatedMinutes += 5
		case "factor_generation":
			estimatedMinutes += 10
		case "model_training":
			estimatedMinutes += 30
		case "strategy_backtest":
			estimatedMinutes += 15
		case "result_analysis":
			estimatedMinutes += 5
		case "report_generation":
			estimatedMinutes += 2
		}
	}
	result.Summary.EstimatedTime = time.Duration(estimatedMinutes) * time.Minute

	// 资源估算
	result.Summary.ResourceUsage = ResourceEstimate{
		Memory:    "2-8GB",
		Storage:   "1-5GB",
		CPU:       "2-4 cores",
		GPUNeeded: false,
	}

	// 输出格式
	result.Summary.OutputFormats = []string{"json", "csv", "html", "pdf"}
}

// getBuiltinTemplates 获取内置模板
func (wcs *WorkflowConfigService) getBuiltinTemplates() []PresetTemplate {
	return []PresetTemplate{
		{
			ID:            "basic_quant_strategy",
			Name:          "Basic Quantitative Strategy",
			Description:   "Complete workflow for beginners including data preparation, factor generation, model training and backtesting",
			Category:      "strategy",
			Icon:          "chart",
			Tags:          []string{"beginner", "complete", "stocks"},
			Difficulty:    "beginner",
			EstimatedTime: "30-60min",
			Config: map[string]interface{}{
				"market":    "CSI300",
				"benchmark": "CSI300",
				"frequency": "daily",
			},
			Steps: []ConfigStep{
				{
					Name:        "Data Preparation",
					Type:        "data_preparation",
					Description: "Fetch and clean stock data",
					Required:    true,
					Enabled:     true,
					Config: map[string]interface{}{
						"instruments": []string{"000001.SZ", "000002.SZ", "600000.SH"},
						"start_time":  "2020-01-01",
						"end_time":    "2023-12-31",
						"fields":      []string{"$close", "$volume", "$high", "$low", "$open"},
					},
				},
				{
					Name:         "Factor Generation",
					Type:         "factor_generation",
					Description:  "Generate technical analysis factors",
					Required:     true,
					Enabled:      true,
					Dependencies: []string{"Data Preparation"},
					Config: map[string]interface{}{
						"factor_expressions": []string{
							"Ref($close, 1) / $close - 1",
							"Mean($close, 5) / $close - 1",
							"Mean($close, 20) / $close - 1",
							"Std($close, 20)",
						},
					},
				},
				{
					Name:         "Model Training",
					Type:         "model_training",
					Description:  "Train LightGBM prediction model",
					Required:     true,
					Enabled:      true,
					Dependencies: []string{"Factor Generation"},
					Config: map[string]interface{}{
						"model_type":    "lightgbm",
						"n_estimators":  100,
						"learning_rate": 0.1,
						"split_date":    "2022-01-01",
					},
				},
				{
					Name:         "Strategy Backtest",
					Type:         "strategy_backtest",
					Description:  "Execute Top-K stock selection strategy backtest",
					Required:     true,
					Enabled:      true,
					Dependencies: []string{"Model Training"},
					Config: map[string]interface{}{
						"top_k":          30,
						"rebalance_freq": "monthly",
					},
				},
				{
					Name:         "Result Analysis",
					Type:         "result_analysis",
					Description:  "Analyze strategy performance",
					Required:     false,
					Enabled:      true,
					Dependencies: []string{"Strategy Backtest"},
					Config:       map[string]interface{}{},
				},
			},
		},
		{
			ID:            "factor_research",
			Name:          "Factor Research Workflow",
			Description:   "Research workflow focused on factor mining and effectiveness analysis",
			Category:      "research",
			Icon:          "search",
			Tags:          []string{"factor research", "data analysis", "intermediate"},
			Difficulty:    "intermediate",
			EstimatedTime: "20-40min",
			Config: map[string]interface{}{
				"research_focus": "factor_mining",
				"analysis_depth": "detailed",
			},
			Steps: []ConfigStep{
				{
					Name:        "Data Preparation",
					Type:        "data_preparation",
					Description: "Prepare research data",
					Required:    true,
					Enabled:     true,
					Config: map[string]interface{}{
						"instruments": []string{"000300.SH"},
						"start_time":  "2019-01-01",
						"end_time":    "2023-12-31",
					},
				},
				{
					Name:         "Factor Generation",
					Type:         "factor_generation",
					Description:  "Generate multiple candidate factors",
					Required:     true,
					Enabled:      true,
					Dependencies: []string{"Data Preparation"},
					Config: map[string]interface{}{
						"factor_expressions": []string{
							"($close - Mean($close, 20)) / Std($close, 20)",
							"Corr($close, $volume, 20)",
							"Rank($volume) / Count($volume)",
						},
					},
				},
				{
					Name:         "Factor Analysis",
					Type:         "result_analysis",
					Description:  "Analyze factor effectiveness",
					Required:     true,
					Enabled:      true,
					Dependencies: []string{"Factor Generation"},
					Config:       map[string]interface{}{},
				},
			},
		},
	}
}
