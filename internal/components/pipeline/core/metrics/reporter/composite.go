package reporter

import (
	"context"
	"fmt"
	"log"
)

// CompositeReporter 组合型指标上报器
type CompositeReporter struct {
	BaseReporter
	// 子上报器列表
	reporters []MetricsReporter
	// 错误处理策略: "fail_fast"或"continue"
	errorPolicy string
}

// NewCompositeReporter 创建新的组合上报器
func NewCompositeReporter(errorPolicy string) *CompositeReporter {
	if errorPolicy != "fail_fast" && errorPolicy != "continue" {
		errorPolicy = "continue" // 默认容错
	}
	
	return &CompositeReporter{
		BaseReporter: BaseReporter{name: "composite"},
		reporters:   make([]MetricsReporter, 0),
		errorPolicy: errorPolicy,
	}
}

// AddReporter 添加子上报器
func (r *CompositeReporter) AddReporter(reporter MetricsReporter) {
	r.reporters = append(r.reporters, reporter)
}

// Report 上报指标到所有子上报器
func (r *CompositeReporter) Report(ctx context.Context, metrics map[string]interface{}) error {
	var lastError error
	
	for _, reporter := range r.reporters {
		err := reporter.Report(ctx, metrics)
		if err != nil {
			lastError = fmt.Errorf("上报器 %s 失败: %w", reporter.Name(), err)
			
			// 如果策略是快速失败，则立即返回
			if r.errorPolicy == "fail_fast" {
				return lastError
			}
			
			// 否则记录错误并继续
			log.Printf("[指标] 上报器 %s 错误: %v", reporter.Name(), err)
		}
	}
	
	return lastError
}

// Close 关闭所有子上报器
func (r *CompositeReporter) Close() error {
	var lastError error
	
	for _, reporter := range r.reporters {
		if err := reporter.Close(); err != nil {
			lastError = err
			log.Printf("[指标] 关闭上报器 %s 错误: %v", reporter.Name(), err)
		}
	}
	
	return lastError
}
