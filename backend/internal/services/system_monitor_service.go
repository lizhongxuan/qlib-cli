package services

import (
	"fmt"
	"runtime"
	"time"

	"qlib-backend/internal/models"

	"gorm.io/gorm"
)

// SystemMonitorService 系统监控服务
type SystemMonitorService struct {
	db               *gorm.DB
	broadcastService *BroadcastService
}

// RealTimeMonitorData 实时监控数据
type RealTimeMonitorData struct {
	Timestamp    time.Time              `json:"timestamp"`
	SystemHealth *SystemHealthMetrics        `json:"system_health"`
	TaskMetrics  *TaskMetrics                `json:"task_metrics"`
	Performance  *SystemPerformanceMetrics  `json:"performance_metrics"`
	Alerts       []SystemAlert          `json:"alerts"`
	History      *MonitorHistory        `json:"history,omitempty"`
}

// SystemHealthMetrics 系统健康指标
type SystemHealthMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
	Status  string         `json:"status"` // healthy, warning, critical
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	Usage      float64 `json:"usage"`       // CPU使用率 (%)
	LoadAvg1   float64 `json:"load_avg_1"`  // 1分钟平均负载
	LoadAvg5   float64 `json:"load_avg_5"`  // 5分钟平均负载
	LoadAvg15  float64 `json:"load_avg_15"` // 15分钟平均负载
	CoreCount  int     `json:"core_count"`  // CPU核心数
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	TotalMB     uint64  `json:"total_mb"`     // 总内存 (MB)
	UsedMB      uint64  `json:"used_mb"`      // 已用内存 (MB)
	AvailableMB uint64  `json:"available_mb"` // 可用内存 (MB)
	Usage       float64 `json:"usage"`        // 内存使用率 (%)
	BuffersMB   uint64  `json:"buffers_mb"`   // 缓冲区 (MB)
	CachedMB    uint64  `json:"cached_mb"`    // 缓存 (MB)
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	TotalGB     uint64  `json:"total_gb"`     // 总容量 (GB)
	UsedGB      uint64  `json:"used_gb"`      // 已用容量 (GB)
	AvailableGB uint64  `json:"available_gb"` // 可用容量 (GB)
	Usage       float64 `json:"usage"`        // 磁盘使用率 (%)
	ReadIOPS    uint64  `json:"read_iops"`    // 读IOPS
	WriteIOPS   uint64  `json:"write_iops"`   // 写IOPS
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	BytesIn    uint64  `json:"bytes_in"`     // 入站字节数
	BytesOut   uint64  `json:"bytes_out"`    // 出站字节数
	PacketsIn  uint64  `json:"packets_in"`   // 入站包数
	PacketsOut uint64  `json:"packets_out"`  // 出站包数
	ErrorsIn   uint64  `json:"errors_in"`    // 入站错误数
	ErrorsOut  uint64  `json:"errors_out"`   // 出站错误数
	Bandwidth  float64 `json:"bandwidth"`    // 带宽使用率 (%)
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	TotalTasks     int `json:"total_tasks"`     // 总任务数
	RunningTasks   int `json:"running_tasks"`   // 运行中任务数
	QueuedTasks    int `json:"queued_tasks"`    // 排队任务数
	CompletedTasks int `json:"completed_tasks"` // 已完成任务数
	FailedTasks    int `json:"failed_tasks"`    // 失败任务数
	AverageWaitTime float64 `json:"average_wait_time"` // 平均等待时间(秒)
	AverageRunTime  float64 `json:"average_run_time"`  // 平均运行时间(秒)
}

// SystemPerformanceMetrics 系统性能指标
type SystemPerformanceMetrics struct {
	ThroughputQPS   float64 `json:"throughput_qps"`   // 请求处理速率 (QPS)
	ResponseTimeMs  float64 `json:"response_time_ms"` // 平均响应时间 (ms)
	ErrorRate       float64 `json:"error_rate"`       // 错误率 (%)
	ActiveSessions  int     `json:"active_sessions"`  // 活跃会话数
	ConcurrentUsers int     `json:"concurrent_users"` // 并发用户数
}

// SystemAlert 系统告警
type SystemAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // warning, error, critical
	Category    string    `json:"category"`    // cpu, memory, disk, network, task
	Message     string    `json:"message"`
	Threshold   float64   `json:"threshold"`   // 告警阈值
	CurrentValue float64  `json:"current_value"` // 当前值
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
}

