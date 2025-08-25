package services

import (
	"context"
	"testing"
	"time"

	"qlib-backend/internal/testutils"
)

func TestDashboardService(t *testing.T) {
	// 设置测试环境
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 设置测试数据库
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB()

	// 创建服务实例
	service := NewDashboardService(db)

	t.Run("GetOverviewStats", func(t *testing.T) {
		ctx := context.Background()
		userID := uint(1)

		stats, err := service.GetOverviewStats(ctx, userID)
		if err != nil {
			t.Errorf("GetOverviewStats failed: %v", err)
		}

		// 验证统计数据结构
		if stats == nil {
			t.Error("Stats should not be nil")
		}

		// 验证必要字段
		requiredFields := []string{"total_models", "total_strategies", "total_factors", "total_datasets"}
		for _, field := range requiredFields {
			if _, exists := stats[field]; !exists {
				t.Errorf("Stats should contain %s field", field)
			}
		}
	})

	t.Run("GetMarketOverview", func(t *testing.T) {
		ctx := context.Background()

		overview, err := service.GetMarketOverview(ctx)
		if err != nil {
			t.Errorf("GetMarketOverview failed: %v", err)
		}

		if overview == nil {
			t.Error("Overview should not be nil")
		}

		// 验证市场概览数据结构
		requiredFields := []string{"indices", "market_summary", "top_performers"}
		for _, field := range requiredFields {
			if _, exists := overview[field]; !exists {
				t.Errorf("Overview should contain %s field", field)
			}
		}
	})

	t.Run("GetPerformanceChart", func(t *testing.T) {
		ctx := context.Background()
		userID := uint(1)
		period := "30d"

		chart, err := service.GetPerformanceChart(ctx, userID, period)
		if err != nil {
			t.Errorf("GetPerformanceChart failed: %v", err)
		}

		if chart == nil {
			t.Error("Chart should not be nil")
		}

		// 验证图表数据结构
		if data, exists := chart["data"]; !exists || data == nil {
			t.Error("Chart should contain data field")
		}

		if labels, exists := chart["labels"]; !exists || labels == nil {
			t.Error("Chart should contain labels field")
		}
	})

	t.Run("GetRecentTasks", func(t *testing.T) {
		ctx := context.Background()
		userID := uint(1)
		limit := 10

		tasks, err := service.GetRecentTasks(ctx, userID, limit)
		if err != nil {
			t.Errorf("GetRecentTasks failed: %v", err)
		}

		if tasks == nil {
			t.Error("Tasks should not be nil")
		}

		// 验证任务列表结构
		taskList, ok := tasks["tasks"].([]interface{})
		if !ok {
			t.Error("Tasks should contain tasks array")
		}

		// 如果有任务，验证任务结构
		if len(taskList) > 0 {
			if firstTask, ok := taskList[0].(map[string]interface{}); ok {
				requiredFields := []string{"id", "type", "status", "created_at"}
				for _, field := range requiredFields {
					if _, exists := firstTask[field]; !exists {
						t.Errorf("Task should contain %s field", field)
					}
				}
			}
		}
	})

	t.Run("GetSystemHealth", func(t *testing.T) {
		ctx := context.Background()

		health, err := service.GetSystemHealth(ctx)
		if err != nil {
			t.Errorf("GetSystemHealth failed: %v", err)
		}

		if health == nil {
			t.Error("Health should not be nil")
		}

		// 验证系统健康状态
		requiredFields := []string{"status", "services", "metrics"}
		for _, field := range requiredFields {
			if _, exists := health[field]; !exists {
				t.Errorf("Health should contain %s field", field)
			}
		}

		// 验证状态值
		if status, exists := health["status"]; exists {
			statusStr, ok := status.(string)
			if !ok {
				t.Error("Status should be a string")
			} else if statusStr != "healthy" && statusStr != "degraded" && statusStr != "unhealthy" {
				t.Errorf("Invalid status value: %s", statusStr)
			}
		}
	})
}

