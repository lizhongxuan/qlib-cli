package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"qlib-backend/internal/models"
	"qlib-backend/internal/qlib"
	"qlib-backend/internal/testutils"
)

// MockWorkflowEngine 模拟工作流引擎
type MockWorkflowEngine struct {
	mock.Mock
}

func (m *MockWorkflowEngine) RunWorkflow(ctx context.Context, config WorkflowConfig) (*WorkflowResult, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(*WorkflowResult), args.Error(1)
}

func (m *MockWorkflowEngine) GetWorkflowStatus(taskID uint) (*WorkflowStatus, error) {
	args := m.Called(taskID)
	return args.Get(0).(*WorkflowStatus), args.Error(1)
}

func (m *MockWorkflowEngine) PauseWorkflow(taskID uint) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockWorkflowEngine) ResumeWorkflow(taskID uint) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func (m *MockWorkflowEngine) StopWorkflow(taskID uint) error {
	args := m.Called(taskID)
	return args.Error(0)
}

// MockTaskManager 模拟任务管理器
type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) CreateTask(name, taskType string, userID uint, config interface{}) (uint, error) {
	args := m.Called(name, taskType, userID, config)
	return uint(args.Int(0)), args.Error(1)
}

func (m *MockTaskManager) UpdateTaskStatus(taskID uint, status string, progress int, result interface{}) error {
	args := m.Called(taskID, status, progress, result)
	return args.Error(0)
}

func (m *MockTaskManager) GetTaskStatus(taskID uint) (interface{}, error) {
	args := m.Called(taskID)
	return args.Get(0), args.Error(1)
}

// MockBroadcastService 模拟广播服务
type MockBroadcastService struct {
	mock.Mock
}

func (m *MockBroadcastService) BroadcastWorkflowProgress(taskID uint, progress WorkflowProgressMessage) {
	m.Called(taskID, progress)
}

type WorkflowServiceTestSuite struct {
	suite.Suite
	service          *WorkflowService
	testDB           *testutils.TestDB
	mockEngine       *MockWorkflowEngine
	mockTaskManager  *MockTaskManager
	mockBroadcast    *MockBroadcastService
}

func (suite *WorkflowServiceTestSuite) SetupSuite() {
	suite.testDB = testutils.SetupTestDB()
	suite.mockEngine = new(MockWorkflowEngine)
	suite.mockTaskManager = new(MockTaskManager)
	suite.mockBroadcast = new(MockBroadcastService)
	
	suite.service = &WorkflowService{
		db:               suite.testDB.DB,
		workflowEngine:   suite.mockEngine,
		taskManager:      suite.mockTaskManager,
		broadcastService: suite.mockBroadcast,
		runningWorkflows: make(map[uint]*WorkflowExecution),
	}
}

func (suite *WorkflowServiceTestSuite) TearDownSuite() {
	suite.testDB.Cleanup()
}

func (suite *WorkflowServiceTestSuite) SetupTest() {
	suite.testDB.CleanupTables()
	suite.mockEngine.ExpectedCalls = nil
	suite.mockTaskManager.ExpectedCalls = nil
	suite.mockBroadcast.ExpectedCalls = nil
}

func (suite *WorkflowServiceTestSuite) TestRunWorkflow() {
	userID := uint(1)
	taskID := uint(123)

	req := WorkflowRunRequest{
		TemplateID: 1,
		Name:       "测试工作流",
		Config: map[string]interface{}{
			"dataset_id": 1,
			"model_type": "lightgbm",
			"strategy_type": "top_k",
		},
		Parameters: map[string]interface{}{
			"train_start": "2020-01-01",
			"train_end":   "2022-12-31",
			"test_start":  "2023-01-01",
			"test_end":    "2023-12-31",
		},
	}

	// 设置模拟期望
	suite.mockTaskManager.On("CreateTask", "测试工作流", "workflow", userID, mock.Anything).Return(int(taskID), nil)
	suite.mockEngine.On("RunWorkflow", mock.Anything, mock.Anything).Return(&WorkflowResult{
		TaskID:  taskID,
		Status:  "completed",
		Results: map[string]interface{}{"model_id": 456, "strategy_id": 789},
	}, nil)
	suite.mockBroadcast.On("BroadcastWorkflowProgress", taskID, mock.Anything).Return()

	result, err := suite.service.RunWorkflow(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), taskID, result.TaskID)
	assert.Equal(suite.T(), "submitted", result.Status)
	assert.Equal(suite.T(), req.Name, result.WorkflowName)

	suite.mockTaskManager.AssertExpectations(suite.T())
	suite.mockBroadcast.AssertExpectations(suite.T())
}

func (suite *WorkflowServiceTestSuite) TestGetWorkflowTemplates() {
	category := "quantitative"

	templates, err := suite.service.GetWorkflowTemplates(category)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), templates)
	assert.Greater(suite.T(), len(templates), 0)

	// 验证模板结构
	for _, template := range templates {
		assert.NotEmpty(suite.T(), template.Name)
		assert.NotEmpty(suite.T(), template.Description)
		assert.NotNil(suite.T(), template.Config)
		assert.Greater(suite.T(), len(template.Steps), 0)
	}

	// 如果指定了分类，验证返回的模板都属于该分类
	if category != "" {
		for _, template := range templates {
			assert.Equal(suite.T(), category, template.Category)
		}
	}
}