// MonitorHistory 监控历史数据
type MonitorHistory struct {
	TimeRange   string                   `json:"time_range"`
	Interval    int                      `json:"interval"` // 数据间隔(秒)
	DataPoints  []HistoricalDataPoint         `json:"data_points"`
	Trends      map[string]SystemTrendAnalysis `json:"trends"`
}

// HistoricalDataPoint 历史数据点
type HistoricalDataPoint struct {
	Timestamp time.Time                `json:"timestamp"`
	Values    map[string]float64       `json:"values"`
}

// SystemTrendAnalysis 系统趋势分析
type SystemTrendAnalysis struct {
	Direction string  `json:"direction"` // up, down, stable
	Rate      float64 `json:"rate"`      // 变化率
	Prediction float64 `json:"prediction"` // 预测值
}

// NewSystemMonitorService 创建新的系统监控服务
func NewSystemMonitorService(db *gorm.DB, broadcastService *BroadcastService) *SystemMonitorService {
	return &SystemMonitorService{
		db:               db,
		broadcastService: broadcastService,
	}
}

// GetRealTimeData 获取实时监控数据
func (sms *SystemMonitorService) GetRealTimeData(userID uint, metrics []string, interval int, includeHistory bool) (*RealTimeMonitorData, error) {
	now := time.Now()
	
	data := &RealTimeMonitorData{
		Timestamp: now,
	}
	
	// 收集系统健康指标
	systemHealth, err := sms.collectSystemHealth(metrics)
	if err != nil {
		return nil, fmt.Errorf("收集系统健康指标失败: %v", err)
	}
	data.SystemHealth = systemHealth
	
	// 收集任务指标
	taskMetrics, err := sms.collectTaskMetrics(userID)
	if err != nil {
		return nil, fmt.Errorf("收集任务指标失败: %v", err)
	}
	data.TaskMetrics = taskMetrics
	
	// 收集性能指标
	performanceMetrics, err := sms.collectPerformanceMetrics()
	if err != nil {
		return nil, fmt.Errorf("收集性能指标失败: %v", err)
	}
	data.Performance = performanceMetrics
	
	// 检查并获取告警
	alerts, err := sms.checkSystemAlerts(systemHealth, taskMetrics, performanceMetrics)
	if err != nil {
		return nil, fmt.Errorf("检查系统告警失败: %v", err)
	}
	data.Alerts = alerts
	
	// 如果需要历史数据
	if includeHistory {
		history, err := sms.getMonitorHistory(metrics, interval, 60) // 最近60个数据点
		if err != nil {
			return nil, fmt.Errorf("获取监控历史失败: %v", err)
		}
		data.History = history
	}
	
	// 异步广播实时数据
	go func() {
		sms.broadcastRealTimeData(userID, data)
	}()
	
	return data, nil
}

// collectSystemHealth 收集系统健康指标
func (sms *SystemMonitorService) collectSystemHealth(metrics []string) (*SystemHealthMetrics, error) {
	health := &SystemHealthMetrics{}
	
	// 收集CPU指标
	if sms.shouldCollectMetric(metrics, "cpu") {
		health.CPU = sms.collectCPUMetrics()
	}
	
	// 收集内存指标
	if sms.shouldCollectMetric(metrics, "memory") {
		health.Memory = sms.collectMemoryMetrics()
	}
	
	// 收集磁盘指标
	if sms.shouldCollectMetric(metrics, "disk") {
		health.Disk = sms.collectDiskMetrics()
	}
	
	// 收集网络指标
	if sms.shouldCollectMetric(metrics, "network") {
		health.Network = sms.collectNetworkMetrics()
	}
	
	// 计算整体健康状态
	health.Status = sms.calculateHealthStatus(health)
	
	return health, nil
}

// collectCPUMetrics 收集CPU指标
func (sms *SystemMonitorService) collectCPUMetrics() CPUMetrics {
	// 简化的CPU指标收集，实际应该使用系统调用
	return CPUMetrics{
		Usage:     float64(runtime.NumGoroutine()) * 0.1, // 模拟CPU使用率
		LoadAvg1:  1.2,  // 模拟数据
		LoadAvg5:  1.5,  // 模拟数据
		LoadAvg15: 1.8,  // 模拟数据
		CoreCount: runtime.NumCPU(),
	}
}

// collectMemoryMetrics 收集内存指标
func (sms *SystemMonitorService) collectMemoryMetrics() MemoryMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 简化的内存指标计算
	totalMB := uint64(8192) // 假设8GB内存
	usedMB := uint64(m.Sys / 1024 / 1024)
	availableMB := totalMB - usedMB
	
	return MemoryMetrics{
		TotalMB:     totalMB,
		UsedMB:      usedMB,
		AvailableMB: availableMB,
		Usage:       float64(usedMB) / float64(totalMB) * 100,
		BuffersMB:   256, // 模拟数据
		CachedMB:    512, // 模拟数据
	}
}

