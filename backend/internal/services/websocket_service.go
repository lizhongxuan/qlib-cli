package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService WebSocket服务
type WebSocketService struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mutex      sync.RWMutex
	upgrader   websocket.Upgrader
}

// Client WebSocket客户端
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	userID   uint
	channels []string // 订阅的频道列表
	hub      *WebSocketService
}

// Message WebSocket消息
type Message struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    uint                   `json:"user_id,omitempty"`
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	Action   string   `json:"action"` // subscribe/unsubscribe
	Channels []string `json:"channels"`
}

// NewWebSocketService 创建新的WebSocket服务
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 在生产环境中应该检查具体的origin
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Run 启动WebSocket服务
func (ws *WebSocketService) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.registerClient(client)
			
		case client := <-ws.unregister:
			ws.unregisterClient(client)
			
		case message := <-ws.broadcast:
			ws.broadcastMessage(message)
		}
	}
}

// HandleWebSocket 处理WebSocket连接
func (ws *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uint) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	
	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		channels: make([]string, 0),
		hub:      ws,
	}
	
	ws.register <- client
	
	// 启动客户端协程
	go client.writePump()
	go client.readPump()
}

// BroadcastToChannel 向指定频道广播消息
func (ws *WebSocketService) BroadcastToChannel(channel string, messageType string, data map[string]interface{}) {
	message := Message{
		Type:      messageType,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now(),
	}
	
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}
	
	ws.mutex.RLock()
	for client := range ws.clients {
		if ws.isClientSubscribed(client, channel) {
			select {
			case client.send <- messageBytes:
			default:
				// 客户端发送缓冲区已满，关闭连接
				close(client.send)
				delete(ws.clients, client)
			}
		}
	}
	ws.mutex.RUnlock()
}

// BroadcastToUser 向指定用户广播消息
func (ws *WebSocketService) BroadcastToUser(userID uint, messageType string, data map[string]interface{}) {
	message := Message{
		Type:      messageType,
		Channel:   "user",
		Data:      data,
		Timestamp: time.Now(),
		UserID:    userID,
	}
	
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}
	
	ws.mutex.RLock()
	for client := range ws.clients {
		if client.userID == userID {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(ws.clients, client)
			}
		}
	}
	ws.mutex.RUnlock()
}

// BroadcastToAll 向所有客户端广播消息
func (ws *WebSocketService) BroadcastToAll(messageType string, data map[string]interface{}) {
	message := Message{
		Type:      messageType,
		Channel:   "global",
		Data:      data,
		Timestamp: time.Now(),
	}
	
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}
	
	ws.broadcast <- messageBytes
}

// GetConnectedClients 获取连接的客户端数量
func (ws *WebSocketService) GetConnectedClients() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
}

// GetClientsByChannel 获取订阅指定频道的客户端数量
func (ws *WebSocketService) GetClientsByChannel(channel string) int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	
	count := 0
	for client := range ws.clients {
		if ws.isClientSubscribed(client, channel) {
			count++
		}
	}
	return count
}

// 内部方法

// registerClient 注册客户端
func (ws *WebSocketService) registerClient(client *Client) {
	ws.mutex.Lock()
	ws.clients[client] = true
	ws.mutex.Unlock()
	
	log.Printf("用户 %d 连接到WebSocket", client.userID)
	
	// 发送欢迎消息
	welcomeMsg := Message{
		Type:      "welcome",
		Channel:   "system",
		Data: map[string]interface{}{
			"message": "WebSocket连接成功",
			"user_id": client.userID,
		},
		Timestamp: time.Now(),
	}
	
	msgBytes, _ := json.Marshal(welcomeMsg)
	select {
	case client.send <- msgBytes:
	default:
		close(client.send)
		delete(ws.clients, client)
	}
}

// unregisterClient 注销客户端
func (ws *WebSocketService) unregisterClient(client *Client) {
	ws.mutex.Lock()
	if _, ok := ws.clients[client]; ok {
		delete(ws.clients, client)
		close(client.send)
	}
	ws.mutex.Unlock()
	
	log.Printf("用户 %d 断开WebSocket连接", client.userID)
}

// broadcastMessage 广播消息
func (ws *WebSocketService) broadcastMessage(message []byte) {
	ws.mutex.RLock()
	for client := range ws.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(ws.clients, client)
		}
	}
	ws.mutex.RUnlock()
}

// isClientSubscribed 检查客户端是否订阅了指定频道
func (ws *WebSocketService) isClientSubscribed(client *Client, channel string) bool {
	for _, ch := range client.channels {
		if ch == channel {
			return true
		}
	}
	return false
}

// 客户端方法

// readPump 读取消息泵
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	// 设置读取超时
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}
		
		// 处理客户端消息
		c.handleMessage(messageBytes)
	}
}

// writePump 写入消息泵
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// 批量发送缓存的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *Client) handleMessage(messageBytes []byte) {
	var request SubscribeRequest
	if err := json.Unmarshal(messageBytes, &request); err != nil {
		log.Printf("解析客户端消息失败: %v", err)
		return
	}
	
	switch request.Action {
	case "subscribe":
		c.subscribe(request.Channels)
	case "unsubscribe":
		c.unsubscribe(request.Channels)
	case "ping":
		c.sendPong()
	default:
		log.Printf("未知的客户端操作: %s", request.Action)
	}
}

