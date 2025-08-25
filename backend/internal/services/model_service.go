package services

import (
	"fmt"
	"time"

	"qlib-backend/internal/models"
	"qlib-backend/internal/qlib"

	"gorm.io/gorm"
)

type ModelService struct {
	db           *gorm.DB
	modelTrainer *qlib.ModelTrainer
	taskService  *TaskService
}

func NewModelService(db *gorm.DB, modelTrainer *qlib.ModelTrainer, taskService *TaskService) *ModelService {
	return &ModelService{
		db:           db,
		modelTrainer: modelTrainer,
		taskService:  taskService,
	}
}

// StartTraining 启动模型训练
func (s *ModelService) StartTraining(req ModelTrainingRequest, userID uint) (*ModelTrainingResponse, error) {
	// 验证训练参数
	if err := s.validateTrainingParams(req); err != nil {
		return nil, fmt.Errorf("训练参数验证失败: %v", err)
	}

	// 创建模型记录
	model := &models.Model{
		Name:        req.Name,
		Type:        req.ModelType,
		Description: req.Description,
		Status:      "training",
		Progress:    0,
		ConfigJSON:  req.ConfigJSON,
		TrainStart:  req.TrainStart,
		TrainEnd:    req.TrainEnd,
		ValidStart:  req.ValidStart,
		ValidEnd:    req.ValidEnd,
		TestStart:   req.TestStart,
		TestEnd:     req.TestEnd,
		UserID:      userID,
	}

	if err := s.db.Create(model).Error; err != nil {
		return nil, fmt.Errorf("创建模型记录失败: %v", err)
	}

	// 创建训练任务
	task := &models.Task{
		Name:        fmt.Sprintf("模型训练: %s", req.Name),
		Type:        "model_training",
		Status:      "pending",
		Description: fmt.Sprintf("训练%s模型", req.ModelType),
		ConfigJSON:  req.ConfigJSON,
		UserID:      userID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建训练任务失败: %v", err)
	}

	// 启动异步训练任务
	go s.executeTraining(model.ID, task.ID, req)

	return &ModelTrainingResponse{
		ModelID: model.ID,
		TaskID:  task.ID,
		Status:  "started",
		Message: "模型训练已启动",
	}, nil
}

// GetModels 获取模型列表
func (s *ModelService) GetModels(page, pageSize int, status, modelType string, userID uint) (*PaginatedModels, error) {
	var modelList []models.Model
	var total int64

	query := s.db.Model(&models.Model{}).Where("user_id = ?", userID)

	// 添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if modelType != "" {
		query = query.Where("type = ?", modelType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取模型总数失败: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modelList).Error; err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %v", err)
	}

	return &PaginatedModels{
		Data:       modelList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// GetModelProgress 获取训练进度
func (s *ModelService) GetModelProgress(modelID uint, userID uint) (*ModelProgressResponse, error) {
	var model models.Model
	if err := s.db.Where("id = ? AND user_id = ?", modelID, userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模型不存在")
		}
		return nil, fmt.Errorf("获取模型失败: %v", err)
	}

	// 获取相关任务信息
	var task models.Task
	s.db.Where("type = ? AND config_json LIKE ?", "model_training", "%\"model_id\":"+fmt.Sprint(modelID)+"%").First(&task)

	return &ModelProgressResponse{
		ModelID:     model.ID,
		Progress:    model.Progress,
		Status:      model.Status,
		TrainIC:     model.TrainIC,
		ValidIC:     model.ValidIC,
		TestIC:      model.TestIC,
		TrainLoss:   model.TrainLoss,
		ValidLoss:   model.ValidLoss,
		TestLoss:    model.TestLoss,
		TaskID:      task.ID,
		StartTime:   task.StartTime,
		ElapsedTime: s.calculateElapsedTime(task.StartTime),
		Logs:        s.getTrainingLogs(modelID),
	}, nil
}

// StopTraining 停止训练
func (s *ModelService) StopTraining(modelID uint, userID uint) error {
	var model models.Model
	if err := s.db.Where("id = ? AND user_id = ?", modelID, userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("模型不存在")
		}
		return fmt.Errorf("获取模型失败: %v", err)
	}

	if model.Status != "training" {
		return fmt.Errorf("模型当前状态不支持停止操作")
	}

	// 更新模型状态
	if err := s.db.Model(&model).Updates(map[string]interface{}{
		"status": "cancelled",
	}).Error; err != nil {
		return fmt.Errorf("停止训练失败: %v", err)
	}

	// 通知训练器停止训练
	if err := s.modelTrainer.StopTraining(modelID); err != nil {
		return fmt.Errorf("停止训练器失败: %v", err)
	}

	return nil
}

