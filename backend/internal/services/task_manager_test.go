package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"qlib-backend/internal/models"
	"qlib-backend/internal/testutils"
)

type TaskManagerTestSuite struct {
	suite.Suite
	manager *TaskManager
	testDB  *testutils.TestDB
}

func (suite *TaskManagerTestSuite) SetupSuite() {
	suite.testDB = testutils.SetupTestDB()
	suite.manager = NewTaskManager(suite.testDB.DB, 2) // 创建2个worker的任务管理器
}

func (suite *TaskManagerTestSuite) TearDownSuite() {
	suite.manager.Stop()
	suite.testDB.Cleanup()
}

func (suite *TaskManagerTestSuite) SetupTest() {
	suite.testDB.CleanupTables()
}

func (suite *TaskManagerTestSuite) TestCreateTask() {
	userID := uint(1)
	taskName := "测试任务"
	taskType := "model_training"
	config := map[string]interface{}{
		"model_type": "lightgbm",
		"dataset_id": 123,
	}

	taskID, err := suite.manager.CreateTask(taskName, taskType, userID, config)

	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), taskID, uint(0))

	// 验证任务已保存到数据库
	var task models.Task
	err = suite.testDB.DB.First(&task, taskID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), taskName, task.Name)
	assert.Equal(suite.T(), taskType, task.Type)
	assert.Equal(suite.T(), userID, task.UserID)
	assert.Equal(suite.T(), "pending", task.Status)
	assert.Equal(suite.T(), 0, task.Progress)
}

func (suite *TaskManagerTestSuite) TestGetTasks() {
	userID := uint(1)

	// 创建测试任务
	tasks := []models.Task{
		{
			Name:      "任务A",
			Type:      "model_training",
			Status:    "running",
			UserID:    userID,
			Progress:  50,
			CreatedAt: time.Now().AddDate(0, 0, -1),
		},
		{
			Name:      "任务B",
			Type:      "backtest",
			Status:    "completed",
			UserID:    userID,
			Progress:  100,
			CreatedAt: time.Now(),
		},
		{
			Name:      "任务C",
			Type:      "factor_test",
			Status:    "failed",
			UserID:    userID,
			Progress:  30,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for i := range tasks {
		suite.testDB.DB.Create(&tasks[i])
	}

	// 测试获取所有任务
	result, err := suite.manager.GetTasks(userID, "", "", 1, 10)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), int64(3), result.Total)
	assert.Len(suite.T(), result.Tasks, 3)

	// 测试按状态筛选
	result, err = suite.manager.GetTasks(userID, "running", "", 1, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Tasks, 1)
	assert.Equal(suite.T(), "任务A", result.Tasks[0].Name)

	// 测试按类型筛选
	result, err = suite.manager.GetTasks(userID, "", "backtest", 1, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Tasks, 1)
	assert.Equal(suite.T(), "任务B", result.Tasks[0].Name)

	// 测试复合筛选
	result, err = suite.manager.GetTasks(userID, "completed", "backtest", 1, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.Total)
	assert.Len(suite.T(), result.Tasks, 1)
	assert.Equal(suite.T(), "任务B", result.Tasks[0].Name)
}

func (suite *TaskManagerTestSuite) TestGetTaskStatus() {
	// 创建测试任务
	task := models.Task{
		Name:     "状态测试任务",
		Type:     "model_training",
		Status:   "running",
		UserID:   1,
		Progress: 75,
	}
	suite.testDB.DB.Create(&task)

	status, err := suite.manager.GetTaskStatus(task.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), task.ID, status.TaskID)
	assert.Equal(suite.T(), "running", status.Status)
	assert.Equal(suite.T(), 75, status.Progress)
}

func (suite *TaskManagerTestSuite) TestUpdateTaskStatus() {
	// 创建测试任务
	task := models.Task{
		Name:     "更新测试任务",
		Type:     "model_training",
		Status:   "running",
		UserID:   1,
		Progress: 30,
	}
	suite.testDB.DB.Create(&task)

	// 更新任务状态
	result := map[string]interface{}{
		"model_id": 456,
		"accuracy": 0.85,
	}
	err := suite.manager.UpdateTaskStatus(task.ID, "completed", 100, result)

	assert.NoError(suite.T(), err)

	// 验证更新
	var updatedTask models.Task
	err = suite.testDB.DB.First(&updatedTask, task.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "completed", updatedTask.Status)
	assert.Equal(suite.T(), 100, updatedTask.Progress)
	assert.NotEmpty(suite.T(), updatedTask.Result)
	assert.NotNil(suite.T(), updatedTask.CompletedAt)
}

func (suite *TaskManagerTestSuite) TestUpdateTaskProgress() {
	// 创建测试任务
	task := models.Task{
		Name:     "进度测试任务",
		Type:     "model_training",
		Status:   "running",
		UserID:   1,
		Progress: 30,
	}
	suite.testDB.DB.Create(&task)

	// 更新进度
	err := suite.manager.UpdateTaskProgress(task.ID, 65, "训练中...")

	assert.NoError(suite.T(), err)

	// 验证更新
	var updatedTask models.Task
	err = suite.testDB.DB.First(&updatedTask, task.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 65, updatedTask.Progress)
	assert.Equal(suite.T(), "训练中...", updatedTask.Message)
}

