package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"Storage/internal/logic/workflows/core/metrics/reporter"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	// 指标数据
	metrics map[string]interface{}

	// 互斥锁
	mu sync.RWMutex

	// 开始时间
	startTime time.Time

	// 最近更新时间
	lastUpdated time.Time
	
	// 指标上报器
	reporter reporter.MetricsReporter
}

// NewMetricsCollector 创建新的指标收集器
// 如果未提供上报器，则创建无上报能力的收集器
func NewMetricsCollector(options ...Option) *MetricsCollector {
	collector := &MetricsCollector{
		metrics:    make(map[string]interface{}),
		startTime:  time.Now(),
		lastUpdated: time.Now(),
	}
	
	// 应用所有选项
	for _, option := range options {
		option(collector)
	}
	
	return collector
}

// Option 是应用于MetricsCollector的函数选项
type Option func(*MetricsCollector)

// WithReporter 添加指标上报器
func WithReporter(r reporter.MetricsReporter) Option {
	return func(c *MetricsCollector) {
		c.reporter = r
	}
}

// Record 记录指标
func (c *MetricsCollector) Record(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[key] = value
	c.lastUpdated = time.Now()
}

// RecordDuration 记录持续时间
func (c *MetricsCollector) RecordDuration(key string, startTime time.Time) {
	duration := time.Since(startTime).Seconds()
	c.Record(key, duration)
}

// GetMetrics 获取所有指标
func (c *MetricsCollector) GetMetrics() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 复制指标，避免外部修改
	result := make(map[string]interface{}, len(c.metrics))
	for k, v := range c.metrics {
		result[k] = v
	}

	// 添加基础指标
	result["start_time"] = c.startTime.Format(time.RFC3339)
	result["last_updated"] = c.lastUpdated.Format(time.RFC3339)
	result["uptime"] = time.Since(c.startTime).Seconds()

	return result
}

// Reset 重置指标
func (c *MetricsCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make(map[string]interface{})
	c.startTime = time.Now()
	c.lastUpdated = time.Now()
}

// RecordStage 记录阶段执行时间
func (c *MetricsCollector) RecordStage(ctx context.Context, stageName string, fn func(context.Context) error) error {
	startTime := time.Now()
	
	// 记录阶段开始
	c.Record(stageName+"_start", startTime.Format(time.RFC3339))
	
	// 执行函数
	err := fn(ctx)
	
	// 记录阶段结束
	endTime := time.Now()
	c.Record(stageName+"_end", endTime.Format(time.RFC3339))
	c.Record(stageName+"_duration", endTime.Sub(startTime).Seconds())
	
	if err != nil {
		c.Record(stageName+"_error", err.Error())
	}
	
	return err
}



// Report 上报当前收集的所有指标
func (c *MetricsCollector) Report(ctx context.Context) error {
	if c.reporter == nil {
		return fmt.Errorf("未配置指标上报器")
	}
	
	metrics := c.GetMetrics()
	return c.reporter.Report(ctx, metrics)
}

// ReportAndReset 上报并重置指标
func (c *MetricsCollector) ReportAndReset(ctx context.Context) error {
	err := c.Report(ctx)
	if err != nil {
		return err
	}
	
	c.Reset()
	return nil
}

// Close 关闭收集器和上报器
func (c *MetricsCollector) Close() error {
	if c.reporter != nil {
		return c.reporter.Close()
	}
	return nil
}

// GetReporter 获取当前使用的上报器
func (c *MetricsCollector) GetReporter() reporter.MetricsReporter {
	return c.reporter
}

// SetReporter 设置指标上报器
func (c *MetricsCollector) SetReporter(r reporter.MetricsReporter) {
	c.reporter = r
}