func (suite *WorkflowServiceTestSuite) TestCreateWorkflowTemplate() {
	userID := uint(1)

	req := WorkflowTemplateRequest{
		Name:        "自定义模板",
		Description: "用户自定义的工作流模板",
		Category:    "custom",
		Config: map[string]interface{}{
			"type": "custom_quantitative",
		},
		Steps: []WorkflowStep{
			{
				Name: "数据加载",
				Type: "data_loading",
				Config: map[string]interface{}{
					"dataset_id": "{{dataset_id}}",
				},
			},
			{
				Name: "模型训练",
				Type: "model_training",
				Config: map[string]interface{}{
					"model_type": "{{model_type}}",
				},
			},
		},
	}

	template, err := suite.service.CreateWorkflowTemplate(req, userID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), template)
	assert.Equal(suite.T(), req.Name, template.Name)
	assert.Equal(suite.T(), req.Description, template.Description)
	assert.Equal(suite.T(), req.Category, template.Category)
	assert.Len(suite.T(), template.Steps, 2)
	assert.Greater(suite.T(), template.ID, uint(0))
}

func (suite *WorkflowServiceTestSuite) TestGetWorkflowStatus() {
	taskID := uint(123)

	suite.mockEngine.On("GetWorkflowStatus", taskID).Return(&WorkflowStatus{
		TaskID:       taskID,
		Status:       "running",
		Progress:     45,
		CurrentStep:  "模型训练中...",
		StartTime:    time.Now().Add(-10 * time.Minute),
		EstimatedEnd: time.Now().Add(15 * time.Minute),
	}, nil)

	status, err := suite.service.GetWorkflowStatus(taskID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), taskID, status.TaskID)
	assert.Equal(suite.T(), "running", status.Status)
	assert.Equal(suite.T(), 45, status.Progress)
	assert.Equal(suite.T(), "模型训练中...", status.CurrentStep)

	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *WorkflowServiceTestSuite) TestPauseWorkflow() {
	taskID := uint(123)

	suite.mockEngine.On("PauseWorkflow", taskID).Return(nil)

	err := suite.service.PauseWorkflow(taskID)

	assert.NoError(suite.T(), err)
	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *WorkflowServiceTestSuite) TestResumeWorkflow() {
	taskID := uint(123)

	suite.mockEngine.On("ResumeWorkflow", taskID).Return(nil)

	err := suite.service.ResumeWorkflow(taskID)

	assert.NoError(suite.T(), err)
	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *WorkflowServiceTestSuite) TestGetWorkflowHistory() {
	userID := uint(1)

	// 创建测试工作流记录
	workflow := models.Workflow{
		Name:        "历史工作流",
		Status:      "completed",
		UserID:      userID,
		TemplateID:  1,
		StartTime:   time.Now().Add(-1 * time.Hour),
		EndTime:     &time.Time{},
		Progress:    100,
		Results:     `{"model_id": 123, "strategy_id": 456}`,
	}
	*workflow.EndTime = time.Now()
	suite.testDB.DB.Create(&workflow)

	page := 1
	pageSize := 10
	status := ""

	history, err := suite.service.GetWorkflowHistory(userID, page, pageSize, status)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), history)
	assert.Equal(suite.T(), int64(1), history.Total)
	assert.Len(suite.T(), history.Workflows, 1)
	assert.Equal(suite.T(), "历史工作流", history.Workflows[0].Name)
	assert.Equal(suite.T(), "completed", history.Workflows[0].Status)
}

func (suite *WorkflowServiceTestSuite) TestValidateWorkflowConfig() {
	// 测试有效配置
	validConfig := WorkflowConfig{
		TemplateID: 1,
		Name:       "有效工作流",
		Steps: []WorkflowStepConfig{
			{
				Name: "数据加载",
				Type: "data_loading",
				Config: map[string]interface{}{
					"dataset_id": 1,
				},
			},
		},
		Parameters: map[string]interface{}{
			"train_start": "2020-01-01",
			"train_end":   "2022-12-31",
		},
	}

	result, err := suite.service.ValidateWorkflowConfig(validConfig)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.IsValid)
	assert.Empty(suite.T(), result.ValidationErrors)
	assert.Greater(suite.T(), result.EstimatedDuration, 0)

	// 测试无效配置
	invalidConfig := WorkflowConfig{
		TemplateID: 0, // 无效的模板ID
		Name:       "", // 空名称
		Steps:      []WorkflowStepConfig{}, // 空步骤
	}

	result, err = suite.service.ValidateWorkflowConfig(invalidConfig)

	assert.NoError(suite.T(), err) // 验证函数本身不应出错
	assert.NotNil(suite.T(), result)
	assert.False(suite.T(), result.IsValid)
	assert.NotEmpty(suite.T(), result.ValidationErrors)
}

func (suite *WorkflowServiceTestSuite) TestStopWorkflow() {
	taskID := uint(123)

	suite.mockEngine.On("StopWorkflow", taskID).Return(nil)

	err := suite.service.StopWorkflow(taskID)

	assert.NoError(suite.T(), err)
	suite.mockEngine.AssertExpectations(suite.T())
}

func (suite *WorkflowServiceTestSuite) TestGetWorkflowResults() {
	taskID := uint(123)

	// 创建测试工作流
	workflow := models.Workflow{
		TaskID:  taskID,
		Name:    "完成的工作流",
		Status:  "completed",
		UserID:  1,
		Results: `{"model_id": 456, "strategy_id": 789, "performance": {"ic": 0.045, "sharpe": 1.23}}`,
	}
	suite.testDB.DB.Create(&workflow)

	results, err := suite.service.GetWorkflowResults(taskID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), results)
	assert.Equal(suite.T(), taskID, results.TaskID)
	assert.Equal(suite.T(), "completed", results.Status)
	assert.NotNil(suite.T(), results.Results)
	assert.Contains(suite.T(), results.Results, "model_id")
	assert.Contains(suite.T(), results.Results, "strategy_id")
	assert.Contains(suite.T(), results.Results, "performance")
}

func TestWorkflowServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowServiceTestSuite))
}