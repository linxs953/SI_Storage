package reporter

import (
	"context"
)

// MetricsReporter 指标上报接口
type MetricsReporter interface {
	// Name 获取上报器名称
	Name() string
	
	// Report 上报指标数据
	Report(ctx context.Context, metrics map[string]interface{}) error
	
	// Close 关闭上报器并释放资源
	Close() error
}

// BaseReporter 基础上报器，包含共享功能
type BaseReporter struct {
	name string
}

// Name 获取上报器名称
func (r *BaseReporter) Name() string {
	return r.name
}
