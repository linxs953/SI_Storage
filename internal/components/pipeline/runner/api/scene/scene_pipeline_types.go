package scene

import (
	api "Storage/internal/logic/workflows/api/apirunner"
	"Storage/internal/logic/workflows/core"
	"Storage/internal/model/task"
	"context"
	"sync"
	"time"
)

type PipelineStatus string

const (
	StatusPending   PipelineStatus = "pending"
	StatusRunning   PipelineStatus = "running"
	StatusCompleted PipelineStatus = "completed"
	StatusFailed    PipelineStatus = "failed"
	StatusCancelled PipelineStatus = "cancelled"
)

// 定义ScenePipeline的类型和接口

type ScenePipeline struct {
	core.BasePipeline

	// 场景定义
	SceneDefinition *SceneDefinition `json:"scene_definition"`

	// 运行器
	SceneRunner ScenePipelineRunner `json:"runner"`

	// 执行统计
	Stats *SceneRunStats `json:"stats,omitempty"`

	// 上下文数据
	ContextData map[string]interface{} `json:"context_data,omitempty"`
}

type ScenePipelineRunner interface {
	core.PipelineRunner
	// 管理多个apipipeline的方法
	StartAllApiPipelines(ctx context.Context) error

	// 接收下游运行产生的指标
	ReceiveMetrics(ctx context.Context, metrics *api.ApiMetrics) error

	// 上报指标给ApiRuntimeManager
	ReportMetrics(ctx context.Context) error
}

type SceneDefinition struct {
	SceneID string `bson:"scene_id,omitempty" json:"scene_id,omitempty"`
	// 场景配置
	ApiPipelines []*api.ApiPipeline `json:"scenes"`
	Strategy     *SceneStrategy     `json:"strategy"`
	SharedMemory *SharedMemory      `json:"shared_memory"`
}

// RuntimeStats 记录执行统计信息
type SceneRunStats struct {
	TotalRequests    int                 `json:"total_requests"`     // 总请求数
	SuccessRequests  int                 `json:"success_requests"`   // 成功的请求数
	FailedRequests   int                 `json:"failed_requests"`    // 失败的请求数
	TotalDuration    int64               `json:"total_duration_ms"`  // 总执行时长（毫秒）
	AverageLatency   int64               `json:"average_latency_ms"` // 平均响应时间（毫秒）
	AssertionsPassed int                 `json:"assertions_passed"`  // 通过的断言数
	AssertionsFailed int                 `json:"assertions_failed"`  // 失败的断言数
	Status           PipelineStatus      `json:"status"`
	StartTime        *time.Time          `json:"start_time,omitempty"`
	FinishTime       *time.Time          `json:"finish_time,omitempty"`
	Error            *core.PipelineError `json:"error,omitempty"`
}

type SceneStrategy struct {
	Timeout *task.TimeoutSetting `bson:"timeout,omitempty" json:"timeout,omitempty"` // 超时配置
	Retry   *task.RetrySetting   `bson:"retry,omitempty" json:"retry,omitempty"`     // 重试配置
}

type SharedMemory struct {
	// memory 是一个并发安全的map，key是string，value是interface{}
	memory sync.Map
}

// Set 设置共享内存中的值
func (s *SharedMemory) Set(key string, value interface{}) {
	s.memory.Store(key, value)
}

// Get 获取共享内存中的值
func (s *SharedMemory) Get(key string) (interface{}, bool) {
	return s.memory.Load(key)
}

// Delete 删除共享内存中的值
func (s *SharedMemory) Delete(key string) {
	s.memory.Delete(key)
}

// Has 判断共享内存中是否存在某个key
func (s *SharedMemory) Has(key string) bool {
	_, ok := s.memory.Load(key)
	return ok
}

// Clear 清空共享内存
func (s *SharedMemory) Clear() {
	s.memory = sync.Map{}
}