// subscribe 订阅频道
func (c *Client) subscribe(channels []string) {
	for _, channel := range channels {
		// 检查是否已订阅
		if !c.hub.isClientSubscribed(c, channel) {
			c.channels = append(c.channels, channel)
		}
	}
	
	// 发送订阅确认
	response := Message{
		Type:    "subscription_confirmed",
		Channel: "system",
		Data: map[string]interface{}{
			"subscribed_channels": channels,
			"message":            "频道订阅成功",
		},
		Timestamp: time.Now(),
	}
	
	responseBytes, _ := json.Marshal(response)
	select {
	case c.send <- responseBytes:
	default:
		// 发送失败，客户端可能已断开
	}
}

// unsubscribe 取消订阅频道
func (c *Client) unsubscribe(channels []string) {
	for _, channel := range channels {
		// 从订阅列表中移除
		for i, ch := range c.channels {
			if ch == channel {
				c.channels = append(c.channels[:i], c.channels[i+1:]...)
				break
			}
		}
	}
	
	// 发送取消订阅确认
	response := Message{
		Type:    "unsubscription_confirmed",
		Channel: "system",
		Data: map[string]interface{}{
			"unsubscribed_channels": channels,
			"message":              "频道取消订阅成功",
		},
		Timestamp: time.Now(),
	}
	
	responseBytes, _ := json.Marshal(response)
	select {
	case c.send <- responseBytes:
	default:
		// 发送失败，客户端可能已断开
	}
}

// sendPong 发送pong消息
func (c *Client) sendPong() {
	response := Message{
		Type:    "pong",
		Channel: "system",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
	
	responseBytes, _ := json.Marshal(response)
	select {
	case c.send <- responseBytes:
	default:
		// 发送失败，客户端可能已断开
	}
}

// 预定义的频道常量
const (
	ChannelTaskProgress      = "task_progress"       // 任务进度
	ChannelTaskStatus        = "task_status"         // 任务状态
	ChannelFactorTest        = "factor_test"         // 因子测试
	ChannelModelTraining     = "model_training"      // 模型训练
	ChannelStrategyBacktest  = "strategy_backtest"   // 策略回测
	ChannelWorkflowProgress  = "workflow_progress"   // 工作流进度
	ChannelSystemMonitor     = "system_monitor"      // 系统监控
	ChannelNotifications     = "notifications"       // 通知
	ChannelSystemStatus      = "system_status"       // 系统状态
	ChannelTaskLogs          = "task_logs"          // 任务日志
)

// 业务方法

// SendTaskProgress 发送任务进度更新
func (ws *WebSocketService) SendTaskProgress(userID uint, taskID uint, progress int, message string, details map[string]interface{}) {
	data := map[string]interface{}{
		"task_id":  taskID,
		"progress": progress,
		"message":  message,
		"details":  details,
	}
	ws.BroadcastToUser(userID, "task_progress", data)
}

// SendTaskStatusUpdate 发送任务状态更新
func (ws *WebSocketService) SendTaskStatusUpdate(userID uint, taskID uint, status string, message string) {
	data := map[string]interface{}{
		"task_id": taskID,
		"status":  status,
		"message": message,
	}
	ws.BroadcastToUser(userID, "task_status", data)
}

// SendFactorTestUpdate 发送因子测试更新
func (ws *WebSocketService) SendFactorTestUpdate(userID uint, factorID uint, progress int, results map[string]interface{}) {
	data := map[string]interface{}{
		"factor_id": factorID,
		"progress":  progress,
		"results":   results,
	}
	ws.BroadcastToChannel(ChannelFactorTest, "factor_test_update", data)
}

// SendModelTrainingUpdate 发送模型训练更新
func (ws *WebSocketService) SendModelTrainingUpdate(userID uint, modelID uint, progress int, metrics map[string]interface{}) {
	data := map[string]interface{}{
		"model_id": modelID,
		"progress": progress,
		"metrics":  metrics,
	}
	ws.BroadcastToChannel(ChannelModelTraining, "model_training_update", data)
}

// SendStrategyBacktestUpdate 发送策略回测更新
func (ws *WebSocketService) SendStrategyBacktestUpdate(userID uint, strategyID uint, progress int, results map[string]interface{}) {
	data := map[string]interface{}{
		"strategy_id": strategyID,
		"progress":    progress,
		"results":     results,
	}
	ws.BroadcastToChannel(ChannelStrategyBacktest, "strategy_backtest_update", data)
}

// SendWorkflowProgressUpdate 发送工作流进度更新
func (ws *WebSocketService) SendWorkflowProgressUpdate(userID uint, workflowID uint, progress int, currentStep string) {
	data := map[string]interface{}{
		"workflow_id":  workflowID,
		"progress":     progress,
		"current_step": currentStep,
	}
	ws.BroadcastToChannel(ChannelWorkflowProgress, "workflow_progress_update", data)
}

// SendSystemMonitorUpdate 发送系统监控更新
func (ws *WebSocketService) SendSystemMonitorUpdate(metrics map[string]interface{}) {
	ws.BroadcastToChannel(ChannelSystemMonitor, "system_monitor_update", metrics)
}

// SendNotification 发送通知
func (ws *WebSocketService) SendNotification(userID uint, notification map[string]interface{}) {
	ws.BroadcastToUser(userID, "notification", notification)
}

// SendSystemStatusUpdate 发送系统状态更新
func (ws *WebSocketService) SendSystemStatusUpdate(status map[string]interface{}) {
	ws.BroadcastToChannel(ChannelSystemStatus, "system_status_update", status)
}

// SendTaskLogs 发送任务日志
func (ws *WebSocketService) SendTaskLogs(userID uint, taskID uint, logs []string) {
	data := map[string]interface{}{
		"task_id": taskID,
		"logs":    logs,
	}
	ws.BroadcastToChannel(ChannelTaskLogs, "task_logs_update", data)
}