// collectDiskMetrics 收集磁盘指标
func (sms *SystemMonitorService) collectDiskMetrics() DiskMetrics {
	// 简化的磁盘指标
	return DiskMetrics{
		TotalGB:     500,  // 500GB
		UsedGB:      200,  // 200GB已用
		AvailableGB: 300,  // 300GB可用
		Usage:       40.0, // 40%使用率
		ReadIOPS:    150,  // 模拟读IOPS
		WriteIOPS:   80,   // 模拟写IOPS
	}
}

// collectNetworkMetrics 收集网络指标
func (sms *SystemMonitorService) collectNetworkMetrics() NetworkMetrics {
	// 简化的网络指标
	return NetworkMetrics{
		BytesIn:    1024 * 1024 * 100, // 100MB入站
		BytesOut:   1024 * 1024 * 50,  // 50MB出站
		PacketsIn:  10000,             // 1万个入站包
		PacketsOut: 8000,              // 8千个出站包
		ErrorsIn:   5,                 // 5个入站错误
		ErrorsOut:  2,                 // 2个出站错误
		Bandwidth:  25.5,              // 25.5%带宽使用率
	}
}

// collectTaskMetrics 收集任务指标
func (sms *SystemMonitorService) collectTaskMetrics(userID uint) (*TaskMetrics, error) {
	metrics := &TaskMetrics{}
	
	// 统计各状态任务数量
	var totalTasks, runningTasks, queuedTasks, completedTasks, failedTasks int64
	
	// 总任务数
	if err := sms.db.Model(&models.Task{}).Where("user_id = ?", userID).Count(&totalTasks).Error; err != nil {
		return nil, err
	}
	
	// 运行中任务数
	if err := sms.db.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "running").Count(&runningTasks).Error; err != nil {
		return nil, err
	}
	
	// 排队任务数
	if err := sms.db.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "queued").Count(&queuedTasks).Error; err != nil {
		return nil, err
	}
	
	// 已完成任务数
	if err := sms.db.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completedTasks).Error; err != nil {
		return nil, err
	}
	
	// 失败任务数
	if err := sms.db.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "failed").Count(&failedTasks).Error; err != nil {
		return nil, err
	}
	
	metrics.TotalTasks = int(totalTasks)
	metrics.RunningTasks = int(runningTasks)
	metrics.QueuedTasks = int(queuedTasks)
	metrics.CompletedTasks = int(completedTasks)
	metrics.FailedTasks = int(failedTasks)
	
	// 计算平均等待时间和运行时间（简化实现）
	metrics.AverageWaitTime = 30.5  // 30.5秒
	metrics.AverageRunTime = 120.8  // 120.8秒
	
	return metrics, nil
}

// collectPerformanceMetrics 收集性能指标
func (sms *SystemMonitorService) collectPerformanceMetrics() (*SystemPerformanceMetrics, error) {
	// 简化的性能指标
	return &SystemPerformanceMetrics{
		ThroughputQPS:   85.2, // 85.2 QPS
		ResponseTimeMs:  145.8, // 145.8ms平均响应时间
		ErrorRate:       0.5,   // 0.5%错误率
		ActiveSessions:  156,   // 156个活跃会话
		ConcurrentUsers: 89,    // 89个并发用户
	}, nil
}

