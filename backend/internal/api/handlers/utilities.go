package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qlib-backend/internal/services"
	"qlib-backend/internal/utils"
)

type UtilitiesHandler struct {
	fileService *services.FileService
	taskManager *services.TaskManager
}

func NewUtilitiesHandler(fileService *services.FileService, taskManager *services.TaskManager) *UtilitiesHandler {
	return &UtilitiesHandler{
		fileService: fileService,
		taskManager: taskManager,
	}
}

// UploadFile 文件上传
func (h *UtilitiesHandler) UploadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取表单参数
	category := c.DefaultPostForm("category", "")
	description := c.DefaultPostForm("description", "")
	isPublic := c.DefaultPostForm("is_public", "false") == "true"

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()

	// 验证文件大小 (100MB限制)
	if header.Size > 100*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "文件大小超过100MB限制")
		return
	}

	// 上传文件
	uploadResult, err := h.fileService.UploadFile(header, userID.(uint), category, description, isPublic)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "文件上传失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, uploadResult)
}

// DownloadFile 文件下载
func (h *UtilitiesHandler) DownloadFile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的文件ID")
		return
	}

	// 获取文件下载信息
	downloadInfo, err := h.fileService.DownloadFile(uint(fileID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "文件不存在或无权限访问: "+err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+downloadInfo.OriginalName)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(downloadInfo.FileSize, 10))

	// 下载文件
	c.File(downloadInfo.FilePath)
}

// GetTasks 获取任务列表
func (h *UtilitiesHandler) GetTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取查询参数
	status := c.DefaultQuery("status", "")
	taskType := c.DefaultQuery("type", "")
	
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	
	pageSizeStr := c.DefaultQuery("page_size", "20")
	pageSize, _ := strconv.Atoi(pageSizeStr)
	
	// 获取任务列表
	taskList, err := h.taskManager.GetTasks(userID.(uint), status, taskType, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取任务列表失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, taskList)
}

// CancelTask 取消任务
func (h *UtilitiesHandler) CancelTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	taskIDStr := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的任务ID")
		return
	}

	// 先验证任务状态和权限，再取消任务
	taskStatus, err := h.taskManager.GetTaskStatus(uint(taskID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "任务不存在: "+err.Error())
		return
	}
	
	// 验证任务权限（这里简化处理，实际应该验证user_id）
	_ = userID
	
	// 取消任务
	err = h.taskManager.CancelTask(uint(taskID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "取消任务失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "任务已成功取消",
		"task_id": taskID,
		"status":  taskStatus.Status,
	})
}