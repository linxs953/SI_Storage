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
	executor.LogChan <- RunFlowLog{
		EventId:     sc.SceneID,
		EventName:   sc.SceneName,
		LogType:     "SCENE",
		TriggerNode: "Scene_Start",
		Message:     "开始执行场景",
		RunId:       execID,
		RootErr:     nil,
	}
	var sceneRunErr error
	for _, action := range sc.Actions {
		// executor.LogChan <- RunFlowLog{
		// 	EventId:      action.SceneID,
		// 	LogType:      "SCENE",
		// 	TriggerNode:  "Scene_Action_Start",
		// 	Message:      fmt.Sprintf("开始执行Action %s --- %s", action.ActionName, action.ActionID),
		// 	RunId:        execID,
		// 	FinishCount:  0,
		// 	SuccessCount: 0,
		// 	FailCount:    0,
		// 	RootErr:      nil,
		// }
		if err := action.TriggerAc(ctx); err != nil {
			logx.Error(err)
			sceneRunErr = err
			break
			// 执行Action出现错误

			// executor.LogChan <- RunFlowLog{
			// 	LogType:     "SCENE",
			// 	EventId:     action.SceneID,
			// 	TriggerNode: "Scene_Action_Failed",
			// 	Message:     fmt.Sprintf("Action执行失败 %s --- %s", action.ActionName, action.ActionID),
			// 	RunId:       execID,
			// 	RootErr:     err,
			// 	FinishCount: adx + 1,
			// }
		}
		// executor.LogChan <- RunFlowLog{
		// 	EventId: action.SceneID,
		// 	LogType: "Scene_Action_Finish",
		// 	Message: fmt.Sprintf("Action执行完成 %s --- %s", action.ActionName, action.ActionID),
		// 	RunId:   execID,
		// 	RootErr: nil,
		// }
	}

	executor.LogChan <- RunFlowLog{
		LogType:     "SCENE",
		EventName:   sc.SceneName,
		EventId:     sc.SceneID,
		TriggerNode: "Scene_Finish",
		Message:     "场景执行完成",
		RunId:       execID,
		RootErr:     sceneRunErr,
	}

	select {
	case <-time.After(time.Duration(sc.Timeout) * time.Second):

		executor.LogChan <- RunFlowLog{
			EventId:   sc.SceneID,
			EventName: sc.SceneName,
			LogType:   "Scene_Run_Timeout",
			Message:   fmt.Sprintf("执行超时,sceneID=%s, execID=%s", sc.SceneID, execID),
			RunId:     execID,
			RootErr:   nil,
		}
		// 取消ctx，通知其他scene goroutine 退出
		ctx.Done()
		return
	case <-ctx.Done():
		// ctx被取消，即任务被取消，做一下清理动作

		executor.LogChan <- RunFlowLog{
			EventId:   sc.SceneID,
			EventName: sc.SceneName,
			LogType:   "Scene_Cancel",
			Message:   "接收到上游取消信号",
			RunId:     execID,
			RootErr:   nil,
		}
		return
	default:
		{
			return
		}
	}

}
