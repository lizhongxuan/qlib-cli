package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"qlib-backend/internal/models"
	"qlib-backend/internal/services"
	"qlib-backend/internal/utils"
)

type BacktestResultsHandler struct {
	backtestResultsService *services.BacktestResultsService
}

func NewBacktestResultsHandler(backtestResultsService *services.BacktestResultsService) *BacktestResultsHandler {
	return &BacktestResultsHandler{
		backtestResultsService: backtestResultsService,
	}
}

// GetDetailedResults 获取详细回测结果
func (h *BacktestResultsHandler) GetDetailedResults(c *gin.Context) {
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

	includeTradeDetails := c.DefaultQuery("include_trade_details", "false") == "true"
	includePositionDetails := c.DefaultQuery("include_position_details", "false") == "true"
	includeRiskMetrics := c.DefaultQuery("include_risk_metrics", "true") == "true"
	timeRange := c.DefaultQuery("time_range", "")

	// 构建查询选项
	options := services.GetDetailedResultsOptions{
		IncludeTradeDetails:    includeTradeDetails,
		IncludePositionDetails: includePositionDetails,
		IncludeRiskMetrics:     includeRiskMetrics,
		TimeRange:              timeRange,
	}

	results, err := h.backtestResultsService.GetDetailedResultsWithOptions(uint(resultID), userID.(uint), options)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取详细回测结果失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, results)
}

// GetChartData 获取图表数据
func (h *BacktestResultsHandler) GetChartData(c *gin.Context) {
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

	chartType := c.Param("chart_type")
	if chartType == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "图表类型不能为空")
		return
	}

	timeRange := c.DefaultQuery("time_range", "")
	resolution := c.DefaultQuery("resolution", "daily")
	benchmark := c.DefaultQuery("benchmark", "")
	indicators := c.QueryArray("indicators")

	// 构建图表查询选项
	options := services.GetChartDataOptions{
		TimeRange:  timeRange,
		Resolution: resolution,
		Benchmark:  benchmark,
		Indicators: indicators,
	}

	chartData, err := h.backtestResultsService.GetChartDataWithOptions(uint(resultID), services.ChartType(chartType), userID.(uint), options)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取图表数据失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, chartData)
}

// ExportBacktestReport 导出回测报告
func (h *BacktestResultsHandler) ExportBacktestReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var req struct {
		ResultIDs     []uint   `json:"result_ids" binding:"required"`
		ReportType    string   `json:"report_type" binding:"required"` // 报告类型: "summary", "detailed", "comparison"
		Format        string   `json:"format" binding:"required"`      // 格式: "pdf", "excel", "html"
		Template      string   `json:"template"`                       // 模板名称
		Sections      []string `json:"sections"`                       // 包含的部分
		IncludeCharts bool     `json:"include_charts"`                 // 是否包含图表
		Benchmark     string   `json:"benchmark"`                      // 基准指标
		Language      string   `json:"language"`                       // 语言: "zh", "en"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 支持多个结果ID的导出
	if req.ReportType == "comparison" && len(req.ResultIDs) < 2 {
		utils.ErrorResponse(c, http.StatusBadRequest, "对比报告至少需要2个回测结果")
		return
	}

	exportReq := models.BacktestReportExportRequestExtended{
		ResultIDs:     req.ResultIDs,
		ReportType:    req.ReportType,
		Format:        req.Format,
		Template:      req.Template,
		Sections:      req.Sections,
		IncludeCharts: req.IncludeCharts,
		Benchmark:     req.Benchmark,
		Language:      req.Language,
	}
	taskID, err := h.backtestResultsService.ExportBacktestReportExtended(exportReq, userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "导出回测报告失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"task_id": taskID,
		"message": "报告导出任务已提交",
	})
}
