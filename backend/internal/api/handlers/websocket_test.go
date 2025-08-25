package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestWebSocketHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		endpoint       string
		handler        gin.HandlerFunc
		expectedEvents []string
	}{
		{
			name:           "工作流进度WebSocket",
			endpoint:       "/ws/workflow-progress/123",
			handler:        HandleWorkflowProgressWS,
			expectedEvents: []string{"connection_status", "progress_update"},
		},
		{
			name:           "因子测试WebSocket",
			endpoint:       "/ws/factor-test/456",
			handler:        HandleFactorTestWS,
			expectedEvents: []string{"test_progress"},
		},
		{
			name:           "系统监控WebSocket",
			endpoint:       "/ws/system-monitor",
			handler:        HandleSystemMonitorWS,
			expectedEvents: []string{"system_status"},
		},
		{
			name:           "通知WebSocket",
			endpoint:       "/ws/notifications",
			handler:        HandleNotificationsWS,
			expectedEvents: []string{"notification"},
		},
		{
			name:           "任务状态WebSocket",
			endpoint:       "/ws/task/789",
			handler:        HandleTaskStatusWS,
			expectedEvents: []string{"task_status"},
		},
		{
			name:           "系统状态WebSocket",
			endpoint:       "/ws/system-status",
			handler:        HandleSystemStatusWS,
			expectedEvents: []string{"system_status"},
		},
		{
			name:           "任务日志WebSocket",
			endpoint:       "/ws/logs/321",
			handler:        HandleTaskLogsWS,
			expectedEvents: []string{"log_message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务器
			router := gin.New()
			
			// 根据不同的endpoint设置路由
			switch {
			case strings.Contains(tt.endpoint, "workflow-progress"):
				router.GET("/ws/workflow-progress/:task_id", tt.handler)
			case strings.Contains(tt.endpoint, "factor-test"):
				router.GET("/ws/factor-test/:test_id", tt.handler)
			case strings.Contains(tt.endpoint, "task/"):
				router.GET("/ws/task/:task_id", tt.handler)
			case strings.Contains(tt.endpoint, "logs"):
				router.GET("/ws/logs/:task_id", tt.handler)
			default:
				router.GET(tt.endpoint, tt.handler)
			}

			server := httptest.NewServer(router)
			defer server.Close()

			// 将HTTP URL转换为WebSocket URL
			u, _ := url.Parse(server.URL)
			u.Scheme = "ws"
			u.Path = tt.endpoint

			// 创建WebSocket连接
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				t.Fatalf("WebSocket连接失败: %v", err)
			}
			defer conn.Close()

			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			// 读取消息并验证
			receivedEvents := make(map[string]bool)
			messageCount := 0
			maxMessages := 5 // 限制最大消息数量，避免无限等待

			for messageCount < maxMessages {
				var message map[string]interface{}
				err := conn.ReadJSON(&message)
				if err != nil {
					// 对于某些WebSocket handler，连接可能会自动关闭
					if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						break
					}
					t.Logf("读取WebSocket消息时出错: %v", err)
					break
				}

				// 验证消息格式
				assert.Contains(t, message, "event", "消息应包含event字段")
				if event, ok := message["event"].(string); ok {
					receivedEvents[event] = true
					t.Logf("收到事件: %s", event)
				}

				messageCount++

				// 对于系统状态WebSocket，只需要验证一个消息
				if tt.endpoint == "/ws/system-status" && messageCount >= 1 {
					break
				}
			}

			// 验证是否收到了期望的事件类型
			foundExpectedEvent := false
			for _, expectedEvent := range tt.expectedEvents {
				if receivedEvents[expectedEvent] {
					foundExpectedEvent = true
					break
				}
			}

			assert.True(t, foundExpectedEvent, "应该收到期望的事件类型之一: %v, 实际收到: %v", tt.expectedEvents, receivedEvents)
			assert.Greater(t, messageCount, 0, "应该收到至少一条消息")
		})
	}
}

func TestWebSocketUpgradeFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/ws/test", HandleWorkflowProgressWS)

	// 创建HTTP请求（不是WebSocket升级请求）
	req, _ := http.NewRequest("GET", "/ws/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// WebSocket升级应该失败，返回错误状态码
	assert.NotEqual(t, http.StatusSwitchingProtocols, w.Code)
}

func TestWebSocketHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(int) string
		input    int
		expected string
	}{
		{
			name:     "getWorkflowStep - 初始化阶段",
			function: getWorkflowStep,
			input:    5,
			expected: "初始化Qlib环境...",
		},
		{
			name:     "getWorkflowStep - 数据加载阶段",
			function: getWorkflowStep,
			input:    20,
			expected: "加载数据和因子...",
		},
		{
			name:     "getWorkflowStep - 模型训练阶段",
			function: getWorkflowStep,
			input:    40,
			expected: "训练模型...",
		},
		{
			name:     "getWorkflowStep - 完成阶段",
			function: getWorkflowStep,
			input:    100,
			expected: "生成分析报告...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetWorkflowLogMessage(t *testing.T) {
	tests := []struct {
		progress int
		expected string
	}{
		{5, "Qlib环境初始化完成"},
		{20, "数据加载完成，共100万条记录"},
		{40, "模型训练中，当前损失: 0.0234"},
		{65, "模型验证完成，IC: 0.0456"},
		{80, "策略回测中，当前收益: 12.3%"},
		{100, "分析报告生成完成"},
	}

	for _, tt := range tests {
		t.Run("progress_"+string(rune(tt.progress)), func(t *testing.T) {
			result := getWorkflowLogMessage(tt.progress)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFactorTestPhase(t *testing.T) {
	tests := []struct {
		phase    string
		expected string
	}{
		{"validation", "语法验证中..."},
		{"data_loading", "数据加载中..."},
		{"calculation", "IC计算中..."},
		{"analysis", "性能分析中..."},
		{"completed", "测试完成"},
		{"unknown", "unknown"}, // 未知阶段应返回原值
	}

	for _, tt := range tests {
		t.Run("phase_"+tt.phase, func(t *testing.T) {
			result := getFactorTestPhase(tt.phase)
			assert.Equal(t, tt.expected, result)
		})
	}
}