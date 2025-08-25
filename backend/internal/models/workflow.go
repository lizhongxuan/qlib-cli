package models

import (
	"time"
	"gorm.io/gorm"
)

// Workflow 工作流模型
type Workflow struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	TemplateID   uint           `json:"template_id" gorm:"not null"`
	Status       string         `json:"status" gorm:"size:50;not null;default:'queued'"` // queued, running, paused, completed, failed, cancelled
	Progress     int            `json:"progress" gorm:"default:0"`
	ConfigJSON   string         `json:"config_json" gorm:"type:text"`
	ResultJSON   string         `json:"result_json" gorm:"type:text"`
	StartTime    *time.Time     `json:"start_time"`
	EndTime      *time.Time     `json:"end_time"`
	ErrorMsg     string         `json:"error_msg" gorm:"type:text"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// 关联关系
	Template *WorkflowTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	User     *User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Tasks    []Task            `json:"tasks,omitempty" gorm:"foreignKey:WorkflowID"`
}

func (Workflow) TableName() string {
	return "workflows"
}

// WorkflowTemplate 工作流模板模型
type WorkflowTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    string         `json:"category" gorm:"size:100"`
	ConfigJSON  string         `json:"config_json" gorm:"type:text"`
	StepsJSON   string         `json:"steps_json" gorm:"type:text"`
	IsBuiltin   bool           `json:"is_builtin" gorm:"default:false"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedBy   uint           `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// 关联关系
	Creator   *User      `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Workflows []Workflow `json:"workflows,omitempty" gorm:"foreignKey:TemplateID"`
}

func (WorkflowTemplate) TableName() string {
	return "workflow_templates"
}

// WorkflowStep 工作流步骤模型（嵌入式，不单独存储）
type WorkflowStep struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
	Required     bool                   `json:"required"`
	Order        int                    `json:"order"`
}

// WorkflowExecution 工作流执行记录模型
type WorkflowExecution struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	WorkflowID   uint           `json:"workflow_id" gorm:"not null"`
	TaskID       uint           `json:"task_id"`
	Status       string         `json:"status" gorm:"size:50;not null"` // queued, running, paused, completed, failed, cancelled
	CurrentStep  string         `json:"current_step" gorm:"size:255"`
	Progress     int            `json:"progress" gorm:"default:0"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time"`
	Duration     *int64         `json:"duration"` // milliseconds
	ResultJSON   string         `json:"result_json" gorm:"type:text"`
	ErrorMsg     string         `json:"error_msg" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// 关联关系
	Workflow *Workflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
	Task     *Task     `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

// WorkflowStepExecution 工作流步骤执行记录模型
type WorkflowStepExecution struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ExecutionID  uint           `json:"execution_id" gorm:"not null"`
	StepName     string         `json:"step_name" gorm:"size:255;not null"`
	StepType     string         `json:"step_type" gorm:"size:100;not null"`
	Status       string         `json:"status" gorm:"size:50;not null"` // queued, running, completed, failed, skipped
	StartTime    *time.Time     `json:"start_time"`
	EndTime      *time.Time     `json:"end_time"`
	Duration     *int64         `json:"duration"` // milliseconds
	InputJSON    string         `json:"input_json" gorm:"type:text"`
	OutputJSON   string         `json:"output_json" gorm:"type:text"`
	ErrorMsg     string         `json:"error_msg" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	
	// 关联关系
	Execution *WorkflowExecution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID"`
}

func (WorkflowStepExecution) TableName() string {
	return "workflow_step_executions"
}

// BeforeCreate 创建前钩子
func (w *Workflow) BeforeCreate(tx *gorm.DB) error {
	if w.Status == "" {
		w.Status = "queued"
	}
	return nil
}

// BeforeCreate 创建前钩子
func (wt *WorkflowTemplate) BeforeCreate(tx *gorm.DB) error {
	if wt.Category == "" {
		wt.Category = "custom"
	}
	return nil
}

// IsRunning 检查工作流是否正在运行
func (w *Workflow) IsRunning() bool {
	return w.Status == "running" || w.Status == "paused"
}

// IsCompleted 检查工作流是否已完成
func (w *Workflow) IsCompleted() bool {
	return w.Status == "completed" || w.Status == "failed" || w.Status == "cancelled"
}

// GetDuration 获取执行时长
func (w *Workflow) GetDuration() *time.Duration {
	if w.StartTime == nil {
		return nil
	}
	
	endTime := w.EndTime
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}
	
	duration := endTime.Sub(*w.StartTime)
	return &duration
}

// WorkflowCategory 工作流分类常量
const (
	WorkflowCategoryStrategy = "strategy"
	WorkflowCategoryResearch = "research"
	WorkflowCategoryAnalysis = "analysis"
	WorkflowCategoryCustom   = "custom"
)

// WorkflowStatus 工作流状态常量
const (
	WorkflowStatusQueued    = "queued"
	WorkflowStatusRunning   = "running"
	WorkflowStatusPaused    = "paused"
	WorkflowStatusCompleted = "completed"
	WorkflowStatusFailed    = "failed"
	WorkflowStatusCancelled = "cancelled"
)

// WorkflowStepType 工作流步骤类型常量
const (
	StepTypeDataPreparation    = "data_preparation"
	StepTypeFactorGeneration   = "factor_generation"
	StepTypeModelTraining      = "model_training"
	StepTypeStrategyBacktest   = "strategy_backtest"
	StepTypeResultAnalysis     = "result_analysis"
	StepTypeReportGeneration   = "report_generation"
	StepTypeCustom             = "custom"
)