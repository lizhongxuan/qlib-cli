package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该更严格
	},
}

// HandleWorkflowProgressWS 工作流进度WebSocket
func HandleWorkflowProgressWS(c *gin.Context) {
	taskID := c.Param("task_id")
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to workflow progress: %s", taskID)

	// 发送连接确认
	conn.WriteJSON(map[string]interface{}{
		"event": "connection_status",
		"data": map[string]interface{}{
			"status":      "connected",
			"task_id":     taskID,
			"server_time": time.Now().Format(time.RFC3339),
		},
	})

	// 模拟发送进度更新
	for i := 0; i <= 100; i += 10 {
		time.Sleep(2 * time.Second)
		
		err := conn.WriteJSON(map[string]interface{}{
			"event": "progress_update",
			"data": map[string]interface{}{
				"task_id":        taskID,
				"status":         "running",
				"progress":       i,
				"current_step":   getWorkflowStep(i),
				"estimated_time": (100 - i) * 12, // 剩余时间估算
				"timestamp":      time.Now().Format(time.RFC3339),
				"log_message":    getWorkflowLogMessage(i),
			},
		})
		
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
		
		if i == 100 {
			conn.WriteJSON(map[string]interface{}{
				"event": "progress_update",
				"data": map[string]interface{}{
					"task_id":   taskID,
					"status":    "completed",
					"progress":  100,
					"timestamp": time.Now().Format(time.RFC3339),
				},
			})
		}
	}
}

