package services

import (
	"qlib-backend/internal/models"
)

type DashboardService struct{}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// GetOverviewStatistics 获取总览统计数据
func (s *DashboardService) GetOverviewStatistics() (map[string]interface{}, error) {
	db := GetDB()

	// 统计数据集数量
	var totalDatasets, readyDatasets int64
	db.Model(&models.Dataset{}).Count(&totalDatasets)
	db.Model(&models.Dataset{}).Where("status = ?", "active").Count(&readyDatasets)

	// 统计模型数量
	var totalModels, trainedModels int64
	db.Model(&models.Model{}).Count(&totalModels)
	db.Model(&models.Model{}).Where("status = ?", "completed").Count(&trainedModels)

	// 统计任务数量
	var runningTasks, completedTasks int64
	db.Model(&models.Task{}).Where("status = ?", "running").Count(&runningTasks)
	db.Model(&models.Task{}).Where("status = ?", "completed").Count(&completedTasks)

	statistics := map[string]interface{}{
		"total_datasets":  totalDatasets,
		"ready_datasets":  readyDatasets,
		"total_models":    totalModels,
		"trained_models":  trainedModels,
		"running_tasks":   runningTasks,
		"completed_tasks": completedTasks,
	}

	return statistics, nil
}

// GetSystemResources 获取系统资源使用情况
func (s *DashboardService) GetSystemResources() (map[string]interface{}, error) {
	// 模拟系统资源数据，实际应用中应该从系统API获取
	resources := map[string]interface{}{
		"cpu_usage":    65,
		"memory_usage": 78,
		"disk_usage":   45,
		"gpu_usage":    23,
	}

	return resources, nil
}

// GetPerformanceMetrics 获取性能指标
func (s *DashboardService) GetPerformanceMetrics() (map[string]interface{}, error) {
	db := GetDB()

	// 从策略表中计算平均性能
	var avgReturn, avgSharpe, avgDrawdown, avgWinRate float64
	
	db.Model(&models.Strategy{}).
		Where("status = ?", "completed").
		Select("AVG(annual_return)").
		Scan(&avgReturn)

	db.Model(&models.Strategy{}).
		Where("status = ?", "completed").
		Select("AVG(sharpe_ratio)").
		Scan(&avgSharpe)

	db.Model(&models.Strategy{}).
		Where("status = ?", "completed").
		Select("AVG(max_drawdown)").
		Scan(&avgDrawdown)

	db.Model(&models.Strategy{}).
		Where("status = ?", "completed").
		Select("AVG(win_rate)").
		Scan(&avgWinRate)

	performance := map[string]interface{}{
		"total_return": avgReturn,
		"sharpe_ratio": avgSharpe,
		"max_drawdown": avgDrawdown,
		"win_rate":     avgWinRate,
	}

	return performance, nil
}