package api

// 定义ApiRuntime的类型和接口
import (
	"Storage/internal/logic/workflows/api/scene"
	"Storage/internal/logic/workflows/core"
	"Storage/internal/logic/workflows/notification"
	"Storage/internal/model/task"
	"time"
)

// ExecutionStatus represents the current state of pipeline execution
type PipelineStatus string

const (
	StatusPending   PipelineStatus = "pending"
	StatusRunning   PipelineStatus = "running"
	StatusCompleted PipelineStatus = "completed"
	StatusFailed    PipelineStatus = "failed"
	StatusCancelled PipelineStatus = "cancelled"
)

// ApiRuntimePipeline 实现API自动化任务的流水线
type ApiRuntimePipeline struct {
	// 基础信息
	PipelineID          string
	PipelineName        string
	PipelineDescription string

	// 场景配置
	Scenes   []*scene.ScenePipeline `json:"scenes"`
	Strategy task.TaskStrategy      `json:"strategy"`

	// 执行状态
	Status     PipelineStatus      `json:"status"`
	StartTime  *time.Time          `json:"start_time,omitempty"`
	FinishTime *time.Time          `json:"finish_time,omitempty"`
	Error      *core.PipelineError `json:"error,omitempty"`

	// 通知管理
	Notifications notification.NotificationManager `json:"notifications"`

	// 执行统计
	Stats *RuntimeStats `json:"stats,omitempty"`
}

// RuntimeStats 记录执行统计信息
type RuntimeStats struct {
	TotalScenes     int `json:"total_scenes"`     // 总场景数
	CompletedScenes int `json:"completed_scenes"` // 完成的场景数
	FailedScenes    int `json:"failed_scenes"`    // 失败的场景数

	TotalRequests    int   `json:"total_requests"`     // 总请求数
	SuccessRequests  int   `json:"success_requests"`   // 成功的请求数
	FailedRequests   int   `json:"failed_requests"`    // 失败的请求数
	TotalDuration    int64 `json:"total_duration_ms"`  // 总执行时长（毫秒）
	AverageLatency   int64 `json:"average_latency_ms"` // 平均响应时间（毫秒）
	AssertionsPassed int   `json:"assertions_passed"`  // 通过的断言数
	AssertionsFailed int   `json:"assertions_failed"`  // 失败的断言数
}