// HandleFactorTestWS 因子测试进度WebSocket
func HandleFactorTestWS(c *gin.Context) {
	testID := c.Param("test_id")
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to factor test: %s", testID)

	// 模拟因子测试进度
	phases := []string{"validation", "data_loading", "calculation", "analysis", "completed"}
	for i, phase := range phases {
		time.Sleep(1 * time.Second)
		
		progress := (i + 1) * 20
		
		err := conn.WriteJSON(map[string]interface{}{
			"event": "test_progress",
			"data": map[string]interface{}{
				"test_id":      testID,
				"factor_name":  "Custom Momentum Factor",
				"progress":     progress,
				"current_phase": getFactorTestPhase(phase),
				"partial_results": map[string]interface{}{
					"ic":               0.0356,
					"periods_processed": i * 50,
					"total_periods":     250,
				},
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})
		
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// HandleSystemMonitorWS 系统监控WebSocket
func HandleSystemMonitorWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to system monitor")

	// 每5秒发送一次系统状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := conn.WriteJSON(map[string]interface{}{
				"event": "system_status",
				"data": map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"cpu": map[string]interface{}{
						"usage":       65.2 + float64(time.Now().Second()%10),
						"cores":       8,
						"load_avg":    []float64{1.2, 1.5, 1.8},
						"temperature": 68,
					},
					"memory": map[string]interface{}{
						"usage":       78.5 + float64(time.Now().Second()%5),
						"total_gb":    16,
						"available_gb": 3.5,
						"swap_usage":  12.3,
					},
					"disk": map[string]interface{}{
						"usage":      45.2,
						"total_gb":   500,
						"available_gb": 275,
						"io_read":    "125MB/s",
						"io_write":   "87MB/s",
					},
					"network": map[string]interface{}{
						"status":         "online",
						"upload_speed":   "1.2MB/s",
						"download_speed": "5.8MB/s",
						"latency":        15,
					},
					"qlib_services": map[string]interface{}{
						"data_provider":   "connected",
						"cache_status":    "healthy",
						"cache_size_gb":   2.3,
						"last_data_update": "2024-01-15T09:30:00Z",
						"active_tasks":    3,
						"queue_length":    2,
					},
					"alerts": []map[string]interface{}{},
				},
			})
			
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// HandleNotificationsWS 通知WebSocket
func HandleNotificationsWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to notifications")

	// 模拟发送通知
	notifications := []map[string]interface{}{
		{
			"id":       123,
			"type":     "success",
			"category": "task_completion",
			"title":    "模型训练完成",
			"message":  "LightGBM模型训练已成功完成，测试IC: 0.0456",
			"timestamp": time.Now().Format(time.RFC3339),
			"action_url": "/models/lgb_model_123",
			"auto_dismiss": false,
			"priority": "normal",
			"related_task_id": "workflow_task_123",
		},
		{
			"id":       124,
			"type":     "info",
			"category": "data_update",
			"title":    "数据更新",
			"message":  "CSI300数据已更新到最新交易日",
			"timestamp": time.Now().Add(10*time.Second).Format(time.RFC3339),
			"priority": "low",
		},
	}

	for i, notification := range notifications {
		time.Sleep(time.Duration(i*10) * time.Second)
		
		err := conn.WriteJSON(map[string]interface{}{
			"event": "notification",
			"data":  notification,
		})
		
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// HandleTaskStatusWS 任务状态WebSocket
func HandleTaskStatusWS(c *gin.Context) {
	taskID := c.Param("task_id")
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to task status: %s", taskID)

	// 模拟任务状态更新
	for i := 0; i <= 100; i += 20 {
		time.Sleep(2 * time.Second)
		
		status := "running"
		if i == 100 {
			status = "completed"
		}
		
		err := conn.WriteJSON(map[string]interface{}{
			"event": "task_status",
			"data": map[string]interface{}{
				"task_id":   taskID,
				"status":    status,
				"progress":  i,
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})
		
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// HandleSystemStatusWS 系统状态WebSocket
func HandleSystemStatusWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to system status")

	// 发送系统状态
	err = conn.WriteJSON(map[string]interface{}{
		"event": "system_status",
		"data": map[string]interface{}{
			"status":  "healthy",
			"uptime":  "2days 3hours",
			"version": "v1.0.0",
			"components": map[string]string{
				"database":  "healthy",
				"qlib_data": "healthy",
				"cache":     "healthy",
			},
		},
	})
	
	if err != nil {
		log.Printf("WebSocket write error: %v", err)
		return
	}
}

// HandleTaskLogsWS 任务日志WebSocket
func HandleTaskLogsWS(c *gin.Context) {
	taskID := c.Param("task_id")
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected to task logs: %s", taskID)

	// 模拟发送日志
	logs := []string{
		"开始初始化Qlib环境",
		"加载数据集完成，共100万条记录",
		"开始模型训练...",
		"Epoch 1/100 完成，train_loss: 0.0567",
		"Epoch 2/100 完成，train_loss: 0.0534",
		"模型训练完成",
	}

	for i, logMessage := range logs {
		time.Sleep(2 * time.Second)
		
		err := conn.WriteJSON(map[string]interface{}{
			"event": "log_message",
			"data": map[string]interface{}{
				"task_id":   taskID,
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "INFO",
				"message":   logMessage,
				"line":      i + 1,
			},
		})
		
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// 辅助函数

func getWorkflowStep(progress int) string {
	switch {
	case progress <= 10:
		return "初始化Qlib环境..."
	case progress <= 25:
		return "加载数据和因子..."
	case progress <= 45:
		return "训练模型..."
	case progress <= 70:
		return "模型验证..."
	case progress <= 85:
		return "策略回测..."
	default:
		return "生成分析报告..."
	}
}

func getWorkflowLogMessage(progress int) string {
	switch {
	case progress <= 10:
		return "Qlib环境初始化完成"
	case progress <= 25:
		return "数据加载完成，共100万条记录"
	case progress <= 45:
		return "模型训练中，当前损失: 0.0234"
	case progress <= 70:
		return "模型验证完成，IC: 0.0456"
	case progress <= 85:
		return "策略回测中，当前收益: 12.3%"
	default:
		return "分析报告生成完成"
	}
}

func getFactorTestPhase(phase string) string {
	phases := map[string]string{
		"validation":   "语法验证中...",
		"data_loading": "数据加载中...",
		"calculation":  "IC计算中...",
		"analysis":     "性能分析中...",
		"completed":    "测试完成",
	}
	if desc, exists := phases[phase]; exists {
		return desc
	}
	return phase
}