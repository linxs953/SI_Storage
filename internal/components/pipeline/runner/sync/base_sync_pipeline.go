package sync

import (
	"Storage/internal/components/pipeline/core"
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// BaseSyncPipeline 基础同步管道实现
type BaseSyncPipeline struct {
	*core.BasePipeline

	// 数据源
	source SyncSource

	// 数据目标
	target SyncTarget

	// 同步配置
	config SyncConfig

	// 同步运行器
	runner SyncRunner

	// 同步结果
	result *SyncResult

	// 处理项计数器
	processedCount int64

	// 成功项计数器
	successCount int64

	// 失败项计数器
	failedCount int64

	// 跳过项计数器
	skippedCount int64

	// 是否正在运行
	running bool

	// 是否已取消
	cancelled bool
}

// NewBaseSyncPipeline 创建新的基础同步管道
func NewBaseSyncPipeline(name, description string) *BaseSyncPipeline {
	basePipeline := core.NewBasePipeline(name, description)
	return &BaseSyncPipeline{
		BasePipeline: basePipeline,
		result: &SyncResult{
			Success: false,
			Metrics: SyncMetrics{},
		},
	}
}

// SetSource 设置数据源
func (p *BaseSyncPipeline) SetSource(source SyncSource) error {
	if p.running {
		return fmt.Errorf("无法在管道运行时设置数据源")
	}
	p.source = source
	return nil
}

// SetTarget 设置数据目标
func (p *BaseSyncPipeline) SetTarget(target SyncTarget) error {
	if p.running {
		return fmt.Errorf("无法在管道运行时设置数据目标")
	}
	p.target = target
	return nil
}

// Initialize 初始化管道
func (p *BaseSyncPipeline) Initialize(ctx context.Context) error {
	// 调用基础初始化
	if err := p.BasePipeline.Initialize(ctx); err != nil {
		return err
	}

	// 重置计数器
	atomic.StoreInt64(&p.processedCount, 0)
	atomic.StoreInt64(&p.successCount, 0)
	atomic.StoreInt64(&p.failedCount, 0)
	atomic.StoreInt64(&p.skippedCount, 0)

	p.running = false
	p.cancelled = false

	// 初始化结果
	p.result = &SyncResult{
		Success: false,
		Metrics: SyncMetrics{},
	}

	return nil
}

// Validate 验证配置
func (p *BaseSyncPipeline) Validate(ctx context.Context) error {
	// 检查数据源
	if p.source == nil {
		return fmt.Errorf("数据源未设置")
	}

	// 检查数据目标
	if p.target == nil {
		return fmt.Errorf("数据目标未设置")
	}

	return nil
}

// Execute 执行同步
func (p *BaseSyncPipeline) Execute(ctx context.Context, spec map[string]interface{}) (map[string]interface{}, error) {
	// 记录开始时间
	startTime := time.Now()
	p.result.Metrics.StartTime = startTime
	p.running = true
	p.Status = core.TaskStatusRunning

	// 转换配置
	if err := p.parseConfig(spec); err != nil {
		p.result.Error = &core.PipelineError{
			Message: fmt.Sprintf("解析配置失败: %v", err),
			Code:    "CONFIG_ERROR",
		}
		p.Status = core.TaskStatusFailed
		p.running = false
		return nil, err
	}

	// 连接数据源
	if err := p.source.Connect(ctx, p.config.SourceConfig); err != nil {
		p.result.Error = &core.PipelineError{
			Message: fmt.Sprintf("连接数据源失败: %v", err),
			Code:    "SOURCE_CONNECT_ERROR",
		}
		p.Status = core.TaskStatusFailed
		p.running = false
		return nil, err
	}

	// 连接数据目标
	if err := p.target.Connect(ctx, p.config.TargetConfig); err != nil {
		p.result.Error = &core.PipelineError{
			Message: fmt.Sprintf("连接数据目标失败: %v", err),
			Code:    "TARGET_CONNECT_ERROR",
		}
		p.Status = core.TaskStatusFailed
		p.running = false
		return nil, err
	}

	// 获取数据
	p.Progress = 0.3
	data, err := p.source.Fetch(ctx, p.config.Options)
	if err != nil {
		p.result.Error = &core.PipelineError{
			Message: fmt.Sprintf("获取数据失败: %v", err),
			Code:    "FETCH_ERROR",
		}
		p.Status = core.TaskStatusFailed
		p.running = false
		return nil, err
	}

	// 转换数据
	p.Progress = 0.6
	transformedData, err := p.source.Transform(ctx, data, p.config.TransformConfig)
	if err != nil {
		p.result.Error = &core.PipelineError{
			Message: fmt.Sprintf("转换数据失败: %v", err),
			Code:    "TRANSFORM_ERROR",
		}
		p.Status = core.TaskStatusFailed
		p.running = false
		return nil, err
	}

	// 写入数据
	p.Progress = 0.8
	if err := p.target.Write(ctx, transformedData, p.config.Options); err != nil {
		p.result.Error = &core.PipelineError{
			Message: fmt.Sprintf("写入数据失败: %v", err),
			Code:    "WRITE_ERROR",
		}
		p.Status = core.TaskStatusFailed
		p.running = false
		return nil, err
	}

	// 完成同步
	p.Progress = 1.0
	endTime := time.Now()
	p.result.Metrics.EndTime = endTime
	p.result.Metrics.Duration = endTime.Sub(startTime).Seconds()
	p.result.Metrics.ProcessedItems = atomic.LoadInt64(&p.processedCount)
	p.result.Metrics.SuccessItems = atomic.LoadInt64(&p.successCount)
	p.result.Metrics.FailedItems = atomic.LoadInt64(&p.failedCount)
	p.result.Metrics.SkippedItems = atomic.LoadInt64(&p.skippedCount)

	// 计算速率
	if p.result.Metrics.Duration > 0 {
		p.result.Metrics.ReadRate = float64(p.result.Metrics.ProcessedItems) / p.result.Metrics.Duration
		p.result.Metrics.WriteRate = float64(p.result.Metrics.SuccessItems) / p.result.Metrics.Duration
	}

	p.result.Success = p.result.Metrics.FailedItems == 0
	p.result.Data = transformedData

	// 设置结果
	resultMap := map[string]interface{}{
		"success":         p.result.Success,
		"processed_items": p.result.Metrics.ProcessedItems,
		"success_items":   p.result.Metrics.SuccessItems,
		"failed_items":    p.result.Metrics.FailedItems,
		"skipped_items":   p.result.Metrics.SkippedItems,
		"duration":        p.result.Metrics.Duration,
	}

	p.Status = core.TaskStatusCompleted
	p.Result = resultMap
	p.EndTime = endTime
	p.running = false

	return resultMap, nil
}

// parseConfig 解析配置
func (p *BaseSyncPipeline) parseConfig(spec map[string]interface{}) error {
	// 解析源配置
	if sourceConfig, ok := spec["source_config"].(map[string]interface{}); ok {
		p.config.SourceConfig = sourceConfig
	}

	// 解析目标配置
	if targetConfig, ok := spec["target_config"].(map[string]interface{}); ok {
		p.config.TargetConfig = targetConfig
	}

	// 解析选项
	if options, ok := spec["options"].(map[string]interface{}); ok {
		p.config.Options = options
	}

	// 解析转换配置
	if transformConfig, ok := spec["transform_config"].(map[string]interface{}); ok {
		p.config.TransformConfig = transformConfig
	}

	return nil
}

// Cancel 取消同步
func (p *BaseSyncPipeline) Cancel(ctx context.Context) error {
	if !p.running {
		return nil
	}

	p.cancelled = true
	p.Status = core.TaskStatusCanceled
	endTime := time.Now()
	p.EndTime = endTime

	// 关闭连接
	if p.source != nil {
		_ = p.source.Close(ctx)
	}

	if p.target != nil {
		_ = p.target.Close(ctx)
	}

	// 更新指标
	p.result.Metrics.EndTime = endTime
	p.result.Metrics.Duration = endTime.Sub(p.result.Metrics.StartTime).Seconds()
	p.result.Success = false
	p.result.Error = &core.PipelineError{
		Message: "同步任务已取消",
		Code:    "TASK_CANCELED",
	}

	p.running = false

	return p.BasePipeline.Cancel(ctx)
}

// GetSyncMetrics 获取同步指标
func (p *BaseSyncPipeline) GetSyncMetrics(ctx context.Context) (*SyncMetrics, error) {
	if p.result == nil {
		return nil, fmt.Errorf("同步尚未开始")
	}

	// 创建副本
	metrics := SyncMetrics{
		StartTime:      p.result.Metrics.StartTime,
		EndTime:        p.result.Metrics.EndTime,
		Duration:       p.result.Metrics.Duration,
		ProcessedItems: atomic.LoadInt64(&p.processedCount),
		SuccessItems:   atomic.LoadInt64(&p.successCount),
		FailedItems:    atomic.LoadInt64(&p.failedCount),
		SkippedItems:   atomic.LoadInt64(&p.skippedCount),
		ReadRate:       p.result.Metrics.ReadRate,
		WriteRate:      p.result.Metrics.WriteRate,
		Errors:         p.result.Metrics.Errors,
	}

	// 如果管道正在运行，更新持续时间
	if p.running && !p.result.Metrics.StartTime.IsZero() {
		metrics.Duration = time.Since(p.result.Metrics.StartTime).Seconds()
		if metrics.Duration > 0 {
			metrics.ReadRate = float64(metrics.ProcessedItems) / metrics.Duration
			metrics.WriteRate = float64(metrics.SuccessItems) / metrics.Duration
		}
	}

	return &metrics, nil
}

// GetResult 获取同步结果
func (p *BaseSyncPipeline) GetResult(ctx context.Context) (*SyncResult, error) {
	if p.result == nil {
		return nil, fmt.Errorf("同步尚未开始")
	}

	return p.result, nil
}

// GetProgress 获取进度
func (p *BaseSyncPipeline) GetProgress(ctx context.Context) (float64, error) {
	return p.Progress, nil
}

// Cleanup 清理资源
func (p *BaseSyncPipeline) Cleanup(ctx context.Context) error {
	// 关闭连接
	if p.source != nil {
		_ = p.source.Close(ctx)
	}

	if p.target != nil {
		_ = p.target.Close(ctx)
	}

	return p.BasePipeline.Cleanup(ctx)
}

// IncrementProcessed 增加处理项计数
func (p *BaseSyncPipeline) IncrementProcessed(count int64) {
	atomic.AddInt64(&p.processedCount, count)
}

// IncrementSuccess 增加成功项计数
func (p *BaseSyncPipeline) IncrementSuccess(count int64) {
	atomic.AddInt64(&p.successCount, count)
}

// IncrementFailed 增加失败项计数
func (p *BaseSyncPipeline) IncrementFailed(count int64) {
	atomic.AddInt64(&p.failedCount, count)
}

// IncrementSkipped 增加跳过项计数
func (p *BaseSyncPipeline) IncrementSkipped(count int64) {
	atomic.AddInt64(&p.skippedCount, count)
}

// GetMetrics 获取执行指标
func (p *BaseSyncPipeline) GetMetrics(ctx context.Context) map[string]interface{} {
	baseMetrics := p.BasePipeline.GetMetrics(ctx)

	// 添加同步特定指标
	syncMetrics, err := p.GetSyncMetrics(ctx)
	if err == nil && syncMetrics != nil {
		baseMetrics["processed_items"] = syncMetrics.ProcessedItems
		baseMetrics["success_items"] = syncMetrics.SuccessItems
		baseMetrics["failed_items"] = syncMetrics.FailedItems
		baseMetrics["skipped_items"] = syncMetrics.SkippedItems
		baseMetrics["read_rate"] = syncMetrics.ReadRate
		baseMetrics["write_rate"] = syncMetrics.WriteRate

		if !syncMetrics.StartTime.IsZero() {
			baseMetrics["sync_start_time"] = syncMetrics.StartTime.Format(time.RFC3339)
		}

		if !syncMetrics.EndTime.IsZero() {
			baseMetrics["sync_end_time"] = syncMetrics.EndTime.Format(time.RFC3339)
		}
	}

	return baseMetrics
}
