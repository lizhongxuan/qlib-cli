package services

import (
	"encoding/json"
	"fmt"
	"qlib-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

// CreateTask 创建新任务
func (s *TaskService) CreateTask(name, taskType, description string, config map[string]interface{}, userID uint) (*models.Task, error) {
	db := GetDB()

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	task := &models.Task{
		Name:        name,
		Type:        taskType,
		Status:      "pending",
		Progress:    0,
		Description: description,
		ConfigJSON:  string(configJSON),
		UserID:      userID,
	}

	if err := db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetRecentTasks 获取最近任务列表
func (s *TaskService) GetRecentTasks(limit int, userID uint) ([]models.Task, error) {
	db := GetDB()

	var tasks []models.Task
	query := db.Order("created_at DESC").Limit(limit)
	
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTaskProgress 更新任务进度
func (s *TaskService) UpdateTaskProgress(taskID uint, progress int, status string, logMessage string) error {
	db := GetDB()

	updates := map[string]interface{}{
		"progress": progress,
		"status":   status,
	}

	if status == "running" && progress == 0 {
		now := time.Now()
		updates["start_time"] = &now
	} else if status == "completed" || status == "failed" {
		now := time.Now()
		updates["end_time"] = &now
	}

	if err := db.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	return nil
}

// UpdateTaskResult 更新任务结果
func (s *TaskService) UpdateTaskResult(taskID uint, result map[string]interface{}) error {
	db := GetDB()

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := db.Model(&models.Task{}).Where("id = ?", taskID).Update("result_json", string(resultJSON)).Error; err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}

	return nil
}

// GetTaskByID 根据ID获取任务
func (s *TaskService) GetTaskByID(taskID uint) (*models.Task, error) {
	db := GetDB()

	var task models.Task
	if err := db.First(&task, taskID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(taskID uint, userID uint) error {
	db := GetDB()

	// 检查任务是否存在且属于当前用户
	var task models.Task
	if err := db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("task not found or access denied")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 只有pending或running状态的任务才能取消
	if task.Status != "pending" && task.Status != "running" {
		return fmt.Errorf("task cannot be cancelled in current status: %s", task.Status)
	}

	// 更新任务状态
	now := time.Now()
	updates := map[string]interface{}{
		"status":   "cancelled",
		"end_time": &now,
	}

	if err := db.Model(&task).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	return nil
}

// GetTasks 获取任务列表（支持分页和过滤）
func (s *TaskService) GetTasks(page, limit int, taskType, status string, userID uint) ([]models.Task, int64, error) {
	db := GetDB()

	var tasks []models.Task
	var total int64

	query := db.Model(&models.Task{})
	
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get tasks: %w", err)
	}

	return tasks, total, nil
}