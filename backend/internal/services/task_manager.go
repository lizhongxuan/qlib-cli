package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

// TaskManager 任务管理器
type TaskManager struct {
	db            *gorm.DB
	runningTasks  map[uint]*TaskContext
	taskQueue     chan *TaskContext
	workers       int
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// TaskContext 任务上下文
type TaskContext struct {
	Task        *models.Task
	Cancel      context.CancelFunc
	ProgressCh  chan TaskProgress
	StatusCh    chan TaskStatus
	ErrorCh     chan error
	CompleteCh  chan TaskResult
}

// TaskProgress 任务进度
type TaskProgress struct {
	TaskID   uint    `json:"task_id"`
	Progress int     `json:"progress"`
	Message  string  `json:"message"`
	Details  map[string]interface{} `json:"details"`
}

// TaskStatus 任务状态
type TaskStatus struct {
	TaskID    uint   `json:"task_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    uint                   `json:"task_id"`
	Success   bool                   `json:"success"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error"`
	Duration  time.Duration          `json:"duration"`
}

// TaskHandler 任务处理函数类型
type TaskHandler func(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error)

// NewTaskManager 创建新的任务管理器
func NewTaskManager(db *gorm.DB, workers int) *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	tm := &TaskManager{
		db:           db,
		runningTasks: make(map[uint]*TaskContext),
		taskQueue:    make(chan *TaskContext, 100),
		workers:      workers,
		ctx:          ctx,
		cancel:       cancel,
	}
	
	// 启动工作协程
	for i := 0; i < workers; i++ {
		go tm.worker()
	}
	
	// 启动任务状态更新协程
	go tm.statusUpdater()
	
	return tm
}

// SubmitTask 提交任务
func (tm *TaskManager) SubmitTask(task *models.Task) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	// 更新任务状态为排队
	if err := tm.db.Model(task).Updates(map[string]interface{}{
		"status": "queued",
	}).Error; err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}
	
	// 创建任务上下文
	_, cancel := context.WithCancel(tm.ctx)
	taskCtx := &TaskContext{
		Task:       task,
		Cancel:     cancel,
		ProgressCh: make(chan TaskProgress, 10),
		StatusCh:   make(chan TaskStatus, 10),
		ErrorCh:    make(chan error, 1),
		CompleteCh: make(chan TaskResult, 1),
	}
	
	tm.runningTasks[task.ID] = taskCtx
	
	// 提交到任务队列
	select {
	case tm.taskQueue <- taskCtx:
		return nil
	case <-tm.ctx.Done():
		return fmt.Errorf("任务管理器已关闭")
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID uint) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	taskCtx, exists := tm.runningTasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在或已完成")
	}
	
	// 取消任务上下文
	taskCtx.Cancel()
	
	// 更新数据库状态
	if err := tm.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":   "cancelled",
		"end_time": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}
	
	// 从运行任务列表中移除
	delete(tm.runningTasks, taskID)
	
	return nil
}

// GetTaskStatus 获取任务状态
func (tm *TaskManager) GetTaskStatus(taskID uint) (*TaskStatusInfo, error) {
	var task models.Task
	if err := tm.db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在")
		}
		return nil, fmt.Errorf("获取任务失败: %v", err)
	}
	
	tm.mutex.RLock()
	_, isRunning := tm.runningTasks[taskID]
	tm.mutex.RUnlock()
	
	status := &TaskStatusInfo{
		TaskID:      task.ID,
		Name:        task.Name,
		Type:        task.Type,
		Status:      task.Status,
		Progress:    task.Progress,
		Description: task.Description,
		CreatedAt:   task.CreatedAt,
		StartTime:   task.StartTime,
		EndTime:     task.EndTime,
		ErrorMsg:    task.ErrorMsg,
		IsRunning:   isRunning,
	}
	
	if task.StartTime != nil {
		if task.EndTime != nil {
			status.Duration = task.EndTime.Sub(*task.StartTime)
		} else {
			status.Duration = time.Since(*task.StartTime)
		}
	}
	
	return status, nil
}

