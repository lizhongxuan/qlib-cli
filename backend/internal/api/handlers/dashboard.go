package handlers

import (
	"strconv"
	"qlib-backend/internal/services"
	"qlib-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetDashboardOverview 获取总览统计数据
func GetDashboardOverview(c *gin.Context) {
	dashboardService := services.NewDashboardService()

	// 获取统计数据
	statistics, err := dashboardService.GetOverviewStatistics()
	if err != nil {
		utils.InternalErrorResponse(c, "获取统计数据失败: "+err.Error())
		return
	}

	// 获取性能指标
	performance, err := dashboardService.GetPerformanceMetrics()
	if err != nil {
		utils.InternalErrorResponse(c, "获取性能指标失败: "+err.Error())
		return
	}

	// 获取系统资源
	systemResources, err := dashboardService.GetSystemResources()
	if err != nil {
		utils.InternalErrorResponse(c, "获取系统资源失败: "+err.Error())
		return
	}

	data := gin.H{
		"statistics":       statistics,
		"performance":      performance,
		"system_resources": systemResources,
	}

	utils.SuccessResponse(c, data)
}

// GetMarketOverview 获取市场数据概览
func GetMarketOverview(c *gin.Context) {
	data := gin.H{
		"markets": []gin.H{
			{
				"symbol": "SH000300",
				"name":   "沪深300",
				"value":  3456.78,
				"change": "+1.23%",
				"trend":  "up",
			},
			{
				"symbol": "SH000905",
				"name":   "中证500",
				"value":  5678.90,
				"change": "-0.45%",
				"trend":  "down",
			},
			{
				"symbol": "SH000001",
				"name":   "上证指数",
				"value":  3123.45,
				"change": "+0.78%",
				"trend":  "up",
			},
		},
		"update_time": "2024-01-15T14:30:00Z",
	}

	utils.SuccessResponse(c, data)
}

// GetPerformanceChart 获取性能图表数据
func GetPerformanceChart(c *gin.Context) {
	chartType := c.DefaultQuery("type", "return")
	period := c.DefaultQuery("period", "1month")

	// 模拟图表数据
	var chartData interface{}
	
	switch chartType {
	case "return":
		chartData = gin.H{
			"labels": []string{"Week1", "Week2", "Week3", "Week4"},
			"datasets": []gin.H{
				{
					"label": "策略收益",
					"data":  []float64{2.3, 3.1, 1.8, 4.2},
					"color": "#1890ff",
				},
				{
					"label": "基准收益",
					"data":  []float64{1.2, 1.8, 0.9, 2.1},
					"color": "#52c41a",
				},
			},
		}
	case "drawdown":
		chartData = gin.H{
			"labels": []string{"Week1", "Week2", "Week3", "Week4"},
			"data":    []float64{-0.02, -0.05, -0.03, -0.01},
		}
	default:
		chartData = gin.H{
			"message": "Unsupported chart type",
		}
	}

	data := gin.H{
		"chart_type": chartType,
		"period":     period,
		"chart_data": chartData,
	}

	utils.SuccessResponse(c, data)
}

// GetRecentTasks 获取最近任务列表
func GetRecentTasks(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	taskService := services.NewTaskService()
	
	// 获取当前用户ID（如果已认证）
	var userID uint
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok {
			userID = id
		}
	}

	tasks, err := taskService.GetRecentTasks(limit, userID)
	if err != nil {
		utils.InternalErrorResponse(c, "获取任务列表失败: "+err.Error())
		return
	}

	data := gin.H{
		"tasks": tasks,
		"total": len(tasks),
		"limit": limit,
	}

	utils.SuccessResponse(c, data)
}