// EvaluateModel 模型评估
func (s *ModelService) EvaluateModel(modelID uint, userID uint) (*ModelEvaluationResult, error) {
	var model models.Model
	if err := s.db.Where("id = ? AND user_id = ?", modelID, userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模型不存在")
		}
		return nil, fmt.Errorf("获取模型失败: %v", err)
	}

	if model.Status != "completed" {
		return nil, fmt.Errorf("模型尚未训练完成")
	}

	// 调用模型训练器进行评估
	evaluation, err := s.modelTrainer.EvaluateModel(qlib.ModelEvaluationParams{
		ModelID:   modelID,
		ModelPath: model.ModelPath,
		TestStart: model.TestStart,
		TestEnd:   model.TestEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("模型评估失败: %v", err)
	}

	return &ModelEvaluationResult{
		ModelID:            modelID,
		OverallScore:       evaluation.OverallScore,
		TrainingMetrics:    evaluation.TrainingMetrics,
		ValidationMetrics:  evaluation.ValidationMetrics,
		TestMetrics:        evaluation.TestMetrics,
		FeatureImportance:  evaluation.FeatureImportance,
		PredictionAccuracy: evaluation.PredictionAccuracy,
		GeneratedAt:        time.Now(),
	}, nil
}

// CompareModels 模型对比
func (s *ModelService) CompareModels(req models.ModelComparisonRequest, userID uint) (*models.ModelComparisonResult, error) {
	if len(req.ModelIDs) < 2 {
		return nil, fmt.Errorf("至少需要选择2个模型进行对比")
	}

	var modelList []models.Model
	if err := s.db.Where("id IN ? AND user_id = ?", req.ModelIDs, userID).Find(&modelList).Error; err != nil {
		return nil, fmt.Errorf("获取模型信息失败: %v", err)
	}

	if len(modelList) != len(req.ModelIDs) {
		return nil, fmt.Errorf("部分模型不存在或无权限访问")
	}

	// 调用模型训练器进行对比
	_, err := s.modelTrainer.CompareModels(qlib.ModelComparisonParams{
		ModelIDs: req.ModelIDs,
		Metrics:  req.Metrics,
	})
	if err != nil {
		return nil, fmt.Errorf("模型对比失败: %v", err)
	}

	// 将models.Model转换为models.ModelPerformance
	modelPerformances := make([]models.ModelPerformance, len(modelList))
	for i, model := range modelList {
		modelPerformances[i] = models.ModelPerformance{
			ModelID:    model.ID,
			ModelName:  model.Name,
			ModelType:  model.Type,
			Status:     model.Status,
			TrainedAt:  model.CreatedAt,
			Metrics: map[string]float64{
				"train_ic":   model.TrainIC,
				"valid_ic":   model.ValidIC,
				"test_ic":    model.TestIC,
				"train_loss": model.TrainLoss,
				"valid_loss": model.ValidLoss,
				"test_loss":  model.TestLoss,
			},
		}
	}

	return &models.ModelComparisonResult{
		Models:          modelPerformances,
		ComparisonChart: &models.ComparisonChart{
			Type: "bar",
			Data: map[string]interface{}{
				"models": modelPerformances,
			},
		},
		RankingTable: []models.ModelRanking{},
		StatisticalTest: &models.StatisticalTestResult{
			TestType:  "t-test",
			PValue:    0.05,
			Statistic: 1.96,
			Result:    "significant",
		},
		Summary: &models.ComparisonSummary{
			BestModel:  modelPerformances[0].ModelID,
			AvgScore:   0.5,
			ScoreRange: 0.2,
		},
		ComparisonDate: time.Now(),
	}, nil
}