// GetTasks 获取任务列表
func (tm *TaskManager) GetTasks(userID uint, status string, taskType string, page, pageSize int) (*PaginatedTasks, error) {
	var tasks []models.Task
	var total int64
	
	query := tm.db.Model(&models.Task{}).Where("user_id = ?", userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取任务总数失败: %v", err)
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %v", err)
	}
	
	return &PaginatedTasks{
		Data:       tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// GetRunningTasks 获取正在运行的任务
func (tm *TaskManager) GetRunningTasks() []TaskStatusInfo {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	var runningTasks []TaskStatusInfo
	for _, taskCtx := range tm.runningTasks {
		task := taskCtx.Task
		status := TaskStatusInfo{
			TaskID:      task.ID,
			Name:        task.Name,
			Type:        task.Type,
			Status:      task.Status,
			Progress:    task.Progress,
			Description: task.Description,
			CreatedAt:   task.CreatedAt,
			StartTime:   task.StartTime,
			IsRunning:   true,
		}
		
		if task.StartTime != nil {
			status.Duration = time.Since(*task.StartTime)
		}
		
		runningTasks = append(runningTasks, status)
	}
	
	return runningTasks
}

// worker 工作协程
func (tm *TaskManager) worker() {
	for {
		select {
		case taskCtx := <-tm.taskQueue:
			tm.executeTask(taskCtx)
		case <-tm.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务
func (tm *TaskManager) executeTask(taskCtx *TaskContext) {
	task := taskCtx.Task
	startTime := time.Now()
	
	// 更新任务状态为运行中
	tm.db.Model(task).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": startTime,
		"progress":   0,
	})
	
	// 发送状态更新
	taskCtx.StatusCh <- TaskStatus{
		TaskID:    task.ID,
		Status:    "running",
		Message:   "任务开始执行",
		Timestamp: startTime,
	}
	
	// 获取任务处理器
	handler := tm.getTaskHandler(task.Type)
	if handler == nil {
		tm.completeTaskWithError(taskCtx, fmt.Errorf("不支持的任务类型: %s", task.Type))
		return
	}
	
	// 创建子上下文
	ctx, cancel := context.WithCancel(tm.ctx)
	defer cancel()
	
	// 执行任务
	result, err := handler(ctx, task, taskCtx.ProgressCh)
	
	// 完成任务
	if err != nil {
		tm.completeTaskWithError(taskCtx, err)
	} else {
		tm.completeTaskWithSuccess(taskCtx, result)
	}
}

// completeTaskWithSuccess 成功完成任务
func (tm *TaskManager) completeTaskWithSuccess(taskCtx *TaskContext, result *TaskResult) {
	task := taskCtx.Task
	endTime := time.Now()
	
	resultJSON, _ := json.Marshal(result.Result)
	
	tm.db.Model(task).Updates(map[string]interface{}{
		"status":      "completed",
		"progress":    100,
		"end_time":    endTime,
		"result_json": string(resultJSON),
	})
	
	taskCtx.StatusCh <- TaskStatus{
		TaskID:    task.ID,
		Status:    "completed",
		Message:   "任务成功完成",
		Timestamp: endTime,
	}
	
	taskCtx.CompleteCh <- *result
	
	tm.cleanupTask(task.ID)
}

// completeTaskWithError 错误完成任务
func (tm *TaskManager) completeTaskWithError(taskCtx *TaskContext, err error) {
	task := taskCtx.Task
	endTime := time.Now()
	
	tm.db.Model(task).Updates(map[string]interface{}{
		"status":    "failed",
		"end_time":  endTime,
		"error_msg": err.Error(),
	})
	
	taskCtx.StatusCh <- TaskStatus{
		TaskID:    task.ID,
		Status:    "failed",
		Message:   err.Error(),
		Timestamp: endTime,
	}
	
	taskCtx.ErrorCh <- err
	
	tm.cleanupTask(task.ID)
}

// cleanupTask 清理任务
func (tm *TaskManager) cleanupTask(taskID uint) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	if taskCtx, exists := tm.runningTasks[taskID]; exists {
		close(taskCtx.ProgressCh)
		close(taskCtx.StatusCh)
		close(taskCtx.ErrorCh)
		close(taskCtx.CompleteCh)
		delete(tm.runningTasks, taskID)
	}
}

// statusUpdater 状态更新协程
func (tm *TaskManager) statusUpdater() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			tm.updateTaskProgress()
		case <-tm.ctx.Done():
			return
		}
	}
}

// updateTaskProgress 更新任务进度
func (tm *TaskManager) updateTaskProgress() {
	tm.mutex.RLock()
	tasks := make([]*TaskContext, 0, len(tm.runningTasks))
	for _, taskCtx := range tm.runningTasks {
		tasks = append(tasks, taskCtx)
	}
	tm.mutex.RUnlock()
	
	for _, taskCtx := range tasks {
		select {
		case progress := <-taskCtx.ProgressCh:
			tm.db.Model(&models.Task{}).Where("id = ?", progress.TaskID).Updates(map[string]interface{}{
				"progress": progress.Progress,
			})
		default:
			// 没有新的进度更新
		}
	}
}

// getTaskHandler 获取任务处理器
func (tm *TaskManager) getTaskHandler(taskType string) TaskHandler {
	handlers := map[string]TaskHandler{
		"model_training":      tm.handleModelTraining,
		"strategy_backtest":   tm.handleStrategyBacktest,
		"factor_test":         tm.handleFactorTest,
		"data_processing":     tm.handleDataProcessing,
		"report_generation":   tm.handleReportGeneration,
		"workflow_execution":  tm.handleWorkflowExecution,
	}
	
	return handlers[taskType]
}

// 任务处理器实现