func TestDashboardServiceWithMockData(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB()

	// 创建测试数据
	user := testutils.CreateTestUser()
	dataset := testutils.CreateTestDataset()
	factor := testutils.CreateTestFactor()
	model := testutils.CreateTestModel()
	strategy := testutils.CreateTestStrategy()

	// 保存测试数据到数据库
	db.Create(user)
	db.Create(dataset)
	db.Create(factor)
	db.Create(model)
	db.Create(strategy)

	service := NewDashboardService(db)
	ctx := context.Background()

	t.Run("GetOverviewStatsWithData", func(t *testing.T) {
		stats, err := service.GetOverviewStats(ctx, user.ID)
		if err != nil {
			t.Errorf("GetOverviewStats failed: %v", err)
		}

		// 验证统计数据
		if totalModels, exists := stats["total_models"]; exists {
			if count, ok := totalModels.(int64); ok && count < 1 {
				t.Error("Total models count should be at least 1")
			}
		}

		if totalStrategies, exists := stats["total_strategies"]; exists {
			if count, ok := totalStrategies.(int64); ok && count < 1 {
				t.Error("Total strategies count should be at least 1")
			}
		}
	})

	t.Run("GetRecentTasksWithData", func(t *testing.T) {
		// 创建测试任务
		task := testutils.CreateTestTask()
		task.UserID = user.ID
		db.Create(task)

		tasks, err := service.GetRecentTasks(ctx, user.ID, 5)
		if err != nil {
			t.Errorf("GetRecentTasks failed: %v", err)
		}

		taskList, ok := tasks["tasks"].([]interface{})
		if !ok {
			t.Error("Tasks should contain tasks array")
			return
		}

		if len(taskList) == 0 {
			t.Error("Should have at least one task")
		}
	})
}

func TestDashboardServiceErrorHandling(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	// 使用无效的数据库连接测试错误处理
	service := NewDashboardService(nil)
	ctx := context.Background()

	t.Run("GetOverviewStatsWithNilDB", func(t *testing.T) {
		_, err := service.GetOverviewStats(ctx, 1)
		if err == nil {
			t.Error("Should return error with nil database")
		}
	})

	t.Run("GetRecentTasksWithNilDB", func(t *testing.T) {
		_, err := service.GetRecentTasks(ctx, 1, 10)
		if err == nil {
			t.Error("Should return error with nil database")
		}
	})
}

func TestDashboardServiceConcurrency(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB()

	service := NewDashboardService(db)
	ctx := context.Background()

	// 测试并发访问
	t.Run("ConcurrentAccess", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(userID uint) {
				defer func() { done <- true }()
				
				_, err := service.GetOverviewStats(ctx, userID)
				if err != nil {
					t.Errorf("Concurrent GetOverviewStats failed: %v", err)
				}
			}(uint(i + 1))
		}

		// 等待所有协程完成
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// 成功完成
			case <-time.After(5 * time.Second):
				t.Error("Timeout waiting for concurrent operations")
				return
			}
		}
	})
}

func TestDashboardServicePerformance(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB()

	service := NewDashboardService(db)
	ctx := context.Background()

	// 性能测试
	t.Run("PerformanceTest", func(t *testing.T) {
		start := time.Now()
		
		for i := 0; i < 100; i++ {
			_, err := service.GetOverviewStats(ctx, 1)
			if err != nil {
				t.Errorf("Performance test failed: %v", err)
			}
		}

		duration := time.Since(start)
		if duration > 10*time.Second {
			t.Errorf("Performance test took too long: %v", duration)
		}
	})
}

func TestDashboardServiceValidation(t *testing.T) {
	testutils.SetupTestEnv()
	defer testutils.CleanupTestEnv()

	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB()

	service := NewDashboardService(db)
	ctx := context.Background()

	t.Run("InvalidUserID", func(t *testing.T) {
		// 测试无效的用户ID
		_, err := service.GetOverviewStats(ctx, 0)
		if err == nil {
			t.Error("Should return error for invalid user ID")
		}
	})

	t.Run("InvalidLimit", func(t *testing.T) {
		// 测试无效的限制参数
		_, err := service.GetRecentTasks(ctx, 1, -1)
		if err == nil {
			t.Error("Should return error for negative limit")
		}

		_, err = service.GetRecentTasks(ctx, 1, 1000)
		if err == nil {
			t.Error("Should return error for too large limit")
		}
	})

	t.Run("InvalidPeriod", func(t *testing.T) {
		// 测试无效的时间周期
		_, err := service.GetPerformanceChart(ctx, 1, "invalid_period")
		if err == nil {
			t.Error("Should return error for invalid period")
		}
	})
}