package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"qlib-backend/internal/services"
	"qlib-backend/internal/utils"
)

type UILayoutHandler struct {
	uiConfigService *services.UIConfigService
}

func NewUILayoutHandler(uiConfigService *services.UIConfigService) *UILayoutHandler {
	return &UILayoutHandler{
		uiConfigService: uiConfigService,
	}
}

// GetLayoutConfig 获取界面布局配置
func (h *UILayoutHandler) GetLayoutConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取查询参数
	configType := c.DefaultQuery("type", "default")
	platform := c.DefaultQuery("platform", "web") // web, mobile, tablet
	theme := c.DefaultQuery("theme", "light")      // light, dark

	// 获取布局配置
	layoutConfig, err := h.uiConfigService.GetLayoutConfig(userID.(uint), configType, platform, theme)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取界面布局配置失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, layoutConfig)
}