// handleModelTraining 处理模型训练任务
func (tm *TaskManager) handleModelTraining(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error) {
	// 这里应该调用实际的模型训练逻辑
	// 为了演示，这里使用模拟实现
	
	for i := 0; i <= 100; i += 10 {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务被取消")
		default:
		}
		
		progressCh <- TaskProgress{
			TaskID:   task.ID,
			Progress: i,
			Message:  fmt.Sprintf("模型训练进度: %d%%", i),
			Details: map[string]interface{}{
				"epoch": i / 10,
				"loss":  0.5 - float64(i)/200,
			},
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Result: map[string]interface{}{
			"model_path": "/models/trained_model.pkl",
			"accuracy":   0.95,
			"loss":       0.05,
		},
		Duration: time.Since(*task.StartTime),
	}, nil
}

// handleStrategyBacktest 处理策略回测任务
func (tm *TaskManager) handleStrategyBacktest(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error) {
	steps := []string{"数据加载", "策略初始化", "回测执行", "结果计算", "报告生成"}
	
	for i, step := range steps {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务被取消")
		default:
		}
		
		progress := (i + 1) * 20
		progressCh <- TaskProgress{
			TaskID:   task.ID,
			Progress: progress,
			Message:  fmt.Sprintf("正在执行: %s", step),
			Details: map[string]interface{}{
				"current_step": step,
				"step_index":   i + 1,
				"total_steps":  len(steps),
			},
		}
		
		time.Sleep(2 * time.Second)
	}
	
	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Result: map[string]interface{}{
			"total_return":  0.156,
			"sharpe_ratio":  1.35,
			"max_drawdown": 0.082,
		},
		Duration: time.Since(*task.StartTime),
	}, nil
}

// handleFactorTest 处理因子测试任务
func (tm *TaskManager) handleFactorTest(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error) {
	for i := 0; i <= 100; i += 20 {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务被取消")
		default:
		}
		
		progressCh <- TaskProgress{
			TaskID:   task.ID,
			Progress: i,
			Message:  fmt.Sprintf("因子测试进度: %d%%", i),
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Result: map[string]interface{}{
			"ic":       0.085,
			"rank_ic":  0.078,
			"turnover": 0.25,
		},
		Duration: time.Since(*task.StartTime),
	}, nil
}

// handleDataProcessing 处理数据处理任务
func (tm *TaskManager) handleDataProcessing(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error) {
	stages := []string{"数据读取", "数据清洗", "特征工程", "数据验证", "数据保存"}
	
	for i, stage := range stages {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务被取消")
		default:
		}
		
		progress := (i + 1) * 20
		progressCh <- TaskProgress{
			TaskID:   task.ID,
			Progress: progress,
			Message:  fmt.Sprintf("正在执行: %s", stage),
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Result: map[string]interface{}{
			"processed_records": 100000,
			"output_path":       "/data/processed_data.csv",
		},
		Duration: time.Since(*task.StartTime),
	}, nil
}

// handleReportGeneration 处理报告生成任务
func (tm *TaskManager) handleReportGeneration(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error) {
	for i := 0; i <= 100; i += 25 {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务被取消")
		default:
		}
		
		progressCh <- TaskProgress{
			TaskID:   task.ID,
			Progress: i,
			Message:  fmt.Sprintf("报告生成进度: %d%%", i),
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Result: map[string]interface{}{
			"report_path": "/reports/analysis_report.pdf",
			"pages":       25,
		},
		Duration: time.Since(*task.StartTime),
	}, nil
}

// handleWorkflowExecution 处理工作流执行任务
func (tm *TaskManager) handleWorkflowExecution(ctx context.Context, task *models.Task, progressCh chan<- TaskProgress) (*TaskResult, error) {
	workflow_steps := []string{"初始化", "数据准备", "模型训练", "策略回测", "结果分析"}
	
	for i, step := range workflow_steps {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("任务被取消")
		default:
		}
		
		progress := (i + 1) * 20
		progressCh <- TaskProgress{
			TaskID:   task.ID,
			Progress: progress,
			Message:  fmt.Sprintf("工作流步骤: %s", step),
		}
		
		time.Sleep(3 * time.Second)
	}
	
	return &TaskResult{
		TaskID:   task.ID,
		Success:  true,
		Result: map[string]interface{}{
			"workflow_id":     task.ID,
			"completed_steps": len(workflow_steps),
			"output_files":    []string{"/output/model.pkl", "/output/backtest_results.json"},
		},
		Duration: time.Since(*task.StartTime),
	}, nil
}

// Close 关闭任务管理器
func (tm *TaskManager) Close() {
	tm.cancel()
	
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	// 取消所有运行中的任务
	for _, taskCtx := range tm.runningTasks {
		taskCtx.Cancel()
	}
}

// 数据结构定义
type TaskStatusInfo struct {
	TaskID      uint           `json:"task_id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Status      string         `json:"status"`
	Progress    int            `json:"progress"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	StartTime   *time.Time     `json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	Duration    time.Duration  `json:"duration"`
	ErrorMsg    string         `json:"error_msg"`
	IsRunning   bool           `json:"is_running"`
}

type PaginatedTasks struct {
	Data       []models.Task `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}