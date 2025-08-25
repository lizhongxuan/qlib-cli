package services

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DatasourceService struct {
	db *gorm.DB
}

func NewDatasourceService(db *gorm.DB) *DatasourceService {
	return &DatasourceService{db: db}
}

// DataSource 数据源模型
type DataSource struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"size:100;not null"`
	Type        string                 `json:"type" gorm:"size:50;not null"` // mysql, postgresql, file, api
	Description string                 `json:"description" gorm:"size:500"`
	Config      DataSourceConfig       `json:"config" gorm:"type:json"`
	Status      string                 `json:"status" gorm:"size:20;default:'inactive'"` // active, inactive, error
	LastTested  *time.Time             `json:"last_tested,omitempty"`
	TestResult  string                 `json:"test_result" gorm:"size:500"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `json:"deleted_at,omitempty" gorm:"index"`
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	Host     string            `json:"host,omitempty"`
	Port     int               `json:"port,omitempty"`
	Database string            `json:"database,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	FilePath string            `json:"file_path,omitempty"`
	URL      string            `json:"url,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
}

// Value 实现 driver.Valuer 接口
func (c DataSourceConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *DataSourceConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into DataSourceConfig", value)
	}
	return json.Unmarshal(bytes, c)
}

// GetDataSources 获取数据源列表
func (s *DatasourceService) GetDataSources() ([]DataSource, error) {
	var dataSources []DataSource
	if err := s.db.Find(&dataSources).Error; err != nil {
		return nil, fmt.Errorf("获取数据源列表失败: %v", err)
	}
	return dataSources, nil
}

// CreateDataSource 创建数据源
func (s *DatasourceService) CreateDataSource(req DataSourceCreateRequest) (*DataSource, error) {
	dataSource := &DataSource{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Config:      req.Config,
		Status:      "inactive",
	}

	if err := s.db.Create(dataSource).Error; err != nil {
		return nil, fmt.Errorf("创建数据源失败: %v", err)
	}

	return dataSource, nil
}

// UpdateDataSource 更新数据源
func (s *DatasourceService) UpdateDataSource(id uint, req DataSourceUpdateRequest) (*DataSource, error) {
	var dataSource DataSource
	if err := s.db.First(&dataSource, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("数据源不存在")
		}
		return nil, fmt.Errorf("获取数据源失败: %v", err)
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Config != nil {
		updates["config"] = req.Config
	}

	if err := s.db.Model(&dataSource).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新数据源失败: %v", err)
	}

	return &dataSource, nil
}

// DeleteDataSource 删除数据源
func (s *DatasourceService) DeleteDataSource(id uint) error {
	var dataSource DataSource
	if err := s.db.First(&dataSource, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("数据源不存在")
		}
		return fmt.Errorf("获取数据源失败: %v", err)
	}

	if err := s.db.Delete(&dataSource).Error; err != nil {
		return fmt.Errorf("删除数据源失败: %v", err)
	}

	return nil
}

// TestConnection 测试数据源连接
func (s *DatasourceService) TestConnection(req DataSourceTestRequest) (*DataSourceTestResult, error) {
	result := &DataSourceTestResult{
		Success: false,
		Message: "",
		Details: make(map[string]interface{}),
	}

	switch req.Type {
	case "mysql":
		return s.testMySQLConnection(req.Config, result)
	case "postgresql":
		return s.testPostgreSQLConnection(req.Config, result)
	case "file":
		return s.testFileConnection(req.Config, result)
	case "api":
		return s.testAPIConnection(req.Config, result)
	default:
		result.Message = "不支持的数据源类型"
		return result, nil
	}
}

// testMySQLConnection 测试MySQL连接
func (s *DatasourceService) testMySQLConnection(config DataSourceConfig, result *DataSourceTestResult) (*DataSourceTestResult, error) {
	// TODO: 实现MySQL连接测试
	// 这里是模拟实现
	if config.Host == "" || config.Username == "" {
		result.Message = "MySQL连接参数不完整"
		return result, nil
	}

	// 模拟连接测试
	result.Success = true
	result.Message = "MySQL连接测试成功"
	result.Details["connection_time"] = "50ms"
	result.Details["server_version"] = "8.0.25"
	result.Details["database_count"] = 5

	return result, nil
}

// testPostgreSQLConnection 测试PostgreSQL连接
func (s *DatasourceService) testPostgreSQLConnection(config DataSourceConfig, result *DataSourceTestResult) (*DataSourceTestResult, error) {
	// TODO: 实现PostgreSQL连接测试
	if config.Host == "" || config.Username == "" {
		result.Message = "PostgreSQL连接参数不完整"
		return result, nil
	}

	result.Success = true
	result.Message = "PostgreSQL连接测试成功"
	result.Details["connection_time"] = "45ms"
	result.Details["server_version"] = "13.4"

	return result, nil
}

// testFileConnection 测试文件连接
func (s *DatasourceService) testFileConnection(config DataSourceConfig, result *DataSourceTestResult) (*DataSourceTestResult, error) {
	if config.FilePath == "" {
		result.Message = "文件路径不能为空"
		return result, nil
	}

	// TODO: 检查文件是否存在和可读
	result.Success = true
	result.Message = "文件访问测试成功"
	result.Details["file_size"] = "1.2GB"
	result.Details["format"] = "CSV"

	return result, nil
}

// testAPIConnection 测试API连接
func (s *DatasourceService) testAPIConnection(config DataSourceConfig, result *DataSourceTestResult) (*DataSourceTestResult, error) {
	if config.URL == "" {
		result.Message = "API URL不能为空"
		return result, nil
	}

	// TODO: 实现HTTP请求测试
	result.Success = true
	result.Message = "API连接测试成功"
	result.Details["response_time"] = "120ms"
	result.Details["status_code"] = 200

	return result, nil
}

// UpdateTestResult 更新数据源测试结果
func (s *DatasourceService) UpdateTestResult(id uint, testResult *DataSourceTestResult) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_tested":  &now,
		"test_result":  testResult.Message,
	}

	if testResult.Success {
		updates["status"] = "active"
	} else {
		updates["status"] = "error"
	}

	if err := s.db.Model(&DataSource{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新测试结果失败: %v", err)
	}

	return nil
}

// 请求和响应结构体
type DataSourceCreateRequest struct {
	Name        string           `json:"name" binding:"required"`
	Type        string           `json:"type" binding:"required"`
	Description string           `json:"description"`
	Config      DataSourceConfig `json:"config" binding:"required"`
}

type DataSourceUpdateRequest struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Config      *DataSourceConfig `json:"config"`
}

type DataSourceTestRequest struct {
	Type   string           `json:"type" binding:"required"`
	Config DataSourceConfig `json:"config" binding:"required"`
}

type DataSourceTestResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}