// checkSystemAlerts 检查系统告警
func (sms *SystemMonitorService) checkSystemAlerts(health *SystemHealthMetrics, tasks *TaskMetrics, perf *SystemPerformanceMetrics) ([]SystemAlert, error) {
	var alerts []SystemAlert
	now := time.Now()
	
	// CPU使用率告警
	if health.CPU.Usage > 80.0 {
		alerts = append(alerts, SystemAlert{
			ID:           fmt.Sprintf("cpu-high-%d", now.Unix()),
			Type:         "warning",
			Category:     "cpu",
			Message:      fmt.Sprintf("CPU使用率过高: %.1f%%", health.CPU.Usage),
			Threshold:    80.0,
			CurrentValue: health.CPU.Usage,
			Timestamp:    now,
			Resolved:     false,
		})
	}
	
	// 内存使用率告警
	if health.Memory.Usage > 85.0 {
		alerts = append(alerts, SystemAlert{
			ID:           fmt.Sprintf("memory-high-%d", now.Unix()),
			Type:         "warning",
			Category:     "memory",
			Message:      fmt.Sprintf("内存使用率过高: %.1f%%", health.Memory.Usage),
			Threshold:    85.0,
			CurrentValue: health.Memory.Usage,
			Timestamp:    now,
			Resolved:     false,
		})
	}
	
	// 磁盘使用率告警
	if health.Disk.Usage > 90.0 {
		alerts = append(alerts, SystemAlert{
			ID:           fmt.Sprintf("disk-high-%d", now.Unix()),
			Type:         "critical",
			Category:     "disk",
			Message:      fmt.Sprintf("磁盘使用率过高: %.1f%%", health.Disk.Usage),
			Threshold:    90.0,
			CurrentValue: health.Disk.Usage,
			Timestamp:    now,
			Resolved:     false,
		})
	}
	
	// 任务失败率告警
	if tasks.TotalTasks > 0 {
		failureRate := float64(tasks.FailedTasks) / float64(tasks.TotalTasks) * 100
		if failureRate > 10.0 {
			alerts = append(alerts, SystemAlert{
				ID:           fmt.Sprintf("task-failure-%d", now.Unix()),
				Type:         "error",
				Category:     "task",
				Message:      fmt.Sprintf("任务失败率过高: %.1f%%", failureRate),
				Threshold:    10.0,
				CurrentValue: failureRate,
				Timestamp:    now,
				Resolved:     false,
			})
		}
	}
	
	// 响应时间告警
	if perf.ResponseTimeMs > 1000.0 {
		alerts = append(alerts, SystemAlert{
			ID:           fmt.Sprintf("response-slow-%d", now.Unix()),
			Type:         "warning",
			Category:     "performance",
			Message:      fmt.Sprintf("响应时间过长: %.1fms", perf.ResponseTimeMs),
			Threshold:    1000.0,
			CurrentValue: perf.ResponseTimeMs,
			Timestamp:    now,
			Resolved:     false,
		})
	}
	
	return alerts, nil
}

// 辅助方法

// shouldCollectMetric 判断是否需要收集指定指标
func (sms *SystemMonitorService) shouldCollectMetric(metrics []string, metric string) bool {
	for _, m := range metrics {
		if m == metric {
			return true
		}
	}
	return false
}

// calculateHealthStatus 计算整体健康状态
func (sms *SystemMonitorService) calculateHealthStatus(health *SystemHealthMetrics) string {
	if health.CPU.Usage > 90 || health.Memory.Usage > 90 || health.Disk.Usage > 95 {
		return "critical"
	}
	if health.CPU.Usage > 70 || health.Memory.Usage > 80 || health.Disk.Usage > 85 {
		return "warning"
	}
	return "healthy"
}

// getMonitorHistory 获取监控历史数据
func (sms *SystemMonitorService) getMonitorHistory(metrics []string, interval, count int) (*MonitorHistory, error) {
	// 简化的历史数据生成
	now := time.Now()
	dataPoints := make([]HistoricalDataPoint, count)
	
	for i := 0; i < count; i++ {
		timestamp := now.Add(-time.Duration((count-i)*interval) * time.Second)
		values := make(map[string]float64)
		
		// 模拟历史数据点
		for _, metric := range metrics {
			switch metric {
			case "cpu":
				values["cpu_usage"] = 20 + float64(i%40) + float64(i)*0.5
			case "memory":
				values["memory_usage"] = 30 + float64(i%30) + float64(i)*0.3
			case "disk":
				values["disk_usage"] = 40 + float64(i%20) + float64(i)*0.1
			case "network":
				values["network_bandwidth"] = 10 + float64(i%50) + float64(i)*0.2
			}
		}
		
		dataPoints[i] = HistoricalDataPoint{
			Timestamp: timestamp,
			Values:    values,
		}
	}
	
	// 计算趋势分析
	trends := make(map[string]SystemTrendAnalysis)
	for _, metric := range metrics {
		trends[metric] = SystemTrendAnalysis{
			Direction:  "up",
			Rate:       0.5,
			Prediction: 75.0,
		}
	}
	
	return &MonitorHistory{
		TimeRange:  fmt.Sprintf("%ds", interval*count),
		Interval:   interval,
		DataPoints: dataPoints,
		Trends:     trends,
	}, nil
}

// broadcastRealTimeData 广播实时数据
func (sms *SystemMonitorService) broadcastRealTimeData(userID uint, data *RealTimeMonitorData) {
	if sms.broadcastService != nil {
		event := Event{
			Type:      "system_monitor",
			Category:  "system",
			Data:      map[string]interface{}{"data": data},
			UserID:    userID,
			Timestamp: time.Now(),
		}
		sms.broadcastService.Publish(event)
	}
}