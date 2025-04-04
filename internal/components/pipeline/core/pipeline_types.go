package core

import (
	"context"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled  TaskStatus = "canceled"
)

// Pipeline 定义pipeline接口
type PipelineRunner interface {
	Hook
	// Initialize 初始化pipeline
	Initialize(ctx context.Context) error

	// Validate 验证pipeline配置
	Validate(ctx context.Context) error

	// Execute 执行pipeline
	Execute(ctx context.Context, spec map[string]interface{}) (map[string]interface{}, error)

	// Cancel 取消pipeline执行
	Cancel(ctx context.Context) error

	// GetStatus 获取pipeline状态
	GetStatus(ctx context.Context) TaskStatus

	// GetProgress 获取执行进度
	GetProgress(ctx context.Context) (float64, error)

	// GetMetrics 获取执行指标
	GetMetrics(ctx context.Context) map[string]interface{}

	// Cleanup 清理资源
	Cleanup(ctx context.Context) error
}

// Hook 定义钩子接口
type Hook interface {
	// OnStart 任务开始时的钩子
	OnStart(ctx context.Context, taskID string, spec map[string]interface{}) error

	// OnSuccess 任务成功时的钩子
	OnSuccess(ctx context.Context, taskID string, result map[string]interface{}) error

	// OnFailure 任务失败时的钩子
	OnFailure(ctx context.Context, taskID string, err error) error

	// OnCancel 任务取消时的钩子
	OnCancel(ctx context.Context, taskID string) error

	// OnComplete 任务完成时的钩子（无论成功失败）
	OnComplete(ctx context.Context, taskID string, result map[string]interface{}) error
}