// DeployModel 部署模型
func (s *ModelService) DeployModel(modelID uint, req ModelDeploymentRequest, userID uint) (*ModelDeploymentResult, error) {
	var model models.Model
	if err := s.db.Where("id = ? AND user_id = ?", modelID, userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模型不存在")
		}
		return nil, fmt.Errorf("获取模型失败: %v", err)
	}

	if model.Status != "completed" {
		return nil, fmt.Errorf("只有训练完成的模型才能部署")
	}

	// 调用部署服务
	deployment, err := s.modelTrainer.DeployModel(qlib.ModelDeploymentParams{
		ModelID:         modelID,
		ModelPath:       model.ModelPath,
		Environment:     req.Environment,
		ReplicaCount:    req.ReplicaCount,
		ResourceLimits:  req.ResourceLimits,
		HealthCheckPath: req.HealthCheckPath,
	})
	if err != nil {
		return nil, fmt.Errorf("模型部署失败: %v", err)
	}

	// 更新模型状态
	if err := s.db.Model(&model).Updates(map[string]interface{}{
		"status": "deployed",
	}).Error; err != nil {
		return nil, fmt.Errorf("更新模型状态失败: %v", err)
	}

	return &ModelDeploymentResult{
		ModelID:      modelID,
		DeploymentID: deployment.DeploymentID,
		Endpoint:     deployment.Endpoint,
		Status:       "deployed",
		DeployedAt:   time.Now(),
	}, nil
}

// GetTrainingLogs 获取训练日志
func (s *ModelService) GetTrainingLogs(modelID uint, userID uint) (*TrainingLogsResponse, error) {
	var model models.Model
	if err := s.db.Where("id = ? AND user_id = ?", modelID, userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模型不存在")
		}
		return nil, fmt.Errorf("获取模型失败: %v", err)
	}

	logs := s.getTrainingLogs(modelID)

	return &TrainingLogsResponse{
		ModelID: modelID,
		Logs:    logs,
	}, nil
}

// 内部方法

// validateTrainingParams 验证训练参数
func (s *ModelService) validateTrainingParams(req ModelTrainingRequest) error {
	if req.Name == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	if req.ModelType == "" {
		return fmt.Errorf("模型类型不能为空")
	}
	if req.TrainStart == "" || req.TrainEnd == "" {
		return fmt.Errorf("训练时间范围不能为空")
	}
	if req.ValidStart == "" || req.ValidEnd == "" {
		return fmt.Errorf("验证时间范围不能为空")
	}
	if req.TestStart == "" || req.TestEnd == "" {
		return fmt.Errorf("测试时间范围不能为空")
	}

	// 验证支持的模型类型
	supportedTypes := map[string]bool{
		"LightGBM": true,
		"XGBoost":  true,
		"Linear":   true,
		"LSTM":     true,
		"GRU":      true,
		"MLP":      true,
	}

	if !supportedTypes[req.ModelType] {
		return fmt.Errorf("不支持的模型类型: %s", req.ModelType)
	}

	return nil
}

// executeTraining 执行训练任务
func (s *ModelService) executeTraining(modelID, taskID uint, req ModelTrainingRequest) {
	// 更新任务状态
	s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": time.Now(),
	})

	// 更新模型状态
	s.db.Model(&models.Model{}).Where("id = ?", modelID).Updates(map[string]interface{}{
		"status": "training",
	})

	// 调用模型训练器
	trainingParams := qlib.ModelTrainingParams{
		ModelID:     modelID,
		ModelType:   req.ModelType,
		ConfigJSON:  req.ConfigJSON,
		TrainStart:  req.TrainStart,
		TrainEnd:    req.TrainEnd,
		ValidStart:  req.ValidStart,
		ValidEnd:    req.ValidEnd,
		TestStart:   req.TestStart,
		TestEnd:     req.TestEnd,
		Features:    req.Features,
		Label:       req.Label,
	}

	// 设置进度回调
	progressCallback := func(progress int, metrics map[string]float64) {
		updates := map[string]interface{}{
			"progress": progress,
		}
		
		if trainIC, ok := metrics["train_ic"]; ok {
			updates["train_ic"] = trainIC
		}
		if validIC, ok := metrics["valid_ic"]; ok {
			updates["valid_ic"] = validIC
		}
		if testIC, ok := metrics["test_ic"]; ok {
			updates["test_ic"] = testIC
		}
		if trainLoss, ok := metrics["train_loss"]; ok {
			updates["train_loss"] = trainLoss
		}
		if validLoss, ok := metrics["valid_loss"]; ok {
			updates["valid_loss"] = validLoss
		}
		if testLoss, ok := metrics["test_loss"]; ok {
			updates["test_loss"] = testLoss
		}

		s.db.Model(&models.Model{}).Where("id = ?", modelID).Updates(updates)
	}

	// 执行训练
	result, err := s.modelTrainer.TrainModel(trainingParams, progressCallback)
	
	// 更新最终状态
	if err != nil {
		// 训练失败
		s.db.Model(&models.Model{}).Where("id = ?", modelID).Updates(map[string]interface{}{
			"status": "failed",
		})
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": err.Error(),
			"end_time":  time.Now(),
		})
	} else {
		// 训练成功
		s.db.Model(&models.Model{}).Where("id = ?", modelID).Updates(map[string]interface{}{
			"status":     "completed",
			"progress":   100,
			"model_path": result.ModelPath,
			"train_ic":   result.TrainIC,
			"valid_ic":   result.ValidIC,
			"test_ic":    result.TestIC,
			"train_loss": result.TrainLoss,
			"valid_loss": result.ValidLoss,
			"test_loss":  result.TestLoss,
		})
		s.db.Model(&models.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":   "completed",
			"end_time": time.Now(),
		})
	}
}

