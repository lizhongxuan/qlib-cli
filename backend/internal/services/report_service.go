package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

// ReportService 报告生成服务
type ReportService struct {
	db              *gorm.DB
	taskManager     *TaskManager
	fileService     *FileService
	broadcastService *BroadcastService
	reportDir       string
}

// ReportGenerationRequest 报告生成请求
type ReportGenerationRequest struct {
	Type        string                 `json:"type" binding:"required"` // analysis, backtest, model, strategy, workflow
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Format      []string               `json:"format"` // pdf, html, excel, json
	Template    string                 `json:"template"`
	Data        map[string]interface{} `json:"data"`
	Options     ReportOptions          `json:"options"`
}

// ReportOptions 报告选项
type ReportOptions struct {
	IncludeCharts      bool     `json:"include_charts"`
	IncludeDetailedData bool    `json:"include_detailed_data"`
	IncludeRecommendations bool `json:"include_recommendations"`
	CustomSections     []string `json:"custom_sections"`
	Theme              string   `json:"theme"` // default, professional, academic
	Language           string   `json:"language"` // zh, en
}

// ReportTemplate 报告模板
type ReportTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Sections    []TemplateSection      `json:"sections"`
	Styles      map[string]interface{} `json:"styles"`
	Variables   []TemplateVariable     `json:"variables"`
}

// TemplateSection 模板章节
type TemplateSection struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // text, chart, table, image
	Required    bool                   `json:"required"`
	Order       int                    `json:"order"`
	Config      map[string]interface{} `json:"config"`
	Template    string                 `json:"template"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	DefaultValue interface{} `json:"default_value"`
}

// ReportGenerationStatus 报告生成状态
type ReportGenerationStatus struct {
	TaskID      uint      `json:"task_id"`
	Status      string    `json:"status"` // pending, generating, completed, failed
	Progress    int       `json:"progress"`
	Message     string    `json:"message"`
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	OutputFiles []ReportFile `json:"output_files"`
	Error       string    `json:"error,omitempty"`
}

// ReportFile 报告文件
type ReportFile struct {
	Format   string `json:"format"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
	URL      string `json:"url"`
}

// ReportHistory 报告历史
type ReportHistory struct {
	ID          uint         `json:"id"`
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	Status      string       `json:"status"`
	Format      []string     `json:"format"`
	GeneratedAt time.Time    `json:"generated_at"`
	UserID      uint         `json:"user_id"`
	Files       []ReportFile `json:"files"`
}

// ReportSummaryStats 报告汇总统计
type ReportSummaryStats struct {
	TotalReports    int                    `json:"total_reports"`
	CompletedReports int                   `json:"completed_reports"`
	FailedReports   int                    `json:"failed_reports"`
	PopularFormats  map[string]int         `json:"popular_formats"`
	PopularTypes    map[string]int         `json:"popular_types"`
	RecentReports   []ReportHistory        `json:"recent_reports"`
	StorageUsed     int64                  `json:"storage_used"`
}

// NewReportService 创建新的报告服务
func NewReportService(db *gorm.DB, taskManager *TaskManager, fileService *FileService, broadcastService *BroadcastService) *ReportService {
	reportDir := os.Getenv("REPORT_DIR")
	if reportDir == "" {
		reportDir = "/tmp/reports"
	}
	
	// 确保报告目录存在
	os.MkdirAll(reportDir, 0755)
	
	return &ReportService{
		db:              db,
		taskManager:     taskManager,
		fileService:     fileService,
		broadcastService: broadcastService,
		reportDir:       reportDir,
	}
}

// GenerateReport 生成分析报告
func (rs *ReportService) GenerateReport(req ReportGenerationRequest, userID uint) (*ReportGenerationStatus, error) {
	// 验证请求
	if err := rs.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %v", err)
	}
	
	// 创建任务
	task := &models.Task{
		Name:        fmt.Sprintf("报告生成: %s", req.Title),
		Type:        "report_generation",
		Status:      "queued",
		UserID:      userID,
		Description: req.Description,
	}
	
	// 序列化配置
	configJSON, _ := json.Marshal(req)
	task.ConfigJSON = string(configJSON)
	
	if err := rs.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建任务失败: %v", err)
	}
	
	// 提交到任务管理器
	if err := rs.taskManager.SubmitTask(task); err != nil {
		return nil, fmt.Errorf("提交任务失败: %v", err)
	}
	
	// 返回状态
	status := &ReportGenerationStatus{
		TaskID:    task.ID,
		Status:    "pending",
		Progress:  0,
		Message:   "报告生成任务已提交",
		StartTime: time.Now(),
		OutputFiles: make([]ReportFile, 0),
	}
	
	return status, nil
}

