package routes

import (
	"qlib-backend/internal/api/handlers"
	"qlib-backend/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine) {
	// 添加日志中间件
	r.Use(middleware.Logger())
	
	// 添加恢复中间件
	r.Use(middleware.Recovery())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 系统总览 API
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/overview", handlers.GetDashboardOverview)
			dashboard.GET("/market-overview", handlers.GetMarketOverview)
			dashboard.GET("/performance-chart", handlers.GetPerformanceChart)
			dashboard.GET("/recent-tasks", handlers.GetRecentTasks)
		}

		// 数据管理 API
		data := v1.Group("/data")
		{
			data.GET("/datasets", handlers.GetDatasets)
			data.POST("/datasets", handlers.CreateDataset)
			data.PUT("/datasets/:id", handlers.UpdateDataset)
			data.DELETE("/datasets/:id", handlers.DeleteDataset)
			data.GET("/sources", handlers.GetDataSources)
			data.POST("/sources/test-connection", handlers.TestDataSourceConnection)
			data.GET("/explore/:dataset_id", handlers.ExploreDataset)
			data.POST("/upload", handlers.UploadData)
		}

		// 因子管理 API
		factors := v1.Group("/factors")
		{
			factors.GET("", handlers.GetFactors)
			factors.POST("", handlers.CreateFactor)
			factors.PUT("/:id", handlers.UpdateFactor)
			factors.DELETE("/:id", handlers.DeleteFactor)
			factors.POST("/test", handlers.TestFactor)
			factors.GET("/:id/analysis", handlers.GetFactorAnalysis)
			factors.POST("/batch-test", handlers.BatchTestFactors)
			factors.GET("/categories", handlers.GetFactorCategories)
			factors.POST("/import", handlers.ImportFactors)
			
			// 因子研究工作台 API
			factors.POST("/ai-chat", handlers.FactorAIChat)
			factors.POST("/validate-syntax", handlers.ValidateFactorSyntax)
			factors.GET("/qlib-functions", handlers.GetQlibFunctions)
			factors.GET("/syntax-reference", handlers.GetSyntaxReference)
			factors.POST("/save-workspace", handlers.SaveWorkspaceFactor)
		}

		// 模型训练 API
		models := v1.Group("/models")
		{
			models.POST("/train", handlers.StartModelTraining)
			models.GET("", handlers.GetModels)
			models.GET("/:id/progress", handlers.GetTrainingProgress)
			models.POST("/:id/stop", handlers.StopTraining)
			models.GET("/:id/evaluate", handlers.EvaluateModel)
			models.POST("/compare", handlers.CompareModels)
			models.POST("/:id/deploy", handlers.DeployModel)
			models.GET("/:id/logs", handlers.GetTrainingLogs)
		}

		// 策略回测 API
		strategies := v1.Group("/strategies")
		{
			strategies.POST("/backtest", handlers.StartStrategyBacktest)
			strategies.GET("", handlers.GetStrategies)
			strategies.GET("/:id/results", handlers.GetBacktestResults)
			strategies.GET("/:id/progress", handlers.GetBacktestProgress)
			strategies.POST("/:id/stop", handlers.StopBacktest)
			strategies.GET("/:id/attribution", handlers.GetStrategyAttribution)
			strategies.POST("/compare", handlers.CompareStrategies)
			strategies.POST("/:id/optimize", handlers.OptimizeParameters)
			strategies.POST("/export", handlers.ExportBacktestReport)
		}

		// Qlib工作流 API
		qlib := v1.Group("/qlib")
		{
			workflow := qlib.Group("/workflow")
			{
				workflow.POST("/run", handlers.RunQlibWorkflow)
				workflow.GET("/templates", handlers.GetWorkflowTemplates)
				workflow.POST("/create-template", handlers.CreateWorkflowTemplate)
				workflow.GET("/:task_id/status", handlers.GetWorkflowStatus)
				workflow.POST("/:task_id/pause", handlers.PauseWorkflow)
				workflow.POST("/:task_id/resume", handlers.ResumeWorkflow)
				workflow.GET("/history", handlers.GetWorkflowHistory)
			}
		}

		// 工作流配置向导 API
		workflow := v1.Group("/workflow")
		{
			workflow.GET("/templates", handlers.GetWorkflowConfigTemplates)
			workflow.POST("/validate-config", handlers.ValidateWorkflowConfig)
			workflow.POST("/generate-yaml", handlers.GenerateWorkflowYAML)
			workflow.GET("/progress/:task_id", handlers.GetWorkflowProgress)
		}

		// 结果分析 API
		analysis := v1.Group("/analysis")
		{
			analysis.GET("/overview", handlers.GetAnalysisOverview)
			analysis.POST("/models/compare", handlers.CompareModelPerformance)
			analysis.GET("/models/:result_id/factor-importance", handlers.GetFactorImportance)
			analysis.GET("/strategies/:result_id/performance", handlers.GetStrategyPerformance)
			analysis.POST("/strategies/compare", handlers.CompareStrategyPerformance)
			analysis.POST("/reports/generate", handlers.GenerateAnalysisReport)
			analysis.GET("/reports/:task_id/status", handlers.GetReportStatus)
			analysis.GET("/results/summary-stats", handlers.GetResultsSummaryStats)
			analysis.POST("/results/multi-compare", handlers.MultiCompareResults)
		}

		// 回测结果展示增强 API
		backtest := v1.Group("/backtest")
		{
			results := backtest.Group("/results")
			{
				results.GET("/:result_id/detailed", handlers.GetDetailedBacktestResults)
			}
			backtest.GET("/charts/:result_id/:chart_type", handlers.GetBacktestChartData)
			backtest.POST("/export-report", handlers.ExportDetailedBacktestReport)
		}

		// 系统监控增强 API
		system := v1.Group("/system")
		{
			monitor := system.Group("/monitor")
			{
				monitor.GET("/real-time", handlers.GetRealTimeMonitorData)
			}
			system.GET("/notifications", handlers.GetSystemNotifications)
			system.PUT("/notifications/:id/read", handlers.MarkNotificationRead)
		}

		// 通用工具 API
		files := v1.Group("/files")
		{
			files.POST("/upload", handlers.UploadFile)
			files.GET("/:file_id/download", handlers.DownloadFile)
		}

		tasks := v1.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)
			tasks.POST("/:task_id/cancel", handlers.CancelTask)
		}

		// 布局和用户界面 API
		ui := v1.Group("/ui")
		{
			layout := ui.Group("/layout")
			{
				layout.GET("/config", handlers.GetUILayoutConfig)
			}
		}
	}

	// WebSocket 路由
	ws := r.Group("/ws")
	{
		ws.GET("/workflow-progress/:task_id", handlers.HandleWorkflowProgressWS)
		ws.GET("/factor-test/:test_id", handlers.HandleFactorTestWS)
		ws.GET("/system-monitor", handlers.HandleSystemMonitorWS)
		ws.GET("/notifications", handlers.HandleNotificationsWS)
		ws.GET("/task/:task_id", handlers.HandleTaskStatusWS)
		ws.GET("/system", handlers.HandleSystemStatusWS)
		ws.GET("/logs/:task_id", handlers.HandleTaskLogsWS)
	}
}