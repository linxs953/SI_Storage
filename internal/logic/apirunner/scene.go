package apirunner

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

)

func (sc *SceneConfig) Execute(ctx context.Context) {
	errChan := make(chan error)
	aec := ctx.Value("apirunner").(ApiExecutorContext)
	execID := aec.ExecID
	for _, action := range sc.Actions {
		action.TriggerAc(ctx)
	}
	select {
	case <-time.After(time.Duration(sc.Timeout) * time.Second):
		logx.Errorf("执行超时,sceneID=%s, execID=%s", sc.SceneID, execID)
		return
	case <-ctx.Done():
		// ctx被取消，即任务被取消，做一下清理动作
		return
	case <-errChan:
		// 执行过程中出现错误,做一下处理
		return
	}
}