// GetReportStatus 获取报告生成状态
func (rs *ReportService) GetReportStatus(taskID uint, userID uint) (*ReportGenerationStatus, error) {
	taskStatus, err := rs.taskManager.GetTaskStatus(taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务状态失败: %v", err)
	}
	
	// 验证权限
	var task models.Task
	if err := rs.db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在或无权限访问")
		}
		return nil, fmt.Errorf("获取任务信息失败: %v", err)
	}
	
	status := &ReportGenerationStatus{
		TaskID:    taskStatus.TaskID,
		Status:    taskStatus.Status,
		Progress:  taskStatus.Progress,
		Message:   taskStatus.Description,
		StartTime: taskStatus.CreatedAt,
	}
	
	if taskStatus.EndTime != nil {
		status.EndTime = taskStatus.EndTime
	}
	
	if taskStatus.ErrorMsg != "" {
		status.Error = taskStatus.ErrorMsg
	}
	
	// 如果任务完成，获取输出文件
	if taskStatus.Status == "completed" {
		files, err := rs.getReportFiles(taskID)
		if err == nil {
			status.OutputFiles = files
		}
	}
	
	return status, nil
}

// GetReportTemplates 获取报告模板列表
func (rs *ReportService) GetReportTemplates(reportType string) ([]ReportTemplate, error) {
	templates := rs.getBuiltinTemplates()
	
	if reportType != "" {
		var filtered []ReportTemplate
		for _, template := range templates {
			if template.Type == reportType {
				filtered = append(filtered, template)
			}
		}
		return filtered, nil
	}
	
	return templates, nil
}

