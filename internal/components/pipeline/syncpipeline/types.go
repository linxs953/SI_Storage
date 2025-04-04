package syncpipeline

import (
	"context"
	"time"

	"Storage/internal/components/pipeline/core"
)

// SyncTaskType 同步任务类型
type SyncTaskType string

const (
	// SyncTaskTypeAPIDoc API文档同步
	SyncTaskTypeAPIDoc SyncTaskType = "apidoc"
	// SyncTaskTypeDB 数据库同步
	SyncTaskTypeDB SyncTaskType = "db"
	// SyncTaskTypeConfig 配置同步
	SyncTaskTypeConfig SyncTaskType = "config"
)

// SyncSource 同步数据源接口
type SyncSource interface {
	// Connect 连接到数据源
	Connect(ctx context.Context, config map[string]interface{}) error

	// Fetch 获取数据
	Fetch(ctx context.Context, options map[string]interface{}) (interface{}, error)

	// Transform 转换数据
	Transform(ctx context.Context, data interface{}, options map[string]interface{}) (interface{}, error)

	// GetStatus 获取数据源状态
	GetStatus(ctx context.Context) *SourceStatus

	// Close 关闭连接
	Close(ctx context.Context) error
}

// SyncTarget 同步数据目标接口
type SyncTarget interface {
	// Connect 连接到目标
	Connect(ctx context.Context, config map[string]interface{}) error

	// Write 写入数据
	Write(ctx context.Context, data interface{}, options map[string]interface{}) error

	// GetStatus 获取目标状态
	GetStatus(ctx context.Context) *TargetStatus

	// Close 关闭连接
	Close(ctx context.Context) error
}

// SyncTask 同步任务接口
type SyncTask interface {
	// Type 返回同步任务类型
	Type() SyncTaskType

	// Configure 配置同步任务
	Configure(ctx context.Context, config *SyncConfig) error

	// GetSource 获取数据源
	GetSource() SyncSource

	// GetTarget 获取数据目标
	GetTarget() SyncTarget

	// PreSync 同步前处理
	PreSync(ctx context.Context) error

	// PostSync 同步后处理
	PostSync(ctx context.Context) error
}

// SyncConfig 同步配置
type SyncConfig struct {
	// TaskType 任务类型
	TaskType SyncTaskType `json:"task_type"`

	// SourceConfig 源配置
	SourceConfig map[string]interface{} `json:"source_config"`

	// TargetConfig 目标配置
	TargetConfig map[string]interface{} `json:"target_config"`

	// TransformConfig 转换配置
	TransformConfig map[string]interface{} `json:"transform_config,omitempty"`

	// Options 同步选项
	Options map[string]interface{} `json:"options,omitempty"`
}

// SourceStatus 数据源状态
type SourceStatus struct {
	// Connected 是否已连接
	Connected bool `json:"connected"`

	// LastSync 最后同步时间
	LastSync time.Time `json:"last_sync"`

	// ItemCount 数据项数量
	ItemCount int64 `json:"item_count"`

	// Error 错误信息
	Error string `json:"error,omitempty"`
}

// TargetStatus 数据目标状态
type TargetStatus struct {
	// Connected 是否已连接
	Connected bool `json:"connected"`

	// LastWrite 最后写入时间
	LastWrite time.Time `json:"last_write"`

	// WrittenCount 已写入数量
	WrittenCount int64 `json:"written_count"`

	// Error 错误信息
	Error string `json:"error,omitempty"`
}

// SyncMetrics 同步指标
type SyncMetrics struct {
	// StartTime 开始时间
	StartTime time.Time `json:"start_time"`

	// EndTime 结束时间
	EndTime time.Time `json:"end_time"`

	// Duration 持续时间(秒)
	Duration float64 `json:"duration"`

	// ProcessedItems 处理项数量
	ProcessedItems int64 `json:"processed_items"`

	// SuccessItems 成功项数量
	SuccessItems int64 `json:"success_items"`

	// FailedItems 失败项数量
	FailedItems int64 `json:"failed_items"`

	// SkippedItems 跳过项数量
	SkippedItems int64 `json:"skipped_items"`

	// ReadRate 读取速率(items/s)
	ReadRate float64 `json:"read_rate"`

	// WriteRate 写入速率(items/s)
	WriteRate float64 `json:"write_rate"`

	// Errors 错误列表
	Errors []string `json:"errors,omitempty"`
}

// SyncResult 同步结果
type SyncResult struct {
	// Success 是否成功
	Success bool `json:"success"`

	// Metrics 同步指标
	Metrics *SyncMetrics `json:"metrics"`

	// Data 同步数据
	Data interface{} `json:"data,omitempty"`

	// Error 错误信息
	Error *core.PipelineError `json:"error,omitempty"`
}
