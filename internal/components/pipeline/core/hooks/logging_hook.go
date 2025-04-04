package hooks

import (
	"context"
	"log"
)

// LoggingHook 日志钩子实现
type LoggingHook struct {
	// 日志前缀
	Prefix string

	// 是否启用详细日志
	Verbose bool
}

// NewLoggingHook 创建新的日志钩子
func NewLoggingHook(prefix string, verbose bool) *LoggingHook {
	return &LoggingHook{
		Prefix:  prefix,
		Verbose: verbose,
	}
}

// OnStart 任务开始时的钩子
func (h *LoggingHook) OnStart(ctx context.Context, taskID string, spec map[string]interface{}) error {
	log.Printf("[%s] 任务开始: %s", h.Prefix, taskID)

	if h.Verbose {
		log.Printf("[%s] 任务详情: %v", h.Prefix, spec)
	}

	return nil
}

// OnSuccess 任务成功时的钩子
func (h *LoggingHook) OnSuccess(ctx context.Context, taskID string, result map[string]interface{}) error {
	log.Printf("[%s] 任务成功: %s", h.Prefix, taskID)

	if h.Verbose {
		log.Printf("[%s] 任务结果: %v", h.Prefix, result)
	}

	return nil
}

// OnFailure 任务失败时的钩子
func (h *LoggingHook) OnFailure(ctx context.Context, taskID string, err error) error {
	log.Printf("[%s] 任务失败: %s, 错误: %v", h.Prefix, taskID, err)
	return nil
}

// OnCancel 任务取消时的钩子
func (h *LoggingHook) OnCancel(ctx context.Context, taskID string) error {
	log.Printf("[%s] 任务取消: %s", h.Prefix, taskID)
	return nil
}

// OnComplete 任务完成时的钩子（无论成功失败）
func (h *LoggingHook) OnComplete(ctx context.Context, taskID string, result map[string]interface{}) error {
	log.Printf("[%s] 任务完成: %s", h.Prefix, taskID)
	return nil
}
