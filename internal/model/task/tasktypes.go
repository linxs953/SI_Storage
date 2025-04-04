package task

import (
	"Storage/storage"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task 任务基础结构
type Task struct {
	// 任务基础信息
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	TaskId   string             `bson:"taskId" json:"taskId"`
	TaskName string             `bson:"taskName" json:"taskName"`
	TaskDesc string             `bson:"taskDesc" json:"taskDesc"`
	Type     int32              `bson:"type" json:"type"`
	Version  int64              `bson:"version" json:"version"`
	Enable   bool               `bson:"enable" json:"enable"`
	CreateAt time.Time          `bson:"createAt" json:"createAt"`
	UpdateAt time.Time          `bson:"updateAt" json:"updateAt"`

	// 任务配置 - 根据Type使用不同的配置结构
	APISpec  *APITaskSpec  `bson:"apiSpec,omitempty" json:"apiSpec,omitempty"`
	SyncSpec *SyncTaskSpec `bson:"syncSpec,omitempty" json:"syncSpec,omitempty"`
}

// APITaskSpec API测试任务配置
type APITaskSpec struct {
	Scenarios []ScenarioRef `bson:"scenarios,omitempty" json:"scenarios,omitempty"`
	Strategy  TaskStrategy  `bson:"strategy,omitempty" json:"strategy,omitempty"`
	// Enable    bool          `bson:"enable" json:"enable"`
	Version int64 `bson:"version,omitempty" json:"version,omitempty"`
}

// SyncTaskSpec 同步任务配置
type SyncTaskSpec struct {
	SyncType    string                     `bson:"sync_type,omitempty" json:"sync_type,omitempty"`     // 同步任务类型，11-apidoc/apifox，12-apidoc/swagger 2-db，3-other
	Source      []*storage.SyncSource      `bson:"source,omitempty" json:"source,omitempty"`           // 数据源
	Destination []*storage.SyncDestination `bson:"destination,omitempty" json:"destination,omitempty"` // 目标存储
	Strategy    *storage.Strategy          `bson:"strategy,omitempty" json:"strategy,omitempty"`       // 任务执行策略，定时/重试/超时
}

type ScenarioRef struct {
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

// 任务策略配置（嵌套结构）
type TaskStrategy struct {
	Timeout     *TimeoutSetting     `bson:"timeout,omitempty" json:"timeout,omitempty"`         // 超时配置
	Retry       *RetrySetting       `bson:"retry,omitempty" json:"retry,omitempty"`             // 重试配置
	AutoExecute *AutoExecuteSetting `bson:"autoExecute,omitempty" json:"autoExecute,omitempty"` // 自动执行配置
}

// 超时设置（示例：30秒超时）
type TimeoutSetting struct {
	Enabled  bool          `bson:"enabled" json:"enabled"`   // 是否启用超时检测
	Duration time.Duration `bson:"duration" json:"duration"` // 超时时长（单位：纳秒）
}

// 重试设置（示例：最多重试3次，间隔5秒）
type RetrySetting struct {
	Enabled     bool          `bson:"enabled" json:"enabled"`
	MaxAttempts int           `bson:"maxAttempts" json:"maxAttempts"` // 最大重试次数（0=不重试）
	Interval    time.Duration `bson:"interval" json:"interval"`       // 重试间隔
}

// 自动执行配置（示例：每天0点执行）
type AutoExecuteSetting struct {
	Enabled bool   `bson:"enabled" json:"enabled"`
	Cron    string `bson:"cron,omitempty" json:"cron,omitempty"` // cron表达式（如："0 0 * * *"）
}