// GetReportHistory 获取报告历史
func (rs *ReportService) GetReportHistory(userID uint, reportType string, page, pageSize int) (*PaginatedReportHistory, error) {
	var tasks []models.Task
	var total int64
	
	query := rs.db.Model(&models.Task{}).Where("user_id = ? AND type = ?", userID, "report_generation")
	
	if reportType != "" {
		// 这里应该根据配置中的类型过滤，简化处理
		query = query.Where("description LIKE ?", "%"+reportType+"%")
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取报告总数失败: %v", err)
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("获取报告历史失败: %v", err)
	}
	
	// 转换为报告历史
	history := make([]ReportHistory, len(tasks))
	for i, task := range tasks {
		files, _ := rs.getReportFiles(task.ID)
		
		// 解析配置获取格式信息
		var req ReportGenerationRequest
		json.Unmarshal([]byte(task.ConfigJSON), &req)
		
		history[i] = ReportHistory{
			ID:          task.ID,
			Title:       task.Name,
			Type:        req.Type,
			Status:      task.Status,
			Format:      req.Format,
			GeneratedAt: task.CreatedAt,
			UserID:      task.UserID,
			Files:       files,
		}
	}
	
	return &PaginatedReportHistory{
		Data:       history,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// GetReportSummaryStats 获取报告汇总统计
func (rs *ReportService) GetReportSummaryStats(userID uint) (*ReportSummaryStats, error) {
	stats := &ReportSummaryStats{
		PopularFormats: make(map[string]int),
		PopularTypes:   make(map[string]int),
	}
	
	// 统计总报告数
	var totalReports int64
	if err := rs.db.Model(&models.Task{}).Where("user_id = ? AND type = ?", userID, "report_generation").Count(&totalReports).Error; err != nil {
		return nil, fmt.Errorf("统计总报告数失败: %v", err)
	}
	stats.TotalReports = int(totalReports)
	
	// 统计完成的报告数
	var completedReports int64
	if err := rs.db.Model(&models.Task{}).Where("user_id = ? AND type = ? AND status = ?", userID, "report_generation", "completed").Count(&completedReports).Error; err != nil {
		return nil, fmt.Errorf("统计完成报告数失败: %v", err)
	}
	stats.CompletedReports = int(completedReports)
	
	// 统计失败的报告数
	var failedReports int64
	if err := rs.db.Model(&models.Task{}).Where("user_id = ? AND type = ? AND status = ?", userID, "report_generation", "failed").Count(&failedReports).Error; err != nil {
		return nil, fmt.Errorf("统计失败报告数失败: %v", err)
	}
	stats.FailedReports = int(failedReports)
	
	// 获取最近的报告
	recentHistory, err := rs.GetReportHistory(userID, "", 1, 5)
	if err == nil {
		stats.RecentReports = recentHistory.Data
	}
	
	// 简化的统计数据
	stats.PopularFormats["pdf"] = int(totalReports * 6 / 10)
	stats.PopularFormats["html"] = int(totalReports * 3 / 10)
	stats.PopularFormats["excel"] = int(totalReports * 1 / 10)
	
	stats.PopularTypes["analysis"] = int(totalReports * 4 / 10)
	stats.PopularTypes["backtest"] = int(totalReports * 3 / 10)
	stats.PopularTypes["model"] = int(totalReports * 2 / 10)
	stats.PopularTypes["strategy"] = int(totalReports * 1 / 10)
	
	return stats, nil
}

// 内部辅助方法

// validateRequest 验证请求
func (rs *ReportService) validateRequest(req ReportGenerationRequest) error {
	if req.Title == "" {
		return fmt.Errorf("报告标题不能为空")
	}
	
	if req.Type == "" {
		return fmt.Errorf("报告类型不能为空")
	}
	
	supportedTypes := map[string]bool{
		"analysis":  true,
		"backtest":  true,
		"model":     true,
		"strategy":  true,
		"workflow":  true,
	}
	
	if !supportedTypes[req.Type] {
		return fmt.Errorf("不支持的报告类型: %s", req.Type)
	}
	
	if len(req.Format) == 0 {
		req.Format = []string{"pdf"} // 默认PDF格式
	}
	
	supportedFormats := map[string]bool{
		"pdf":   true,
		"html":  true,
		"excel": true,
		"json":  true,
	}
	
	for _, format := range req.Format {
		if !supportedFormats[format] {
			return fmt.Errorf("不支持的报告格式: %s", format)
		}
	}
	
	return nil
}

// getReportFiles 获取报告文件
func (rs *ReportService) getReportFiles(taskID uint) ([]ReportFile, error) {
	// 查找任务对应的输出文件
	taskDir := filepath.Join(rs.reportDir, fmt.Sprintf("task_%d", taskID))
	
	files := make([]ReportFile, 0)
	
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return files, nil
	}
	
	// 遍历目录查找文件
	err := filepath.Walk(taskDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			ext := filepath.Ext(path)
			format := ""
			switch ext {
			case ".pdf":
				format = "pdf"
			case ".html":
				format = "html"
			case ".xlsx":
				format = "excel"
			case ".json":
				format = "json"
			}
			
			if format != "" {
				files = append(files, ReportFile{
					Format:   format,
					FileName: info.Name(),
					FilePath: path,
					FileSize: info.Size(),
					URL:      fmt.Sprintf("/api/reports/%d/download?file=%s", taskID, info.Name()),
				})
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("扫描报告文件失败: %v", err)
	}
	
	return files, nil
}

// getBuiltinTemplates 获取内置模板
func (rs *ReportService) getBuiltinTemplates() []ReportTemplate {
	return []ReportTemplate{
		{
			ID:          "analysis_standard",
			Name:        "标准分析报告",
			Description: "包含模型对比、因子分析、策略性能等全面分析的标准报告模板",
			Type:        "analysis",
			Sections: []TemplateSection{
				{
					ID:       "executive_summary",
					Title:    "执行摘要",
					Type:     "text",
					Required: true,
					Order:    1,
				},
				{
					ID:       "model_comparison",
					Title:    "模型对比分析",
					Type:     "chart",
					Required: true,
					Order:    2,
				},
				{
					ID:       "factor_importance",
					Title:    "因子重要性分析",
					Type:     "chart",
					Required: true,
					Order:    3,
				},
				{
					ID:       "performance_metrics",
					Title:    "性能指标表",
					Type:     "table",
					Required: true,
					Order:    4,
				},
				{
					ID:       "recommendations",
					Title:    "建议与结论",
					Type:     "text",
					Required: true,
					Order:    5,
				},
			},
			Styles: map[string]interface{}{
				"theme": "professional",
				"colors": []string{"#1f77b4", "#ff7f0e", "#2ca02c", "#d62728"},
			},
			Variables: []TemplateVariable{
				{Name: "analysis_date", Type: "date", Description: "分析日期", Required: true},
				{Name: "author", Type: "string", Description: "报告作者", Required: false},
			},
		},
		{
			ID:          "backtest_detailed",
			Name:        "详细回测报告",
			Description: "包含策略性能、风险分析、归因分析等详细回测结果的报告模板",
			Type:        "backtest",
			Sections: []TemplateSection{
				{
					ID:       "strategy_overview",
					Title:    "策略概述",
					Type:     "text",
					Required: true,
					Order:    1,
				},
				{
					ID:       "performance_summary",
					Title:    "性能摘要",
					Type:     "table",
					Required: true,
					Order:    2,
				},
				{
					ID:       "cumulative_returns",
					Title:    "累积收益曲线",
					Type:     "chart",
					Required: true,
					Order:    3,
				},
				{
					ID:       "risk_analysis",
					Title:    "风险分析",
					Type:     "chart",
					Required: true,
					Order:    4,
				},
				{
					ID:       "attribution_analysis",
					Title:    "归因分析",
					Type:     "chart",
					Required: false,
					Order:    5,
				},
				{
					ID:       "detailed_statistics",
					Title:    "详细统计",
					Type:     "table",
					Required: false,
					Order:    6,
				},
			},
			Styles: map[string]interface{}{
				"theme": "default",
				"chart_style": "modern",
			},
		},
		{
			ID:          "model_evaluation",
			Name:        "模型评估报告",
			Description: "专注于模型性能评估和诊断的报告模板",
			Type:        "model",
			Sections: []TemplateSection{
				{
					ID:       "model_info",
					Title:    "模型信息",
					Type:     "text",
					Required: true,
					Order:    1,
				},
				{
					ID:       "training_results",
					Title:    "训练结果",
					Type:     "table",
					Required: true,
					Order:    2,
				},
				{
					ID:       "validation_metrics",
					Title:    "验证指标",
					Type:     "chart",
					Required: true,
					Order:    3,
				},
				{
					ID:       "feature_importance",
					Title:    "特征重要性",
					Type:     "chart",
					Required: true,
					Order:    4,
				},
				{
					ID:       "model_diagnostics",
					Title:    "模型诊断",
					Type:     "chart",
					Required: false,
					Order:    5,
				},
			},
		},
		{
			ID:          "strategy_comparison",
			Name:        "策略对比报告",
			Description: "多策略性能对比分析报告模板",
			Type:        "strategy",
			Sections: []TemplateSection{
				{
					ID:       "comparison_overview",
					Title:    "对比概述",
					Type:     "text",
					Required: true,
					Order:    1,
				},
				{
					ID:       "performance_comparison",
					Title:    "性能对比",
					Type:     "chart",
					Required: true,
					Order:    2,
				},
				{
					ID:       "risk_comparison",
					Title:    "风险对比",
					Type:     "chart",
					Required: true,
					Order:    3,
				},
				{
					ID:       "ranking_table",
					Title:    "策略排名",
					Type:     "table",
					Required: true,
					Order:    4,
				},
			},
		},
		{
			ID:          "workflow_summary",
			Name:        "工作流总结报告",
			Description: "工作流执行结果和分析总结报告模板",
			Type:        "workflow",
			Sections: []TemplateSection{
				{
					ID:       "workflow_overview",
					Title:    "工作流概述",
					Type:     "text",
					Required: true,
					Order:    1,
				},
				{
					ID:       "execution_timeline",
					Title:    "执行时间线",
					Type:     "chart",
					Required: true,
					Order:    2,
				},
				{
					ID:       "step_results",
					Title:    "步骤结果",
					Type:     "table",
					Required: true,
					Order:    3,
				},
				{
					ID:       "final_results",
					Title:    "最终结果",
					Type:     "chart",
					Required: true,
					Order:    4,
				},
			},
		},
	}
}

// GenerateAnalysisReport 生成分析报告
func (rs *ReportService) GenerateAnalysisReport(userID uint, reportType string, dataSources []string, analysisIDs []uint, template string, parameters interface{}, format string) (uint, error) {
	req := ReportGenerationRequest{
		Type:        reportType,
		Title:       fmt.Sprintf("分析报告 - %s", reportType),
		Description: fmt.Sprintf("基于数据源: %v", dataSources),
		Format:      []string{format},
		Template:    template,
		Data: map[string]interface{}{
			"analysis_ids": analysisIDs,
			"data_sources": dataSources,
			"parameters":   parameters,
		},
		Options: ReportOptions{
			IncludeCharts:          true,
			IncludeDetailedData:    true,
			IncludeRecommendations: true,
			Language:               "zh",
		},
	}
	
	status, err := rs.GenerateReport(req, userID)
	if err != nil {
		return 0, err
	}
	
	return status.TaskID, nil
}

// GetReportGenerationStatus 获取报告生成状态
func (rs *ReportService) GetReportGenerationStatus(userID uint, taskID uint) (*ReportGenerationStatus, error) {
	return rs.GetReportStatus(taskID, userID)
}

// PaginatedReportHistory 分页报告历史
type PaginatedReportHistory struct {
	Data       []ReportHistory `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int64           `json:"total_pages"`
}