package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"qlib-backend/internal/models"
	"qlib-backend/internal/qlib"

	"gorm.io/gorm"
)

// WorkflowService 工作流服务
type WorkflowService struct {
	db              *gorm.DB
	workflowEngine  *qlib.WorkflowEngine
	taskManager     *TaskManager
	broadcastService *BroadcastService
	runningWorkflows map[uint]*WorkflowExecution
	mutex           sync.RWMutex
}

// WorkflowExecution 工作流执行实例
type WorkflowExecution struct {
	WorkflowID      uint
	TaskID          uint
	Status          string
	CurrentStep     string
	Progress        int
	StartTime       time.Time
	EndTime         *time.Time
	Results         map[string]interface{}
	Error           string
	Context         context.Context
	Cancel          context.CancelFunc
}

// WorkflowTemplate 工作流模板
type WorkflowTemplate struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Config      map[string]interface{} `json:"config"`
	Steps       []WorkflowStep         `json:"steps"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Dependencies []string              `json:"dependencies"`
	Required    bool                   `json:"required"`
}

// WorkflowRunRequest 工作流运行请求
type WorkflowRunRequest struct {
	TemplateID uint                   `json:"template_id"`
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"`
	UserID     uint                   `json:"user_id"`
}

// WorkflowHistory 工作流历史记录
type WorkflowHistory struct {
	ID           uint                   `json:"id"`
	WorkflowID   uint                   `json:"workflow_id"`
	Name         string                 `json:"name"`
	Status       string                 `json:"status"`
	Progress     int                    `json:"progress"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	Duration     *time.Duration         `json:"duration,omitempty"`
	Results      map[string]interface{} `json:"results"`
	Error        string                 `json:"error"`
	UserID       uint                   `json:"user_id"`
	CreatedAt    time.Time              `json:"created_at"`
}

// NewWorkflowService 创建新的工作流服务
func NewWorkflowService(db *gorm.DB, taskManager *TaskManager, broadcastService *BroadcastService) *WorkflowService {
	workflowEngine := qlib.NewWorkflowEngine(db)
	
	return &WorkflowService{
		db:               db,
		workflowEngine:   workflowEngine,
		taskManager:      taskManager,
		broadcastService: broadcastService,
		runningWorkflows: make(map[uint]*WorkflowExecution),
	}
}

// RunWorkflow 运行完整工作流
func (ws *WorkflowService) RunWorkflow(req WorkflowRunRequest) (*WorkflowExecution, error) {
	// 获取模板
	template, err := ws.GetTemplate(req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("获取工作流模板失败: %v", err)
	}

	// 创建工作流记录
	workflow := &models.Workflow{
		Name:         req.Name,
		TemplateID:   req.TemplateID,
		Status:       "queued",
		Progress:     0,
		ConfigJSON:   ws.mapToJSON(req.Config),
		UserID:       req.UserID,
	}

	if err := ws.db.Create(workflow).Error; err != nil {
		return nil, fmt.Errorf("创建工作流记录失败: %v", err)
	}

	// 创建任务
	task := &models.Task{
		Name:        fmt.Sprintf("工作流执行: %s", req.Name),
		Type:        "workflow_execution",
		Status:      "queued",
		UserID:      req.UserID,
		Description: fmt.Sprintf("执行工作流模板: %s", template.Name),
	}

	if err := ws.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建任务失败: %v", err)
	}

	// 创建执行实例
	ctx, cancel := context.WithCancel(context.Background())
	execution := &WorkflowExecution{
		WorkflowID:  workflow.ID,
		TaskID:      task.ID,
		Status:      "queued",
		CurrentStep: "初始化",
		Progress:    0,
		StartTime:   time.Now(),
		Results:     make(map[string]interface{}),
		Context:     ctx,
		Cancel:      cancel,
	}

	ws.mutex.Lock()
	ws.runningWorkflows[workflow.ID] = execution
	ws.mutex.Unlock()

	// 异步执行工作流
	go ws.executeWorkflow(execution, template, req.Config)

	return execution, nil
}

// executeWorkflow 执行工作流
func (ws *WorkflowService) executeWorkflow(execution *WorkflowExecution, template *WorkflowTemplate, config map[string]interface{}) {
	defer func() {
		ws.mutex.Lock()
		delete(ws.runningWorkflows, execution.WorkflowID)
		ws.mutex.Unlock()
	}()

	// 更新状态为运行中
	execution.Status = "running"
	execution.StartTime = time.Now()
	
	ws.updateWorkflowStatus(execution.WorkflowID, "running", 0, "工作流开始执行")

	// 转换模板为qlib格式
	qlibTemplate := ws.convertToQlibTemplate(template)
	
	// 执行工作流
	results, err := ws.workflowEngine.Execute(execution.Context, qlibTemplate, config, func(step string, progress int, message string) {
		execution.CurrentStep = step
		execution.Progress = progress
		
		ws.updateWorkflowStatus(execution.WorkflowID, "running", progress, message)
		
		// 广播进度更新
		if ws.broadcastService != nil {
			workflow := &models.Workflow{}
			ws.db.First(workflow, execution.WorkflowID)
			ws.broadcastService.PublishWorkflowProgress(workflow.UserID, execution.WorkflowID, progress, step)
		}
	})

	endTime := time.Now()
	execution.EndTime = &endTime

	if err != nil {
		// 执行失败
		execution.Status = "failed"
		execution.Error = err.Error()
		
		ws.updateWorkflowStatus(execution.WorkflowID, "failed", execution.Progress, err.Error())
		
		// 广播失败事件
		if ws.broadcastService != nil {
			workflow := &models.Workflow{}
			ws.db.First(workflow, execution.WorkflowID)
			ws.broadcastService.Publish(Event{
				Type:     "workflow_failed",
				Category: "workflow",
				UserID:   workflow.UserID,
				Data: map[string]interface{}{
					"workflow_id": execution.WorkflowID,
					"error":       err.Error(),
				},
			})
		}
	} else {
		// 执行成功
		execution.Status = "completed"
		execution.Progress = 100
		execution.Results = results
		
		ws.updateWorkflowStatus(execution.WorkflowID, "completed", 100, "工作流执行完成")
		
		// 保存结果
		resultJSON, _ := json.Marshal(results)
		ws.db.Model(&models.Workflow{}).Where("id = ?", execution.WorkflowID).Updates(map[string]interface{}{
			"result_json": string(resultJSON),
		})
		
		// 广播完成事件
		if ws.broadcastService != nil {
			workflow := &models.Workflow{}
			ws.db.First(workflow, execution.WorkflowID)
			ws.broadcastService.Publish(Event{
				Type:     "workflow_completed",
				Category: "workflow",
				UserID:   workflow.UserID,
				Data: map[string]interface{}{
					"workflow_id": execution.WorkflowID,
					"results":     results,
				},
			})
		}
	}

	// 更新任务状态
	taskStatus := "completed"
	if err != nil {
		taskStatus = "failed"
	}
	
	ws.db.Model(&models.Task{}).Where("id = ?", execution.TaskID).Updates(map[string]interface{}{
		"status":   taskStatus,
		"progress": execution.Progress,
		"end_time": endTime,
	})
}

// updateWorkflowStatus 更新工作流状态
func (ws *WorkflowService) updateWorkflowStatus(workflowID uint, status string, progress int, message string) {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}
	
	if status == "running" && progress == 0 {
		updates["start_time"] = time.Now()
	} else if status == "completed" || status == "failed" {
		updates["end_time"] = time.Now()
	}
	
	ws.db.Model(&models.Workflow{}).Where("id = ?", workflowID).Updates(updates)
}

// GetTemplates 获取工作流模板列表
func (ws *WorkflowService) GetTemplates(category string) ([]WorkflowTemplate, error) {
	var templates []models.WorkflowTemplate
	query := ws.db.Model(&models.WorkflowTemplate{})
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	if err := query.Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("获取工作流模板失败: %v", err)
	}
	
	result := make([]WorkflowTemplate, len(templates))
	for i, t := range templates {
		result[i] = WorkflowTemplate{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category,
			Config:      ws.jsonToMap(t.ConfigJSON),
			Steps:       ws.parseSteps(t.StepsJSON),
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		}
	}
	
	return result, nil
}

// GetTemplate 获取单个工作流模板
func (ws *WorkflowService) GetTemplate(templateID uint) (*WorkflowTemplate, error) {
	var template models.WorkflowTemplate
	if err := ws.db.First(&template, templateID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("工作流模板不存在")
		}
		return nil, fmt.Errorf("获取工作流模板失败: %v", err)
	}
	
	return &WorkflowTemplate{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		Category:    template.Category,
		Config:      ws.jsonToMap(template.ConfigJSON),
		Steps:       ws.parseSteps(template.StepsJSON),
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}, nil
}

// CreateTemplate 创建工作流模板
func (ws *WorkflowService) CreateTemplate(template WorkflowTemplate, userID uint) (*WorkflowTemplate, error) {
	stepsJSON, _ := json.Marshal(template.Steps)
	configJSON, _ := json.Marshal(template.Config)
	
	dbTemplate := &models.WorkflowTemplate{
		Name:        template.Name,
		Description: template.Description,
		Category:    template.Category,
		ConfigJSON:  string(configJSON),
		StepsJSON:   string(stepsJSON),
		CreatedBy:   userID,
	}
	
	if err := ws.db.Create(dbTemplate).Error; err != nil {
		return nil, fmt.Errorf("创建工作流模板失败: %v", err)
	}
	
	template.ID = dbTemplate.ID
	template.CreatedAt = dbTemplate.CreatedAt
	template.UpdatedAt = dbTemplate.UpdatedAt
	
	return &template, nil
}

// GetWorkflowStatus 获取工作流状态
func (ws *WorkflowService) GetWorkflowStatus(workflowID uint) (*WorkflowExecution, error) {
	// 先检查运行中的工作流
	ws.mutex.RLock()
	if execution, exists := ws.runningWorkflows[workflowID]; exists {
		ws.mutex.RUnlock()
		return execution, nil
	}
	ws.mutex.RUnlock()
	
	// 从数据库获取已完成的工作流
	var workflow models.Workflow
	if err := ws.db.First(&workflow, workflowID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("工作流不存在")
		}
		return nil, fmt.Errorf("获取工作流失败: %v", err)
	}
	
	execution := &WorkflowExecution{
		WorkflowID:  workflow.ID,
		Status:      workflow.Status,
		Progress:    workflow.Progress,
		StartTime:   *workflow.StartTime,
		EndTime:     workflow.EndTime,
		Results:     ws.jsonToMap(workflow.ResultJSON),
		Error:       workflow.ErrorMsg,
	}
	
	return execution, nil
}

// PauseWorkflow 暂停工作流
func (ws *WorkflowService) PauseWorkflow(workflowID uint) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	
	execution, exists := ws.runningWorkflows[workflowID]
	if !exists {
		return fmt.Errorf("工作流不存在或已完成")
	}
	
	if execution.Status != "running" {
		return fmt.Errorf("工作流状态不允许暂停")
	}
	
	execution.Status = "paused"
	ws.updateWorkflowStatus(workflowID, "paused", execution.Progress, "工作流已暂停")
	
	return nil
}

// ResumeWorkflow 恢复工作流
func (ws *WorkflowService) ResumeWorkflow(workflowID uint) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	
	execution, exists := ws.runningWorkflows[workflowID]
	if !exists {
		return fmt.Errorf("工作流不存在或已完成")
	}
	
	if execution.Status != "paused" {
		return fmt.Errorf("工作流状态不允许恢复")
	}
	
	execution.Status = "running"
	ws.updateWorkflowStatus(workflowID, "running", execution.Progress, "工作流已恢复")
	
	return nil
}

// GetWorkflowHistory 获取工作流历史
func (ws *WorkflowService) GetWorkflowHistory(userID uint, page, pageSize int) (*PaginatedWorkflowHistory, error) {
	var workflows []models.Workflow
	var total int64
	
	query := ws.db.Model(&models.Workflow{}).Where("user_id = ?", userID)
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取工作流总数失败: %v", err)
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("获取工作流历史失败: %v", err)
	}
	
	history := make([]WorkflowHistory, len(workflows))
	for i, w := range workflows {
		history[i] = WorkflowHistory{
			ID:         w.ID,
			WorkflowID: w.ID,
			Name:       w.Name,
			Status:     w.Status,
			Progress:   w.Progress,
			StartTime:  *w.StartTime,
			EndTime:    w.EndTime,
			Results:    ws.jsonToMap(w.ResultJSON),
			Error:      w.ErrorMsg,
			UserID:     w.UserID,
			CreatedAt:  w.CreatedAt,
		}
		
		if w.StartTime != nil && w.EndTime != nil {
			duration := w.EndTime.Sub(*w.StartTime)
			history[i].Duration = &duration
		}
	}
	
	return &PaginatedWorkflowHistory{
		Data:       history,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// 辅助方法

func (ws *WorkflowService) mapToJSON(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

func (ws *WorkflowService) jsonToMap(jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return make(map[string]interface{})
	}
	
	var data map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &data)
	return data
}

func (ws *WorkflowService) parseSteps(stepsJSON string) []WorkflowStep {
	if stepsJSON == "" {
		return []WorkflowStep{}
	}
	
	var steps []WorkflowStep
	json.Unmarshal([]byte(stepsJSON), &steps)
	return steps
}

// 数据结构定义

type PaginatedWorkflowHistory struct {
	Data       []WorkflowHistory `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int64             `json:"total_pages"`
}

// 辅助方法

// convertToQlibTemplate 将服务层的WorkflowTemplate转换为qlib.WorkflowTemplate
func (ws *WorkflowService) convertToQlibTemplate(template *WorkflowTemplate) *qlib.WorkflowTemplate {
	qlibSteps := make([]qlib.WorkflowStep, len(template.Steps))
	for i, step := range template.Steps {
		qlibSteps[i] = qlib.WorkflowStep{
			Name:         step.Name,
			Type:         step.Type,
			Description:  step.Description,
			Config:       step.Config,
			Dependencies: step.Dependencies,
			Required:     step.Required,
		}
	}
	
	return &qlib.WorkflowTemplate{
		ID:          template.ID,
		Name:        template.Name,
		Description: template.Description,
		Category:    template.Category,
		Config:      template.Config,
		Steps:       qlibSteps,
	}
}