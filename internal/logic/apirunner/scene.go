package apirunner

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (sc *SceneConfig) Execute(ctx context.Context, executor *ApiExecutor) {
	aec := ctx.Value("apirunner").(ApiExecutorContext)
	execID := aec.ExecID
	writeLogFunc := aec.WriteLog
	writeLogFunc("Scene", sc.SceneID, "Scene_Run_Action", "开始执行场景", nil)
	for _, action := range sc.Actions {
		msg := fmt.Sprintf("开始执行Action --- %s", action.ActionName)
		writeLogFunc("Scene", action.ActionID, "Scene_Run_Action", msg, nil)
		if err := action.TriggerAc(ctx); err != nil {
			// 执行Action出现错误
			writeLogFunc(
				"Scene",
				sc.SceneID,
				"Scene_Run_Action",
				"Action执行错误",
				err,
			)
		}
	}
	select {
	case <-time.After(time.Duration(sc.Timeout) * time.Second):
		msg := fmt.Sprintf("执行超时,sceneID=%s, execID=%s", sc.SceneID, execID)
		logx.Error(msg)
		writeLogFunc(
			"Scene",
			sc.SceneID,
			"Scene_Run_Action",
			msg,
			nil,
		)
		// 取消ctx，通知其他scene goroutine 退出
		ctx.Done()
		return
	case <-ctx.Done():
		// ctx被取消，即任务被取消，做一下清理动作
		writeLogFunc(
			"Scene",
			sc.SceneID,
			"Scene_Cancel_Action",
			"接收到上游取消信号",
			nil,
		)
		return
	default:
		{
			return
		}
	}

}
