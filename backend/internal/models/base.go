package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Dataset 数据集模型
type Dataset struct {
	BaseModel
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	DataPath    string `json:"data_path" gorm:"size:255;not null"`
	Status      string `json:"status" gorm:"size:20;default:'active'"` // active, inactive, error
	Market      string `json:"market" gorm:"size:50"`                  // csi300, csi500, etc
	StartDate   string `json:"start_date" gorm:"size:10"`
	EndDate     string `json:"end_date" gorm:"size:10"`
	FileSize    int64  `json:"file_size"`
	RecordCount int64  `json:"record_count"`
	UserID      uint   `json:"user_id,omitempty"`                      // 创建者ID
}

// Factor 因子模型
type Factor struct {
	BaseModel
	Name        string  `json:"name" gorm:"size:100;not null"`
	Expression  string  `json:"expression" gorm:"type:text;not null"`
	Description string  `json:"description" gorm:"size:500"`
	Category    string  `json:"category" gorm:"size:50"`     // price, volume, momentum, etc
	Status      string  `json:"status" gorm:"size:20"`       // active, testing, disabled
	IC          float64 `json:"ic"`                          // Information Coefficient
	IR          float64 `json:"ir"`                          // Information Ratio
	RankIC      float64 `json:"rank_ic"`                     // Rank Information Coefficient
	Turnover    float64 `json:"turnover"`                    // 换手率
	Coverage    float64 `json:"coverage"`                    // 覆盖率
	UserID      uint    `json:"user_id,omitempty"`           // 创建者ID
	IsPublic    bool    `json:"is_public" gorm:"default:0"`  // 是否公开
}

// Model 模型实体
type Model struct {
	BaseModel
	Name         string  `json:"name" gorm:"size:100;not null"`
	Type         string  `json:"type" gorm:"size:50;not null"`    // LightGBM, XGBoost, Linear, etc
	Description  string  `json:"description" gorm:"size:500"`
	Status       string  `json:"status" gorm:"size:20"`           // training, completed, failed, deployed
	Progress     int     `json:"progress" gorm:"default:0"`       // 训练进度 0-100
	ModelPath    string  `json:"model_path" gorm:"size:255"`      // 模型文件路径
	ConfigJSON   string  `json:"config_json" gorm:"type:text"`    // 模型配置JSON
	TrainIC      float64 `json:"train_ic"`                        // 训练集IC
	ValidIC      float64 `json:"valid_ic"`                        // 验证集IC
	TestIC       float64 `json:"test_ic"`                         // 测试集IC
	TrainLoss    float64 `json:"train_loss"`                      // 训练损失
	ValidLoss    float64 `json:"valid_loss"`                      // 验证损失
	TestLoss     float64 `json:"test_loss"`                       // 测试损失
	TrainStart   string  `json:"train_start" gorm:"size:10"`      // 训练开始日期
	TrainEnd     string  `json:"train_end" gorm:"size:10"`        // 训练结束日期
	ValidStart   string  `json:"valid_start" gorm:"size:10"`      // 验证开始日期
	ValidEnd     string  `json:"valid_end" gorm:"size:10"`        // 验证结束日期
	TestStart    string  `json:"test_start" gorm:"size:10"`       // 测试开始日期
	TestEnd      string  `json:"test_end" gorm:"size:10"`         // 测试结束日期
	UserID       uint    `json:"user_id,omitempty"`               // 创建者ID
}

// Strategy 策略模型
type Strategy struct {
	BaseModel
	Name           string  `json:"name" gorm:"size:100;not null"`
	Type           string  `json:"type" gorm:"size:50;not null"`         // TopkDropoutStrategy, etc
	Description    string  `json:"description" gorm:"size:500"`
	Status         string  `json:"status" gorm:"size:20"`                // backtesting, completed, failed
	Progress       int     `json:"progress" gorm:"default:0"`            // 回测进度 0-100
	ConfigJSON     string  `json:"config_json" gorm:"type:text"`         // 策略配置JSON
	ModelID        uint    `json:"model_id"`                             // 关联模型ID
	BacktestStart  string  `json:"backtest_start" gorm:"size:10"`        // 回测开始日期
	BacktestEnd    string  `json:"backtest_end" gorm:"size:10"`          // 回测结束日期
	TotalReturn    float64 `json:"total_return"`                         // 总收益率
	AnnualReturn   float64 `json:"annual_return"`                        // 年化收益率
	BenchmarkReturn float64 `json:"benchmark_return"`                    // 基准收益率
	ExcessReturn   float64 `json:"excess_return"`                        // 超额收益率
	SharpeRatio    float64 `json:"sharpe_ratio"`                         // 夏普比率
	MaxDrawdown    float64 `json:"max_drawdown"`                         // 最大回撤
	Volatility     float64 `json:"volatility"`                           // 波动率
	WinRate        float64 `json:"win_rate"`                             // 胜率
	UserID         uint    `json:"user_id,omitempty"`                    // 创建者ID
	Model          Model   `json:"model,omitempty" gorm:"foreignKey:ModelID"`
}

