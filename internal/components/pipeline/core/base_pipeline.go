package core

import (
	"context"
	"time"
)

// BasePipeline 基础管道实现
type BasePipeline struct {
	// 管道名称
	Name string

	// 管道描述
	Description string

	// 管道状态
	Status TaskStatus

	// 开始时间
	StartTime time.Time

	// 结束时间
	EndTime time.Time

	// 执行进度
	Progress float64

	// 结果数据
	Result map[string]interface{}

	// 错误信息
	Error error
}

// OnCancel implements PipelineRunner.
func (p *BasePipeline) OnCancel(ctx context.Context, taskID string) error {
	panic("unimplemented")
}

// OnComplete implements PipelineRunner.
func (p *BasePipeline) OnComplete(ctx context.Context, taskID string, result map[string]interface{}) error {
	panic("unimplemented")
}

// OnFailure implements PipelineRunner.
func (p *BasePipeline) OnFailure(ctx context.Context, taskID string, err error) error {
	panic("unimplemented")
}

// OnStart implements PipelineRunner.
func (p *BasePipeline) OnStart(ctx context.Context, taskID string, spec map[string]interface{}) error {
	panic("unimplemented")
}

// OnSuccess implements PipelineRunner.
func (p *BasePipeline) OnSuccess(ctx context.Context, taskID string, result map[string]interface{}) error {
	panic("unimplemented")
}

// NewBasePipeline 创建新的基础管道
func NewBasePipeline(name, description string) *BasePipeline {
	return &BasePipeline{
		Name:        name,
		Description: description,
		Status:      TaskStatusPending,
		Progress:    0.0,
		Result:      make(map[string]interface{}),
	}
}

// Initialize 初始化管道
func (p *BasePipeline) Initialize(ctx context.Context) error {
	p.Status = TaskStatusPending
	p.Progress = 0.0
	p.StartTime = time.Time{}
	p.EndTime = time.Time{}
	p.Result = make(map[string]interface{})
	p.Error = nil

	// 调用钩子方法
	// for _, hook := range p.Hooks {
	// 	if err := hook.OnStart(ctx, p.Name, nil); err != nil {
	// 		return err
	// 	}
	// } // Added comment for closing brace

	return nil
}

// Validate 验证管道配置
func (p *BasePipeline) Validate(ctx context.Context) error {
	// 基础实现无需验证
	return nil
}

// Execute 执行管道
func (p *BasePipeline) Execute(ctx context.Context, spec map[string]interface{}) (map[string]interface{}, error) {
	// 记录开始时间
	p.StartTime = time.Now()
	p.Status = TaskStatusRunning

	// 调用钩子方法
	// for _, hook := range p.Hooks {
	// 	if err := hook.OnStart(ctx, p.Name, spec); err != nil {
	// 		p.Error = err
	// 		p.Status = TaskStatusFailed
	// 		return nil, err
	// 	}
	// }

	// 基础实现不做实际工作，返回空结果
	p.Result = make(map[string]interface{})
	p.Status = TaskStatusCompleted
	// p.Progress = 1.0
	p.EndTime = time.Now()

	// 调用成功钩子
	// for _, hook := range p.Hooks {
	// 	if err := hook.OnSuccess(ctx, p.Name, p.Result); err != nil {
	// 		p.Error = err
	// 		// 不中断执行流程，但记录错误
	// 	}
	// }

	return p.Result, nil
}

// Cancel 取消管道执行
func (p *BasePipeline) Cancel(ctx context.Context) error {
	if p.Status == TaskStatusRunning {
		p.Status = TaskStatusCanceled
		p.EndTime = time.Now()

		// 调用取消钩子
		// for _, hook := range p.Hooks {
		// 	if err := hook.OnCancel(ctx, p.Name); err != nil {
		// 		// 记录错误但不中断
		// 		p.Error = err
		// 	}
		// }
	}

	return nil
}

// GetStatus 获取管道状态
func (p *BasePipeline) GetStatus(ctx context.Context) TaskStatus {
	return p.Status
}

// GetProgress 获取执行进度
func (p *BasePipeline) GetProgress(ctx context.Context) (float64, error) {
	return p.Progress, nil
}

// GetMetrics 获取执行指标
func (p *BasePipeline) GetMetrics(ctx context.Context) map[string]interface{} {
	metrics := make(map[string]interface{})
	metrics["name"] = p.Name
	metrics["status"] = p.Status

	if !p.StartTime.IsZero() {
		metrics["start_time"] = p.StartTime.Format(time.RFC3339)
	}

	if !p.EndTime.IsZero() {
		metrics["end_time"] = p.EndTime.Format(time.RFC3339)
		metrics["duration"] = p.EndTime.Sub(p.StartTime).Seconds()
	}

	metrics["progress"] = p.Progress

	if p.Error != nil {
		metrics["error"] = p.Error.Error()
	}

	return metrics
}

// Cleanup 清理资源
func (p *BasePipeline) Cleanup(ctx context.Context) error {
	// 调用完成钩子
	// for _, hook := range p.Hooks {
	// 	if err := hook.OnComplete(ctx, p.Name, p.Result); err != nil {
	// 		// 记录错误但不中断
	// 		p.Error = err
	// 	}
	// }

	return nil
}

// AddHook 添加钩子
func (p *BasePipeline) AddHook(hook Hook) {
	// p.Hooks = append(p.Hooks, hook)
}
