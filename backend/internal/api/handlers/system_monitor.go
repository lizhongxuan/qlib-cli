package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qlib-backend/internal/services"
	"qlib-backend/internal/utils"
)

type SystemMonitorHandler struct {
	systemMonitorService  *services.SystemMonitorService
	notificationService   *services.NotificationService
}

func NewSystemMonitorHandler(systemMonitorService *services.SystemMonitorService, notificationService *services.NotificationService) *SystemMonitorHandler {
	return &SystemMonitorHandler{
		systemMonitorService:  systemMonitorService,
		notificationService:   notificationService,
	}
}

// GetRealTimeMonitorData 获取实时监控数据
func (h *SystemMonitorHandler) GetRealTimeMonitorData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取查询参数
	metrics := c.QueryArray("metrics")
	if len(metrics) == 0 {
		metrics = []string{"cpu", "memory", "disk", "network", "tasks"}
	}
	
	intervalStr := c.DefaultQuery("interval", "5")
	interval, _ := strconv.Atoi(intervalStr)
	
	includeHistory := c.DefaultQuery("include_history", "false") == "true"

	monitorData, err := h.systemMonitorService.GetRealTimeData(userID.(uint), metrics, interval, includeHistory)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取实时监控数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, monitorData)
}

// GetSystemNotifications 获取系统通知
func (h *SystemMonitorHandler) GetSystemNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取查询参数
	unreadOnly := c.DefaultQuery("unread_only", "false") == "true"
	notificationType := c.DefaultQuery("type", "")
	priority := c.DefaultQuery("priority", "")
	
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, _ := strconv.Atoi(pageSizeStr)

	notifications, err := h.notificationService.GetNotifications(userID.(uint), unreadOnly, notificationType, priority, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取系统通知失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, notifications)
}

// MarkNotificationAsRead 标记通知已读
func (h *SystemMonitorHandler) MarkNotificationAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的通知ID")
		return
	}

	err = h.notificationService.MarkAsRead(userID.(uint), uint(notificationID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "标记通知已读失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "通知已标记为已读",
	})
}