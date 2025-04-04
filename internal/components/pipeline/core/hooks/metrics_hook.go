package hooks

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"Storage/internal/logic/workflows/core/metrics/reporter"
)

// MetricsHook 指标收集钩子实现
type MetricsHook struct {
	// 指标数据锁
	mu sync.RWMutex

	// 任务ID映射到指标的映射
	metricsMap map[string]*TaskMetrics

	// 全局指标
	globalMetrics GlobalMetrics
	
	// 指标上报器
	reporter reporter.MetricsReporter
	
	// 是否在每个任务完成时上报
	reportOnComplete bool
	
	// 是否在任务失败时上报
	reportOnFailure bool
}

// TaskMetrics 单个任务的指标
type TaskMetrics struct {
	TaskID      string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Status      string // "success", "failure", "canceled"
	Error       string
	SpecData    map[string]interface{}
	ResultData  map[string]interface{}
	IsCompleted bool
}

// GlobalMetrics 全局指标统计
type GlobalMetrics struct {
	TotalTasks    int
	SuccessTasks  int
	FailureTasks  int
	CanceledTasks int
	TotalDuration time.Duration
	AvgDuration   time.Duration
}

// MetricsHookOption 配置选项函数类型
type MetricsHookOption func(*MetricsHook)

// WithReporter 设置指标上报器
func WithReporter(r reporter.MetricsReporter) MetricsHookOption {
	return func(h *MetricsHook) {
		h.reporter = r
	}
}

// WithReportOnComplete 设置是否在任务完成时上报
func WithReportOnComplete(enabled bool) MetricsHookOption {
	return func(h *MetricsHook) {
		h.reportOnComplete = enabled
	}
}

// WithReportOnFailure 设置是否在任务失败时上报
func WithReportOnFailure(enabled bool) MetricsHookOption {
	return func(h *MetricsHook) {
		h.reportOnFailure = enabled
	}
}

// NewMetricsHook 创建新的指标收集钩子
func NewMetricsHook(options ...MetricsHookOption) *MetricsHook {
	h := &MetricsHook{
		metricsMap:    make(map[string]*TaskMetrics),
		globalMetrics: GlobalMetrics{},
		reportOnComplete: false,
		reportOnFailure: true,
	}
	
	// 应用选项
	for _, option := range options {
		option(h)
	}
	
	return h
}

// OnStart 任务开始时的钩子
func (h *MetricsHook) OnStart(ctx context.Context, taskID string, spec map[string]interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 复制spec数据，避免后续被修改导致问题
	specCopy := make(map[string]interface{})
	for k, v := range spec {
		specCopy[k] = v
	}

	// 创建任务指标
	h.metricsMap[taskID] = &TaskMetrics{
		TaskID:      taskID,
		StartTime:   time.Now(),
		Status:      "running",
		SpecData:    specCopy,
		IsCompleted: false,
	}

	h.globalMetrics.TotalTasks++

	return nil
}

// OnSuccess 任务成功时的钩子
func (h *MetricsHook) OnSuccess(ctx context.Context, taskID string, result map[string]interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	metrics, exists := h.metricsMap[taskID]
	if !exists {
		// 如果不存在，创建一个新的
		metrics = &TaskMetrics{
			TaskID:    taskID,
			StartTime: time.Now(), // 近似值
		}
		h.metricsMap[taskID] = metrics
	}

	// 更新指标
	now := time.Now()
	metrics.EndTime = now
	metrics.Duration = now.Sub(metrics.StartTime)
	metrics.Status = "success"
	
	// 复制结果数据
	resultCopy := make(map[string]interface{})
	for k, v := range result {
		resultCopy[k] = v
	}
	metrics.ResultData = resultCopy

	// 更新全局指标
	h.globalMetrics.SuccessTasks++
	h.globalMetrics.TotalDuration += metrics.Duration
	h.globalMetrics.AvgDuration = h.globalMetrics.TotalDuration / time.Duration(h.globalMetrics.SuccessTasks+h.globalMetrics.FailureTasks)

	return nil
}

// OnFailure 任务失败时的钩子
func (h *MetricsHook) OnFailure(ctx context.Context, taskID string, err error) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	metrics, exists := h.metricsMap[taskID]
	if !exists {
		// 如果不存在，创建一个新的
		metrics = &TaskMetrics{
			TaskID:    taskID,
			StartTime: time.Now(), // 近似值
		}
		h.metricsMap[taskID] = metrics
	}

	// 更新指标
	now := time.Now()
	metrics.EndTime = now
	metrics.Duration = now.Sub(metrics.StartTime)
	metrics.Status = "failure"
	if err != nil {
		metrics.Error = err.Error()
	}

	// 更新全局指标
	h.globalMetrics.FailureTasks++
	h.globalMetrics.TotalDuration += metrics.Duration
	h.globalMetrics.AvgDuration = h.globalMetrics.TotalDuration / time.Duration(h.globalMetrics.SuccessTasks+h.globalMetrics.FailureTasks)
	
	// 如果配置了在失败时上报且有上报器
	if h.reportOnFailure && h.reporter != nil {
		metricsData := h.GetMetricsForExport()
		if err := h.reporter.Report(ctx, metricsData); err != nil {
			// 只记录错误，不影响正常流程
			return nil
		}
	}

	return nil
}