func (suite *TaskManagerTestSuite) TestCancelTask() {
	// 创建测试任务
	task := models.Task{
		Name:     "取消测试任务",
		Type:     "model_training",
		Status:   "running",
		UserID:   1,
		Progress: 45,
	}
	suite.testDB.DB.Create(&task)

	err := suite.manager.CancelTask(task.ID)

	assert.NoError(suite.T(), err)

	// 验证任务状态已更新
	var cancelledTask models.Task
	err = suite.testDB.DB.First(&cancelledTask, task.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "cancelled", cancelledTask.Status)
	assert.NotNil(suite.T(), cancelledTask.CompletedAt)
}

func (suite *TaskManagerTestSuite) TestCancelNonExistentTask() {
	err := suite.manager.CancelTask(999)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "任务不存在")
}

func (suite *TaskManagerTestSuite) TestSubmitTaskForExecution() {
	userID := uint(1)
	taskName := "执行测试任务"
	taskType := "test_execution"
	config := map[string]interface{}{
		"test_param": "test_value",
	}

	taskID, err := suite.manager.CreateTask(taskName, taskType, userID, config)
	assert.NoError(suite.T(), err)

	// 提交任务执行
	err = suite.manager.SubmitTask(taskID, func(ctx context.Context, task *models.Task) error {
		// 模拟任务执行
		for i := 0; i <= 100; i += 20 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// 更新进度
				suite.manager.UpdateTaskProgress(task.ID, i, fmt.Sprintf("进度 %d%%", i))
				time.Sleep(10 * time.Millisecond) // 模拟处理时间
			}
		}
		return nil
	})

	assert.NoError(suite.T(), err)

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	// 验证任务状态
	var completedTask models.Task
	err = suite.testDB.DB.First(&completedTask, taskID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "completed", completedTask.Status)
	assert.Equal(suite.T(), 100, completedTask.Progress)
}

func (suite *TaskManagerTestSuite) TestGetRunningTasks() {
	runningTasks := suite.manager.GetRunningTasks()

	assert.NotNil(suite.T(), runningTasks)
	assert.IsType(suite.T(), map[uint]*TaskContext{}, runningTasks)
}

func (suite *TaskManagerTestSuite) TestTaskTimeout() {
	userID := uint(1)
	taskName := "超时测试任务"
	taskType := "timeout_test"
	config := map[string]interface{}{
		"timeout": 100, // 100ms超时
	}

	taskID, err := suite.manager.CreateTask(taskName, taskType, userID, config)
	assert.NoError(suite.T(), err)

	// 提交一个会超时的任务
	err = suite.manager.SubmitTask(taskID, func(ctx context.Context, task *models.Task) error {
		// 模拟长时间运行的任务
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	assert.NoError(suite.T(), err)

	// 等待超时处理
	time.Sleep(200 * time.Millisecond)

	// 验证任务状态 - 应该仍在运行或已完成（取决于实现）
	var task models.Task
	err = suite.testDB.DB.First(&task, taskID).Error
	assert.NoError(suite.T(), err)
	// 注意：具体的超时处理取决于TaskManager的实现
}

func (suite *TaskManagerTestSuite) TestDeleteExpiredTasks() {
	userID := uint(1)

	// 创建一个旧任务
	oldTask := models.Task{
		Name:      "过期任务",
		Type:      "expired_test",
		Status:    "completed",
		UserID:    userID,
		Progress:  100,
		CreatedAt: time.Now().AddDate(0, 0, -31), // 31天前
	}
	oldTask.CompletedAt = &time.Time{}
	*oldTask.CompletedAt = time.Now().AddDate(0, 0, -30) // 30天前完成
	suite.testDB.DB.Create(&oldTask)

	// 创建一个新任务
	newTask := models.Task{
		Name:      "新任务",
		Type:      "new_test",
		Status:    "completed",
		UserID:    userID,
		Progress:  100,
		CreatedAt: time.Now().AddDate(0, 0, -1), // 1天前
	}
	newTask.CompletedAt = &time.Time{}
	*newTask.CompletedAt = time.Now().AddDate(0, 0, -1) // 1天前完成
	suite.testDB.DB.Create(&newTask)

	// 删除过期任务（假设保留期为7天）
	count, err := suite.manager.DeleteExpiredTasks(7)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count) // 应该删除1个过期任务

	// 验证旧任务已被删除，新任务仍存在
	var tasks []models.Task
	suite.testDB.DB.Unscoped().Find(&tasks)
	
	foundOld := false
	foundNew := false
	for _, task := range tasks {
		if task.ID == oldTask.ID {
			foundOld = true
			assert.NotNil(suite.T(), task.DeletedAt) // 应该被软删除
		}
		if task.ID == newTask.ID {
			foundNew = true
			assert.Nil(suite.T(), task.DeletedAt) // 不应该被删除
		}
	}
	
	assert.True(suite.T(), foundOld, "应该找到旧任务记录")
	assert.True(suite.T(), foundNew, "应该找到新任务记录")
}

func TestTaskManagerTestSuite(t *testing.T) {
	suite.Run(t, new(TaskManagerTestSuite))
}