package handlers

import (
	"net/http"
	"qlib-backend/internal/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"qlib-backend/internal/services"
	"qlib-backend/internal/utils"
)

type AnalysisHandler struct {
	analysisService *services.AnalysisService
	reportService   *services.ReportService
}

func NewAnalysisHandler(analysisService *services.AnalysisService, reportService *services.ReportService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
		reportService:   reportService,
	}
}

// GetAnalysisOverview 获取分析结果概览
func (h *AnalysisHandler) GetAnalysisOverview(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	overview, err := h.analysisService.GetAnalysisOverview(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取分析概览失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, overview)
}

// CompareModels 模型性能对比
func (h *AnalysisHandler) CompareModels(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var req struct {
		ModelIDs    []uint   `json:"model_ids" binding:"required,min=2"`
		Metrics     []string `json:"metrics"`      // 对比指标 - 传递给service层使用
		CompareType string   `json:"compare_type"` // 对比类型 - 映射到service层的Granularity参数
		TimeRange   string   `json:"time_range"`   // 时间范围 - 用于指定对比的时间段
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	compareReq := models.ModelComparisonRequest{
		ModelIDs:    req.ModelIDs,
		Metrics:     req.Metrics,
		TimeRange:   req.TimeRange,   // 正确使用TimeRange参数
		Granularity: req.CompareType, // CompareType映射到Granularity
	}
	comparison, err := h.analysisService.CompareModels(compareReq, userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "模型对比失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, comparison)
}

// GetFactorImportance 因子重要性分析
func (h *AnalysisHandler) GetFactorImportance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	resultIDStr := c.Param("result_id")
	resultID, err := strconv.ParseUint(resultIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的结果ID")
		return
	}

	method := c.DefaultQuery("method", "default")
	topN := c.DefaultQuery("top_n", "20")
	topNInt, _ := strconv.Atoi(topN)

	factorReq := services.FactorImportanceRequest{
		ModelID:  uint(resultID),
		ResultID: uint(resultID),
		TopN:     topNInt,
		Method:   method,
	}
	importance, err := h.analysisService.GetFactorImportance(factorReq, userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取因子重要性失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, importance)
}

// GetStrategyPerformance 策略绩效分析
func (h *AnalysisHandler) GetStrategyPerformance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	resultIDStr := c.Param("result_id")
	resultID, err := strconv.ParseUint(resultIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的结果ID")
		return
	}

	performance, err := h.analysisService.GetStrategyPerformance(uint(resultID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取策略绩效失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, performance)
}

// CompareStrategies 多策略对比
func (h *AnalysisHandler) CompareStrategies(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var req struct {
		StrategyIDs []uint   `json:"strategy_ids" binding:"required,min=2"`
		Metrics     []string `json:"metrics"`
		CompareType string   `json:"compare_type"`
		TimeRange   string   `json:"time_range"`
		Benchmark   string   `json:"benchmark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	comparison, err := h.analysisService.CompareStrategies(userID.(uint), req.StrategyIDs, req.Metrics, req.CompareType, req.TimeRange, req.Benchmark)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "策略对比失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, comparison)
}

// GenerateAnalysisReport 生成分析报告
func (h *AnalysisHandler) GenerateAnalysisReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var req struct {
		ReportType  string      `json:"report_type" binding:"required"`
		DataSources []string    `json:"data_sources" binding:"required"`
		AnalysisIDs []uint      `json:"analysis_ids"`
		Template    string      `json:"template"`
		Parameters  interface{} `json:"parameters"`
		Format      string      `json:"format"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	taskID, err := h.reportService.GenerateAnalysisReport(userID.(uint), req.ReportType, req.DataSources, req.AnalysisIDs, req.Template, req.Parameters, req.Format)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "生成报告失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"task_id": taskID,
		"message": "报告生成任务已提交",
	})
}

// GetReportGenerationStatus 报告生成状态
func (h *AnalysisHandler) GetReportGenerationStatus(c *gin.Context) {
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

	status, err := h.reportService.GetReportGenerationStatus(userID.(uint), uint(taskID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取报告状态失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, status)
}

// GetSummaryStats 汇总统计
func (h *AnalysisHandler) GetSummaryStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	dataType := c.DefaultQuery("data_type", "all")
	timeRange := c.DefaultQuery("time_range", "30d")
	groupBy := c.DefaultQuery("group_by", "")

	// 解析 analysis_ids 参数
	analysisIDsParam := c.DefaultQuery("analysis_ids", "")
	var analysisIDs []uint
	if analysisIDsParam != "" {
		idStrs := strings.Split(analysisIDsParam, ",")
		for _, idStr := range idStrs {
			if id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32); err == nil {
				analysisIDs = append(analysisIDs, uint(id))
			}
		}
	}

	stats, err := h.analysisService.GetSummaryStats(userID.(uint), dataType, timeRange, groupBy, analysisIDs)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取汇总统计失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// MultiCompareResults 多结果对比
func (h *AnalysisHandler) MultiCompareResults(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var req struct {
		ResultIDs      []uint             `json:"result_ids" binding:"required,min=2"`
		ResultTypes    []string           `json:"result_types" binding:"required"`
		CompareMetrics []string           `json:"compare_metrics"`
		GroupBy        string             `json:"group_by"`
		Weights        map[string]float64 `json:"weights"`
		Benchmark      string             `json:"benchmark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	comparison, err := h.analysisService.MultiCompareResults(userID.(uint), req.ResultIDs, req.ResultTypes, req.CompareMetrics, req.GroupBy, req.Weights, req.Benchmark)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "多结果对比失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, comparison)
}