// OnCancel 任务取消时的钩子
func (h *MetricsHook) OnCancel(ctx context.Context, taskID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	metrics, exists := h.metricsMap[taskID]
	if !exists {
		// 如果不存在，创建一个新的
		metrics = &TaskMetrics{
			TaskID:    taskID,
			StartTime: time.Now(), // 近似值
		}
		h.metricsMap[taskID] = metrics
	}

	// 更新指标
	now := time.Now()
	metrics.EndTime = now
	metrics.Duration = now.Sub(metrics.StartTime)
	metrics.Status = "canceled"

	// 更新全局指标
	h.globalMetrics.CanceledTasks++

	return nil
}

// OnComplete 任务完成时的钩子（无论成功失败）
func (h *MetricsHook) OnComplete(ctx context.Context, taskID string, result map[string]interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	metrics, exists := h.metricsMap[taskID]
	if exists {
		metrics.IsCompleted = true
		
		// 如果配置了在完成时上报且有上报器
		if h.reportOnComplete && h.reporter != nil {
			metricsData := h.GetMetricsForExport()
			if err := h.reporter.Report(ctx, metricsData); err != nil {
				// 只记录错误，不影响正常流程
				return nil
			}
		}
	}

	return nil
}

// GetTaskMetrics 获取指定任务的指标
func (h *MetricsHook) GetTaskMetrics(taskID string) *TaskMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if metrics, exists := h.metricsMap[taskID]; exists {
		return metrics
	}
	return nil
}

// GetAllTaskMetrics 获取所有任务的指标
func (h *MetricsHook) GetAllTaskMetrics() map[string]*TaskMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// 创建副本以避免并发问题
	result := make(map[string]*TaskMetrics, len(h.metricsMap))
	for k, v := range h.metricsMap {
		result[k] = v
	}
	
	return result
}

// GetGlobalMetrics 获取全局指标
func (h *MetricsHook) GetGlobalMetrics() GlobalMetrics {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return h.globalMetrics
}

// Reset 重置所有指标
func (h *MetricsHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.metricsMap = make(map[string]*TaskMetrics)
	h.globalMetrics = GlobalMetrics{}
}

// ReportMetrics 手动上报当前收集的指标
func (h *MetricsHook) ReportMetrics(ctx context.Context) error {
	if h.reporter == nil {
		return fmt.Errorf("未配置指标上报器")
	}
	
	metricsData := h.GetMetricsForExport()
	return h.reporter.Report(ctx, metricsData)
}

// SetReporter 设置指标上报器
func (h *MetricsHook) SetReporter(r reporter.MetricsReporter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.reporter = r
}

// GetReporter 获取当前使用的上报器
func (h *MetricsHook) GetReporter() reporter.MetricsReporter {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return h.reporter
}

// SetReportOnComplete 设置是否在任务完成时自动上报
func (h *MetricsHook) SetReportOnComplete(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.reportOnComplete = enabled
}

// SetReportOnFailure 设置是否在任务失败时自动上报
func (h *MetricsHook) SetReportOnFailure(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.reportOnFailure = enabled
}

// GetMetricsForExport 获取用于导出的指标数据
func (h *MetricsHook) GetMetricsForExport() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// 准备用于导出的数据结构
	export := map[string]interface{}{
		"global_metrics": map[string]interface{}{
			"total_tasks":     h.globalMetrics.TotalTasks,
			"success_tasks":   h.globalMetrics.SuccessTasks,
			"failure_tasks":   h.globalMetrics.FailureTasks,
			"canceled_tasks":  h.globalMetrics.CanceledTasks,
			"total_duration":  h.globalMetrics.TotalDuration.String(),
			"avg_duration":    h.globalMetrics.AvgDuration.String(),
			"success_rate":    float64(0),
		},
		"tasks": make(map[string]interface{}),
	}
	
	// 计算成功率
	if h.globalMetrics.TotalTasks > 0 {
		export["global_metrics"].(map[string]interface{})["success_rate"] = 
			float64(h.globalMetrics.SuccessTasks) / float64(h.globalMetrics.TotalTasks)
	}
	
	// 添加每个任务的指标
	for taskID, metrics := range h.metricsMap {
		taskMetrics := map[string]interface{}{
			"task_id":      metrics.TaskID,
			"start_time":   metrics.StartTime.Format(time.RFC3339),
			"status":       metrics.Status,
			"is_completed": metrics.IsCompleted,
		}
		
		if !metrics.EndTime.IsZero() {
			taskMetrics["end_time"] = metrics.EndTime.Format(time.RFC3339)
			taskMetrics["duration"] = metrics.Duration.String()
		}
		
		if metrics.Error != "" {
			taskMetrics["error"] = metrics.Error
		}
		
		export["tasks"].(map[string]interface{})[taskID] = taskMetrics
	}
	
	return export
}
