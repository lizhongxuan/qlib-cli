package services

import (
	"encoding/json"
	"fmt"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db               *gorm.DB
	broadcastService *BroadcastService
}

// Notification 通知结构
type Notification struct {
	ID          uint                   `json:"id"`
	UserID      uint                   `json:"user_id"`
	Type        string                 `json:"type"`        // system, task, alert, info
	Priority    string                 `json:"priority"`    // low, normal, high, critical
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"` // 额外数据
	IsRead      bool                   `json:"is_read"`
	CreatedAt   time.Time              `json:"created_at"`
	ActionURL   string                 `json:"action_url"`
	TaskID      *uint                  `json:"task_id,omitempty"`
}

// NotificationList 通知列表响应
type NotificationList struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
	UnreadCount   int64          `json:"unread_count"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	TotalPages    int64          `json:"total_pages"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalCount   int64                  `json:"total_count"`
	UnreadCount  int64                  `json:"unread_count"`
	TypeCounts   map[string]int64       `json:"type_counts"`
	PriorityCounts map[string]int64     `json:"priority_counts"`
	RecentNotifications []Notification  `json:"recent_notifications"`
}

// NotificationRequest 创建通知请求
type NotificationRequest struct {
	UserID    uint                   `json:"user_id" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Priority  string                 `json:"priority"`
	Title     string                 `json:"title" binding:"required"`
	Message   string                 `json:"message" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	ExpiresAt *time.Time             `json:"expires_at"`
}

// NewNotificationService 创建新的通知服务
func NewNotificationService(db *gorm.DB, broadcastService *BroadcastService) *NotificationService {
	// 自动迁移通知表
	db.AutoMigrate(&models.Notification{})
	
	return &NotificationService{
		db:               db,
		broadcastService: broadcastService,
	}
}

// GetNotifications 获取通知列表
func (ns *NotificationService) GetNotifications(userID uint, unreadOnly bool, notificationType, priority string, page, pageSize int) (*NotificationList, error) {
	var notifications []models.Notification
	var total, unreadCount int64
	
	// 构建查询
	query := ns.db.Model(&models.Notification{}).Where("user_id = ?", userID)
	
	if unreadOnly {
		query = query.Where("is_read = false")
	}
	
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}
	
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	
	// 过滤已过期的通知
	query = query.Where("expires_at IS NULL OR expires_at > ?", time.Now())
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取通知总数失败: %v", err)
	}
	
	// 获取未读数量
	unreadQuery := ns.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = false", userID)
	if err := unreadQuery.Count(&unreadCount).Error; err != nil {
		return nil, fmt.Errorf("获取未读通知数失败: %v", err)
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("获取通知列表失败: %v", err)
	}
	
	// 转换为响应格式
	notificationList := make([]Notification, len(notifications))
	for i, notification := range notifications {
		notificationList[i] = Notification{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Type:      notification.Type,
			Priority:  notification.Priority,
			Title:     notification.Title,
			Message:   notification.Message,
			IsRead:    notification.IsRead,
			CreatedAt: notification.CreatedAt,
			ActionURL: notification.ActionURL,
			TaskID:    notification.TaskID,
		}
		
		// 反序列化 Data 字段
		if notification.Data != "" {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(notification.Data), &data); err == nil {
				notificationList[i].Data = data
			}
		}
	}
	
	return &NotificationList{
		Notifications: notificationList,
		Total:         total,
		UnreadCount:   unreadCount,
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    (total + int64(pageSize) - 1) / int64(pageSize),
	}, nil
}

// CreateNotification 创建通知
func (ns *NotificationService) CreateNotification(req NotificationRequest) (*Notification, error) {
	// 序列化数据
	var dataJSON string
	if req.Data != nil {
		dataBytes, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("序列化通知数据失败: %v", err)
		}
		dataJSON = string(dataBytes)
	}
	
	// 设置默认优先级
	if req.Priority == "" {
		req.Priority = "normal"
	}
	
	// 创建通知记录
	notification := &models.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Priority:  req.Priority,
		Title:     req.Title,
		Message:   req.Message,
		Data:      dataJSON,
		IsRead:    false,
		ExpiresAt: req.ExpiresAt,
	}
	
	if err := ns.db.Create(notification).Error; err != nil {
		return nil, fmt.Errorf("创建通知失败: %v", err)
	}
	
	// 转换为响应格式
	result := &Notification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Priority:  notification.Priority,
		Title:     notification.Title,
		Message:   notification.Message,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
		ActionURL: notification.ActionURL,
		TaskID:    notification.TaskID,
	}
	
	// 反序列化 Data 字段
	if notification.Data != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(notification.Data), &data); err == nil {
			result.Data = data
		}
	}
	
	// 异步推送通知
	go func() {
		ns.broadcastNotification(result)
	}()
	
	return result, nil
}