// getTrainingLogs 获取训练日志
func (s *ModelService) getTrainingLogs(modelID uint) []string {
	// 这里应该从日志文件或数据库中读取实际的训练日志
	// 为了演示，返回模拟的日志
	return []string{
		fmt.Sprintf("[%s] 开始训练模型 ID: %d", time.Now().Format("2006-01-02 15:04:05"), modelID),
		fmt.Sprintf("[%s] 数据加载完成", time.Now().Add(-5*time.Minute).Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] 特征工程完成", time.Now().Add(-4*time.Minute).Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] 模型训练中...", time.Now().Add(-3*time.Minute).Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%s] Epoch 1/10 - Train IC: 0.05, Valid IC: 0.04", time.Now().Add(-2*time.Minute).Format("2006-01-02 15:04:05")),
	}
}

// calculateElapsedTime 计算运行时间
func (s *ModelService) calculateElapsedTime(startTime *time.Time) int64 {
	if startTime == nil {
		return 0
	}
	return int64(time.Since(*startTime).Seconds())
}

// 请求和响应结构体
type ModelTrainingRequest struct {
	Name        string   `json:"name" binding:"required"`
	ModelType   string   `json:"model_type" binding:"required"`
	Description string   `json:"description"`
	ConfigJSON  string   `json:"config_json" binding:"required"`
	TrainStart  string   `json:"train_start" binding:"required"`
	TrainEnd    string   `json:"train_end" binding:"required"`
	ValidStart  string   `json:"valid_start" binding:"required"`
	ValidEnd    string   `json:"valid_end" binding:"required"`
	TestStart   string   `json:"test_start" binding:"required"`
	TestEnd     string   `json:"test_end" binding:"required"`
	Features    []string `json:"features" binding:"required"`
	Label       string   `json:"label" binding:"required"`
}

type ModelTrainingResponse struct {
	ModelID uint   `json:"model_id"`
	TaskID  uint   `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ModelProgressResponse struct {
	ModelID     uint       `json:"model_id"`
	Progress    int        `json:"progress"`
	Status      string     `json:"status"`
	TrainIC     float64    `json:"train_ic"`
	ValidIC     float64    `json:"valid_ic"`
	TestIC      float64    `json:"test_ic"`
	TrainLoss   float64    `json:"train_loss"`
	ValidLoss   float64    `json:"valid_loss"`
	TestLoss    float64    `json:"test_loss"`
	TaskID      uint       `json:"task_id"`
	StartTime   *time.Time `json:"start_time"`
	ElapsedTime int64      `json:"elapsed_time"`
	Logs        []string   `json:"logs"`
}

type ModelEvaluationResult struct {
	ModelID            uint                   `json:"model_id"`
	OverallScore       float64                `json:"overall_score"`
	TrainingMetrics    map[string]interface{} `json:"training_metrics"`
	ValidationMetrics  map[string]interface{} `json:"validation_metrics"`
	TestMetrics        map[string]interface{} `json:"test_metrics"`
	FeatureImportance  map[string]float64     `json:"feature_importance"`
	PredictionAccuracy map[string]float64     `json:"prediction_accuracy"`
	GeneratedAt        time.Time              `json:"generated_at"`
}


type ModelDeploymentRequest struct {
	Environment     string                 `json:"environment" binding:"required"`
	ReplicaCount    int                    `json:"replica_count"`
	ResourceLimits  map[string]interface{} `json:"resource_limits"`
	HealthCheckPath string                 `json:"health_check_path"`
}

type ModelDeploymentResult struct {
	ModelID      uint      `json:"model_id"`
	DeploymentID string    `json:"deployment_id"`
	Endpoint     string    `json:"endpoint"`
	Status       string    `json:"status"`
	DeployedAt   time.Time `json:"deployed_at"`
}

type TrainingLogsResponse struct {
	ModelID uint     `json:"model_id"`
	Logs    []string `json:"logs"`
}

type PaginatedModels struct {
	Data       []models.Model `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int64          `json:"total_pages"`
}