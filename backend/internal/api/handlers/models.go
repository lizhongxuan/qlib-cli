package handlers

import (
	"qlib-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// StartModelTraining 启动模型训练
func StartModelTraining(c *gin.Context) {
	var req struct {
		Name       string                 `json:"name" binding:"required"`
		Type       string                 `json:"type" binding:"required"`
		Config     map[string]interface{} `json:"config" binding:"required"`
		DatasetID  int                    `json:"dataset_id" binding:"required"`
		FactorIDs  []int                  `json:"factor_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	// 模拟启动模型训练
	result := gin.H{
		"task_id":        "model_train_123",
		"model_id":       123,
		"name":           req.Name,
		"type":           req.Type,
		"status":         "pending",
		"progress":       0,
		"estimated_time": 1800, // 30分钟
		"start_time":     "2024-01-15T10:00:00Z",
	}

	utils.SuccessWithMessage(c, "模型训练已启动", result)
}

// GetModels 获取模型列表
func GetModels(c *gin.Context) {
	models := []gin.H{
		{
			"id":          1,
			"name":        "LightGBM-Alpha158",
			"type":        "LightGBM",
			"status":      "completed",
			"progress":    100,
			"train_ic":    0.0456,
			"valid_ic":    0.0398,
			"test_ic":     0.0367,
			"created_at":  "2024-01-10T08:00:00Z",
			"updated_at":  "2024-01-10T10:30:00Z",
		},
		{
			"id":          2,
			"name":        "XGBoost-Alpha360",
			"type":        "XGBoost", 
			"status":      "training",
			"progress":    65,
			"train_ic":    0.0398,
			"valid_ic":    0.0345,
			"created_at":  "2024-01-15T09:00:00Z",
			"updated_at":  "2024-01-15T10:00:00Z",
		},
	}

	utils.SuccessResponse(c, gin.H{"models": models})
}

// GetTrainingProgress 获取训练进度
func GetTrainingProgress(c *gin.Context) {
	id := c.Param("id")

	progress := gin.H{
		"model_id":       id,
		"status":         "training",
		"progress":       65,
		"current_epoch":  65,
		"total_epochs":   100,
		"current_loss":   0.0234,
		"best_valid_ic":  0.0398,
		"estimated_time": 600, // 剩余10分钟
		"logs": []string{
			"Epoch 60: train_loss=0.0245, valid_ic=0.0387",
			"Epoch 61: train_loss=0.0241, valid_ic=0.0392",
			"Epoch 62: train_loss=0.0238, valid_ic=0.0395",
		},
	}

	utils.SuccessResponse(c, progress)
}

// StopTraining 停止训练
func StopTraining(c *gin.Context) {
	id := c.Param("id")

	result := gin.H{
		"model_id": id,
		"status":   "stopped",
		"message":  "训练已停止",
	}

	utils.SuccessWithMessage(c, "模型训练已停止", result)
}

// EvaluateModel 模型评估
func EvaluateModel(c *gin.Context) {
	id := c.Param("id")

	evaluation := gin.H{
		"model_id": id,
		"metrics": gin.H{
			"train_ic":    0.0456,
			"valid_ic":    0.0398,
			"test_ic":     0.0367,
			"train_rank_ic": 0.0612,
			"valid_rank_ic": 0.0534,
			"test_rank_ic":  0.0489,
			"mse":         0.0023,
			"mae":         0.0187,
		},
		"feature_importance": []gin.H{
			{"feature": "RESI5", "importance": 0.125},
			{"feature": "WVMA5", "importance": 0.098},
			{"feature": "RSQR10", "importance": 0.087},
		},
	}

	utils.SuccessResponse(c, evaluation)
}

// CompareModels 模型对比
func CompareModels(c *gin.Context) {
	var req struct {
		ModelIDs []int `json:"model_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	comparison := gin.H{
		"models": []gin.H{
			{
				"id":      1,
				"name":    "LightGBM-Alpha158",
				"test_ic": 0.0367,
				"test_rank_ic": 0.0489,
			},
			{
				"id":      2,
				"name":    "XGBoost-Alpha360", 
				"test_ic": 0.0312,
				"test_rank_ic": 0.0434,
			},
		},
		"best_model": gin.H{
			"id":     1,
			"metric": "test_ic",
			"value":  0.0367,
		},
	}

	utils.SuccessResponse(c, comparison)
}

// DeployModel 部署模型
func DeployModel(c *gin.Context) {
	id := c.Param("id")

	result := gin.H{
		"model_id":     id,
		"deployment_id": "deploy_123",
		"status":       "deploying",
		"endpoint":     "/api/v1/models/predict/" + id,
	}

	utils.SuccessWithMessage(c, "模型部署已启动", result)
}

// GetTrainingLogs 获取训练日志
func GetTrainingLogs(c *gin.Context) {
	id := c.Param("id")

	logs := gin.H{
		"model_id": id,
		"logs": []gin.H{
			{
				"timestamp": "2024-01-15T10:00:00Z",
				"level":     "INFO",
				"message":   "开始模型训练",
			},
			{
				"timestamp": "2024-01-15T10:01:00Z",
				"level":     "INFO",
				"message":   "加载训练数据完成，共100万条记录",
			},
			{
				"timestamp": "2024-01-15T10:05:00Z",
				"level":     "INFO",
				"message":   "Epoch 1/100 完成，train_loss: 0.0567",
			},
		},
	}

	utils.SuccessResponse(c, logs)
}