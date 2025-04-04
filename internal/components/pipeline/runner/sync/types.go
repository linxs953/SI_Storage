package sync

import (
	base "Storage/internal/components/pipeline/core"
	"context"
	"time"
)

// SyncSource 同步数据源接口
type SyncSource interface {
	// 连接到数据源
	Connect(ctx context.Context, config map[string]interface{}) error

	// 获取数据
	Fetch(ctx context.Context, options map[string]interface{}) (interface{}, error)

	// 转换数据
	Transform(ctx context.Context, data interface{}, options map[string]interface{}) (interface{}, error)

	// 获取数据源状态
	GetStatus(ctx context.Context) SourceStatus

	// 关闭连接
	Close(ctx context.Context) error
}

// SyncTarget 同步数据目标接口
type SyncTarget interface {
	// 连接到目标
	Connect(ctx context.Context, config map[string]interface{}) error

	// 写入数据
	Write(ctx context.Context, data interface{}, options map[string]interface{}) error

	// 获取目标状态
	GetStatus(ctx context.Context) TargetStatus

	// 关闭连接
	Close(ctx context.Context) error
}

// SyncConfig 同步配置
type SyncConfig struct {
	// 源配置
	SourceConfig map[string]interface{} `json:"source_config"`

	// 目标配置
	TargetConfig map[string]interface{} `json:"target_config"`

	// 同步选项
	Options map[string]interface{} `json:"options"`

	// 转换配置
	TransformConfig map[string]interface{} `json:"transform_config"`
}

// SourceStatus 数据源状态
type SourceStatus struct {
	// 连接状态
	Connected bool `json:"connected"`

	// 上次同步时间
	LastSync time.Time `json:"last_sync"`

	// 数据项数量
	ItemCount int64 `json:"item_count"`

	// 错误信息
	Error string `json:"error,omitempty"`
}

// TargetStatus 数据目标状态
type TargetStatus struct {
	// 连接状态
	Connected bool `json:"connected"`

	// 上次写入时间
	LastWrite time.Time `json:"last_write"`

	// 写入项数量
	WrittenCount int64 `json:"written_count"`

	// 错误信息
	Error string `json:"error,omitempty"`
}

// SyncMetrics 同步指标
type SyncMetrics struct {
	// 开始时间
	StartTime time.Time `json:"start_time"`

	// 结束时间
	EndTime time.Time `json:"end_time"`

	// 持续时间（秒）
	Duration float64 `json:"duration"`

	// 处理的项数
	ProcessedItems int64 `json:"processed_items"`

	// 成功的项数
	SuccessItems int64 `json:"success_items"`

	// 失败的项数
	FailedItems int64 `json:"failed_items"`

	// 跳过的项数
	SkippedItems int64 `json:"skipped_items"`

	// 源读取速率（项/秒）
	ReadRate float64 `json:"read_rate"`

	// 目标写入速率（项/秒）
	WriteRate float64 `json:"write_rate"`

	// 错误信息
	Errors []string `json:"errors,omitempty"`
}

// SyncResult 同步结果
type SyncResult struct {
	// 同步成功状态
	Success bool `json:"success"`

	// 同步指标
	Metrics SyncMetrics `json:"metrics"`

	// 详细数据
	Data interface{} `json:"data,omitempty"`

	// 错误信息
	Error *base.PipelineError `json:"error,omitempty"`
}

// SyncPipeline 同步管道接口
type SyncPipeline interface {
	base.PipelineRunner

	// 设置数据源
	SetSource(source SyncSource) error

	// 设置数据目标
	SetTarget(target SyncTarget) error

	// 获取同步结果
	GetResult(ctx context.Context) (*SyncResult, error)

	// 获取同步指标
	GetSyncMetrics(ctx context.Context) (*SyncMetrics, error)
}

type SyncRunner interface {
}
