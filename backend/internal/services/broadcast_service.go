package services

import (
	"encoding/json"
	"sync"
	"time"
)

// BroadcastService 广播服务
type BroadcastService struct {
	wsService   *WebSocketService
	subscribers map[string][]Subscriber
	mutex       sync.RWMutex
}

// Subscriber 订阅者
type Subscriber struct {
	ID       string
	UserID   uint
	Callback func(event Event)
}

// Event 事件
type Event struct {
	Type      string                 `json:"type"`
	Category  string                 `json:"category"`
	Data      map[string]interface{} `json:"data"`
	UserID    uint                   `json:"user_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewBroadcastService 创建新的广播服务
func NewBroadcastService(wsService *WebSocketService) *BroadcastService {
	return &BroadcastService{
		wsService:   wsService,
		subscribers: make(map[string][]Subscriber),
	}
}

// Subscribe 订阅事件
func (bs *BroadcastService) Subscribe(eventType string, subscriber Subscriber) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	if _, exists := bs.subscribers[eventType]; !exists {
		bs.subscribers[eventType] = make([]Subscriber, 0)
	}
	
	bs.subscribers[eventType] = append(bs.subscribers[eventType], subscriber)
}

// Unsubscribe 取消订阅
func (bs *BroadcastService) Unsubscribe(eventType string, subscriberID string) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	if subscribers, exists := bs.subscribers[eventType]; exists {
		for i, sub := range subscribers {
			if sub.ID == subscriberID {
				bs.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
}

// Publish 发布事件
func (bs *BroadcastService) Publish(event Event) {
	event.Timestamp = time.Now()
	
	// 通过WebSocket广播
	bs.broadcastViaWebSocket(event)
	
	// 通知订阅者
	bs.notifySubscribers(event)
}

// broadcastViaWebSocket 通过WebSocket广播事件
func (bs *BroadcastService) broadcastViaWebSocket(event Event) {
	if bs.wsService == nil {
		return
	}
	
	switch event.Category {
	case "task":
		bs.handleTaskEvent(event)
	case "factor":
		bs.handleFactorEvent(event)
	case "model":
		bs.handleModelEvent(event)
	case "strategy":
		bs.handleStrategyEvent(event)
	case "workflow":
		bs.handleWorkflowEvent(event)
	case "system":
		bs.handleSystemEvent(event)
	case "notification":
		bs.handleNotificationEvent(event)
	default:
		// 默认广播给所有用户
		if event.UserID > 0 {
			bs.wsService.BroadcastToUser(event.UserID, event.Type, event.Data)
		} else {
			bs.wsService.BroadcastToAll(event.Type, event.Data)
		}
	}
}

// handleTaskEvent 处理任务事件
func (bs *BroadcastService) handleTaskEvent(event Event) {
	switch event.Type {
	case "task_progress":
		if taskID, ok := event.Data["task_id"].(uint); ok {
			if progress, ok := event.Data["progress"].(int); ok {
				message := ""
				if msg, ok := event.Data["message"].(string); ok {
					message = msg
				}
				details := make(map[string]interface{})
				if det, ok := event.Data["details"].(map[string]interface{}); ok {
					details = det
				}
				bs.wsService.SendTaskProgress(event.UserID, taskID, progress, message, details)
			}
		}
	case "task_status":
		if taskID, ok := event.Data["task_id"].(uint); ok {
			if status, ok := event.Data["status"].(string); ok {
				message := ""
				if msg, ok := event.Data["message"].(string); ok {
					message = msg
				}
				bs.wsService.SendTaskStatusUpdate(event.UserID, taskID, status, message)
			}
		}
	case "task_logs":
		if taskID, ok := event.Data["task_id"].(uint); ok {
			if logs, ok := event.Data["logs"].([]string); ok {
				bs.wsService.SendTaskLogs(event.UserID, taskID, logs)
			}
		}
	}
}

// handleFactorEvent 处理因子事件
func (bs *BroadcastService) handleFactorEvent(event Event) {
	switch event.Type {
	case "factor_test_progress":
		if factorID, ok := event.Data["factor_id"].(uint); ok {
			if progress, ok := event.Data["progress"].(int); ok {
				results := make(map[string]interface{})
				if res, ok := event.Data["results"].(map[string]interface{}); ok {
					results = res
				}
				bs.wsService.SendFactorTestUpdate(event.UserID, factorID, progress, results)
			}
		}
	case "factor_created":
		bs.wsService.BroadcastToUser(event.UserID, "factor_created", event.Data)
	case "factor_updated":
		bs.wsService.BroadcastToUser(event.UserID, "factor_updated", event.Data)
	}
}

// handleModelEvent 处理模型事件
func (bs *BroadcastService) handleModelEvent(event Event) {
	switch event.Type {
	case "model_training_progress":
		if modelID, ok := event.Data["model_id"].(uint); ok {
			if progress, ok := event.Data["progress"].(int); ok {
				metrics := make(map[string]interface{})
				if met, ok := event.Data["metrics"].(map[string]interface{}); ok {
					metrics = met
				}
				bs.wsService.SendModelTrainingUpdate(event.UserID, modelID, progress, metrics)
			}
		}
	case "model_training_completed":
		bs.wsService.BroadcastToUser(event.UserID, "model_training_completed", event.Data)
	case "model_training_failed":
		bs.wsService.BroadcastToUser(event.UserID, "model_training_failed", event.Data)
	}
}

// handleStrategyEvent 处理策略事件
func (bs *BroadcastService) handleStrategyEvent(event Event) {
	switch event.Type {
	case "strategy_backtest_progress":
		if strategyID, ok := event.Data["strategy_id"].(uint); ok {
			if progress, ok := event.Data["progress"].(int); ok {
				results := make(map[string]interface{})
				if res, ok := event.Data["results"].(map[string]interface{}); ok {
					results = res
				}
				bs.wsService.SendStrategyBacktestUpdate(event.UserID, strategyID, progress, results)
			}
		}
	case "strategy_backtest_completed":
		bs.wsService.BroadcastToUser(event.UserID, "strategy_backtest_completed", event.Data)
	case "strategy_backtest_failed":
		bs.wsService.BroadcastToUser(event.UserID, "strategy_backtest_failed", event.Data)
	}
}

// handleWorkflowEvent 处理工作流事件
func (bs *BroadcastService) handleWorkflowEvent(event Event) {
	switch event.Type {
	case "workflow_progress":
		if workflowID, ok := event.Data["workflow_id"].(uint); ok {
			if progress, ok := event.Data["progress"].(int); ok {
				currentStep := ""
				if step, ok := event.Data["current_step"].(string); ok {
					currentStep = step
				}
				bs.wsService.SendWorkflowProgressUpdate(event.UserID, workflowID, progress, currentStep)
			}
		}
	case "workflow_completed":
		bs.wsService.BroadcastToUser(event.UserID, "workflow_completed", event.Data)
	case "workflow_failed":
		bs.wsService.BroadcastToUser(event.UserID, "workflow_failed", event.Data)
	}
}

// handleSystemEvent 处理系统事件
func (bs *BroadcastService) handleSystemEvent(event Event) {
	switch event.Type {
	case "system_monitor":
		bs.wsService.SendSystemMonitorUpdate(event.Data)
	case "system_status":
		bs.wsService.SendSystemStatusUpdate(event.Data)
	case "system_alert":
		bs.wsService.BroadcastToAll("system_alert", event.Data)
	case "maintenance_notice":
		bs.wsService.BroadcastToAll("maintenance_notice", event.Data)
	}
}

// handleNotificationEvent 处理通知事件
func (bs *BroadcastService) handleNotificationEvent(event Event) {
	switch event.Type {
	case "user_notification":
		bs.wsService.SendNotification(event.UserID, event.Data)
	case "global_notification":
		bs.wsService.BroadcastToAll("notification", event.Data)
	}
}

// notifySubscribers 通知订阅者
func (bs *BroadcastService) notifySubscribers(event Event) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	if subscribers, exists := bs.subscribers[event.Type]; exists {
		for _, subscriber := range subscribers {
			// 检查用户权限
			if event.UserID == 0 || subscriber.UserID == event.UserID {
				go subscriber.Callback(event)
			}
		}
	}
}

// 便捷方法

// PublishTaskProgress 发布任务进度事件
func (bs *BroadcastService) PublishTaskProgress(userID, taskID uint, progress int, message string, details map[string]interface{}) {
	event := Event{
		Type:     "task_progress",
		Category: "task",
		UserID:   userID,
		Data: map[string]interface{}{
			"task_id":  taskID,
			"progress": progress,
			"message":  message,
			"details":  details,
		},
	}
	bs.Publish(event)
}

// PublishTaskStatus 发布任务状态事件
func (bs *BroadcastService) PublishTaskStatus(userID, taskID uint, status, message string) {
	event := Event{
		Type:     "task_status",
		Category: "task",
		UserID:   userID,
		Data: map[string]interface{}{
			"task_id": taskID,
			"status":  status,
			"message": message,
		},
	}
	bs.Publish(event)
}

// PublishFactorTestProgress 发布因子测试进度事件
func (bs *BroadcastService) PublishFactorTestProgress(userID, factorID uint, progress int, results map[string]interface{}) {
	event := Event{
		Type:     "factor_test_progress",
		Category: "factor",
		UserID:   userID,
		Data: map[string]interface{}{
			"factor_id": factorID,
			"progress":  progress,
			"results":   results,
		},
	}
	bs.Publish(event)
}

// PublishModelTrainingProgress 发布模型训练进度事件
func (bs *BroadcastService) PublishModelTrainingProgress(userID, modelID uint, progress int, metrics map[string]interface{}) {
	event := Event{
		Type:     "model_training_progress",
		Category: "model",
		UserID:   userID,
		Data: map[string]interface{}{
			"model_id": modelID,
			"progress": progress,
			"metrics":  metrics,
		},
	}
	bs.Publish(event)
}

// PublishStrategyBacktestProgress 发布策略回测进度事件
func (bs *BroadcastService) PublishStrategyBacktestProgress(userID, strategyID uint, progress int, results map[string]interface{}) {
	event := Event{
		Type:     "strategy_backtest_progress",
		Category: "strategy",
		UserID:   userID,
		Data: map[string]interface{}{
			"strategy_id": strategyID,
			"progress":    progress,
			"results":     results,
		},
	}
	bs.Publish(event)
}

// PublishWorkflowProgress 发布工作流进度事件
func (bs *BroadcastService) PublishWorkflowProgress(userID, workflowID uint, progress int, currentStep string) {
	event := Event{
		Type:     "workflow_progress",
		Category: "workflow",
		UserID:   userID,
		Data: map[string]interface{}{
			"workflow_id":  workflowID,
			"progress":     progress,
			"current_step": currentStep,
		},
	}
	bs.Publish(event)
}

// PublishSystemMonitor 发布系统监控事件
func (bs *BroadcastService) PublishSystemMonitor(metrics map[string]interface{}) {
	event := Event{
		Type:     "system_monitor",
		Category: "system",
		Data:     metrics,
	}
	bs.Publish(event)
}

// PublishUserNotification 发布用户通知事件
func (bs *BroadcastService) PublishUserNotification(userID uint, notification map[string]interface{}) {
	event := Event{
		Type:     "user_notification",
		Category: "notification",
		UserID:   userID,
		Data:     notification,
	}
	bs.Publish(event)
}

// PublishGlobalNotification 发布全局通知事件
func (bs *BroadcastService) PublishGlobalNotification(notification map[string]interface{}) {
	event := Event{
		Type:     "global_notification",
		Category: "notification",
		Data:     notification,
	}
	bs.Publish(event)
}

// PublishSystemAlert 发布系统警报事件
func (bs *BroadcastService) PublishSystemAlert(alert map[string]interface{}) {
	event := Event{
		Type:     "system_alert",
		Category: "system",
		Data:     alert,
	}
	bs.Publish(event)
}

// GetSubscriberCount 获取订阅者数量
func (bs *BroadcastService) GetSubscriberCount(eventType string) int {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	if subscribers, exists := bs.subscribers[eventType]; exists {
		return len(subscribers)
	}
	return 0
}

// GetAllSubscriberCounts 获取所有事件类型的订阅者数量
func (bs *BroadcastService) GetAllSubscriberCounts() map[string]int {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	counts := make(map[string]int)
	for eventType, subscribers := range bs.subscribers {
		counts[eventType] = len(subscribers)
	}
	return counts
}

// ClearSubscribers 清空指定事件类型的订阅者
func (bs *BroadcastService) ClearSubscribers(eventType string) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	delete(bs.subscribers, eventType)
}

// ClearAllSubscribers 清空所有订阅者
func (bs *BroadcastService) ClearAllSubscribers() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.subscribers = make(map[string][]Subscriber)
}

// GetEventHistory 获取事件历史（可以扩展为持久化存储）
type EventHistory struct {
	events []Event
	mutex  sync.RWMutex
	maxSize int
}

// NewEventHistory 创建事件历史记录
func NewEventHistory(maxSize int) *EventHistory {
	if maxSize <= 0 {
		maxSize = 1000 // 默认最大1000个事件
	}
	return &EventHistory{
		events:  make([]Event, 0),
		maxSize: maxSize,
	}
}

// AddEvent 添加事件到历史记录
func (eh *EventHistory) AddEvent(event Event) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
	
	eh.events = append(eh.events, event)
	
	// 保持最大数量限制
	if len(eh.events) > eh.maxSize {
		eh.events = eh.events[1:]
	}
}

// GetEvents 获取事件历史
func (eh *EventHistory) GetEvents(eventType string, limit int) []Event {
	eh.mutex.RLock()
	defer eh.mutex.RUnlock()
	
	var filteredEvents []Event
	for i := len(eh.events) - 1; i >= 0 && len(filteredEvents) < limit; i-- {
		if eventType == "" || eh.events[i].Type == eventType {
			filteredEvents = append(filteredEvents, eh.events[i])
		}
	}
	
	return filteredEvents
}

// GetEventsJSON 获取事件历史的JSON格式
func (eh *EventHistory) GetEventsJSON(eventType string, limit int) ([]byte, error) {
	events := eh.GetEvents(eventType, limit)
	return json.Marshal(events)
}