// Task 任务模型
type Task struct {
	BaseModel
	Name        string `json:"name" gorm:"size:100;not null"`
	Type        string `json:"type" gorm:"size:50;not null"`      // model_training, strategy_backtest, factor_test, workflow_execution
	Status      string `json:"status" gorm:"size:20;default:'queued'"` // queued, running, paused, completed, failed, cancelled
	Progress    int    `json:"progress" gorm:"default:0"`         // 进度 0-100
	Priority    int    `json:"priority" gorm:"default:1"`         // 任务优先级
	Description string `json:"description" gorm:"size:500"`
	ConfigJSON  string `json:"config_json" gorm:"type:text"`      // 任务配置JSON
	ResultJSON  string `json:"result_json" gorm:"type:text"`      // 结果JSON
	LogPath     string `json:"log_path" gorm:"size:255"`          // 日志文件路径
	ErrorMsg    string `json:"error_msg" gorm:"type:text"`        // 错误信息
	StartTime   *time.Time `json:"start_time,omitempty"`           // 开始时间
	EndTime     *time.Time `json:"end_time,omitempty"`             // 结束时间
	EstimatedTime int      `json:"estimated_time"`                 // 预估耗时（秒）
	UserID      uint   `json:"user_id,omitempty"`                 // 创建者ID
	WorkflowID  *uint  `json:"workflow_id,omitempty"`             // 关联工作流ID
}

// User 用户模型
type User struct {
	BaseModel
	Username  string `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Email     string `json:"email" gorm:"size:100;not null;uniqueIndex"`
	Password  string `json:"-" gorm:"size:255;not null"` // 不在JSON中返回密码
	FirstName string `json:"first_name" gorm:"size:50"`
	LastName  string `json:"last_name" gorm:"size:50"`
	RealName  string `json:"real_name" gorm:"size:50"`
	Avatar    string `json:"avatar" gorm:"size:500"`
	Phone     string `json:"phone" gorm:"size:20"`
	Status    string `json:"status" gorm:"size:20;default:'active'"` // active, inactive, banned
	Role      string `json:"role" gorm:"size:20;default:'user'"`     // admin, user
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// Notification 通知模型
type Notification struct {
	BaseModel
	Type      string     `json:"type" gorm:"size:50;not null"`       // success, info, warning, error
	Category  string     `json:"category" gorm:"size:50"`            // task_completion, system_alert, etc
	Title     string     `json:"title" gorm:"size:200;not null"`
	Message   string     `json:"message" gorm:"type:text;not null"`
	Data      string     `json:"data" gorm:"type:text"`              // JSON字符串存储额外数据
	IsRead    bool       `json:"is_read" gorm:"default:0"`
	ReadAt    *time.Time `json:"read_at,omitempty"`                  // 阅读时间
	ActionURL string     `json:"action_url" gorm:"size:255"`         // 操作链接
	Priority  string     `json:"priority" gorm:"size:20;default:'normal'"` // high, normal, low
	ExpiresAt *time.Time `json:"expires_at,omitempty"`              // 过期时间
	UserID    uint       `json:"user_id"`                            // 接收用户ID
	TaskID    *uint      `json:"task_id,omitempty"`                  // 关联任务ID
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Task      *Task      `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// UIConfig 界面配置模型
type UIConfig struct {
	BaseModel
	UserID     uint   `json:"user_id" gorm:"not null"`
	ConfigType string `json:"config_type" gorm:"size:50;not null"` // default, custom, analysis, research
	Platform   string `json:"platform" gorm:"size:20;not null"`    // web, mobile, tablet
	ConfigData string `json:"config_data" gorm:"type:text"`        // JSON格式的配置数据
}