// MarkAsRead 标记通知为已读
func (ns *NotificationService) MarkAsRead(userID, notificationID uint) error {
	now := time.Now()
	result := ns.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})
	
	if result.Error != nil {
		return fmt.Errorf("标记通知已读失败: %v", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("通知不存在或无权限访问")
	}
	
	return nil
}

// MarkAllAsRead 标记所有通知为已读
func (ns *NotificationService) MarkAllAsRead(userID uint) error {
	now := time.Now()
	if err := ns.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error; err != nil {
		return fmt.Errorf("标记所有通知已读失败: %v", err)
	}
	
	return nil
}

// DeleteNotification 删除通知
func (ns *NotificationService) DeleteNotification(userID, notificationID uint) error {
	result := ns.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{})
	
	if result.Error != nil {
		return fmt.Errorf("删除通知失败: %v", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("通知不存在或无权限访问")
	}
	
	return nil
}

// GetNotificationStats 获取通知统计
func (ns *NotificationService) GetNotificationStats(userID uint) (*NotificationStats, error) {
	stats := &NotificationStats{
		TypeCounts:     make(map[string]int64),
		PriorityCounts: make(map[string]int64),
	}
	
	// 总通知数
	if err := ns.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&stats.TotalCount).Error; err != nil {
		return nil, fmt.Errorf("获取总通知数失败: %v", err)
	}
	
	// 未读通知数
	if err := ns.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&stats.UnreadCount).Error; err != nil {
		return nil, fmt.Errorf("获取未读通知数失败: %v", err)
	}
	
	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	if err := ns.db.Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("按类型统计失败: %v", err)
	}
	
	for _, stat := range typeStats {
		stats.TypeCounts[stat.Type] = stat.Count
	}
	
	// 按优先级统计
	var priorityStats []struct {
		Priority string `json:"priority"`
		Count    int64  `json:"count"`
	}
	if err := ns.db.Model(&models.Notification{}).
		Select("priority, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("priority").
		Scan(&priorityStats).Error; err != nil {
		return nil, fmt.Errorf("按优先级统计失败: %v", err)
	}
	
	for _, stat := range priorityStats {
		stats.PriorityCounts[stat.Priority] = stat.Count
	}
	
	// 最近通知
	var recentNotifications []models.Notification
	if err := ns.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentNotifications).Error; err != nil {
		return nil, fmt.Errorf("获取最近通知失败: %v", err)
	}
	
	// 转换格式
	stats.RecentNotifications = make([]Notification, len(recentNotifications))
	for i, notification := range recentNotifications {
		stats.RecentNotifications[i] = Notification{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Type:      notification.Type,
			Priority:  notification.Priority,
			Title:     notification.Title,
			Message:   notification.Message,
			IsRead:    notification.IsRead,
			CreatedAt: notification.CreatedAt,
			ActionURL: notification.ActionURL,
			TaskID:    notification.TaskID,
		}
	}
	
	return stats, nil
}

// CleanupExpiredNotifications 清理过期通知
func (ns *NotificationService) CleanupExpiredNotifications() error {
	result := ns.db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&models.Notification{})
	
	if result.Error != nil {
		return fmt.Errorf("清理过期通知失败: %v", result.Error)
	}
	
	return nil
}

// 便利方法

// CreateSystemNotification 创建系统通知
func (ns *NotificationService) CreateSystemNotification(userID uint, title, message string, priority string) error {
	req := NotificationRequest{
		UserID:   userID,
		Type:     "system",
		Priority: priority,
		Title:    title,
		Message:  message,
	}
	
	_, err := ns.CreateNotification(req)
	return err
}

// CreateTaskNotification 创建任务通知
func (ns *NotificationService) CreateTaskNotification(userID uint, taskID uint, title, message string) error {
	req := NotificationRequest{
		UserID:   userID,
		Type:     "task",
		Priority: "normal",
		Title:    title,
		Message:  message,
		Data: map[string]interface{}{
			"task_id": taskID,
		},
	}
	
	_, err := ns.CreateNotification(req)
	return err
}

// CreateAlertNotification 创建告警通知
func (ns *NotificationService) CreateAlertNotification(userID uint, alertType, title, message string) error {
	priority := "normal"
	if alertType == "critical" {
		priority = "critical"
	} else if alertType == "warning" {
		priority = "high"
	}
	
	req := NotificationRequest{
		UserID:   userID,
		Type:     "alert",
		Priority: priority,
		Title:    title,
		Message:  message,
		Data: map[string]interface{}{
			"alert_type": alertType,
		},
	}
	
	_, err := ns.CreateNotification(req)
	return err
}

// 内部方法

// broadcastNotification 广播通知
func (ns *NotificationService) broadcastNotification(notification *Notification) {
	if ns.broadcastService != nil {
		event := Event{
			Type:      "notification",
			Category:  "notification",
			Data:      map[string]interface{}{"notification": notification},
			UserID:    notification.UserID,
			Timestamp: time.Now(),
		}
		ns.broadcastService.Publish(event)
	}
}