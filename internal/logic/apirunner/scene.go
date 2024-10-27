package apirunner

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (sc *SceneConfig) collectLog(logChan chan RunFlowLog, execID string, trigger string, message string, rootErr error) {
	logChan <- RunFlowLog{
		LogType:     "SCENE",
		EventId:     sc.SceneID,
		EventName:   sc.SceneName,
		TriggerNode: trigger,
		Message:     message,
		RunId:       execID,
		RootErr:     rootErr,
	}
}

func (sc *SceneConfig) Execute(ctx context.Context, executor *ApiExecutor) {
	aec := ctx.Value("apirunner").(ApiExecutorContext)
	execID := aec.ExecID
	// executor.LogChan <- RunFlowLog{
	// 	EventId:     sc.SceneID,
	// 	EventName:   sc.SceneName,
	// 	LogType:     "SCENE",
	// 	TriggerNode: "Scene_Start",
	// 	Message:     "开始执行场景",
	// 	RunId:       execID,
	// 	RootErr:     nil,
	// }
	sc.collectLog(executor.LogChan, execID, "Scene_Start", "开始执行场景", nil)
	var sceneRunErr error
	for _, action := range sc.Actions {
		if err := action.TriggerAc(ctx); err != nil {
			logx.Error(err)
			sceneRunErr = err
			break
		}

	}

	// executor.LogChan <- RunFlowLog{
	// 	LogType:     "SCENE",
	// 	EventName:   sc.SceneName,
	// 	EventId:     sc.SceneID,
	// 	TriggerNode: "Scene_Finish",
	// 	Message:     "场景执行完成",
	// 	RunId:       execID,
	// 	RootErr:     sceneRunErr,
	// }
	sc.collectLog(executor.LogChan, execID, "Scene_Finish", "场景执行完成", sceneRunErr)

	select {
	case <-time.After(time.Duration(sc.Timeout) * time.Second):

		// executor.LogChan <- RunFlowLog{
		// 	EventId:   sc.SceneID,
		// 	EventName: sc.SceneName,
		// 	LogType:   "Scene_Run_Timeout",
		// 	Message:   fmt.Sprintf("执行超时,sceneID=%s, execID=%s", sc.SceneID, execID),
		// 	RunId:     execID,
		// 	RootErr:   nil,
		// }
		sc.collectLog(executor.LogChan, execID, "Scene_Run_Timeout", fmt.Sprintf("执行超时,sceneID=%s, execID=%s", sc.SceneID, execID), nil)

		// 取消ctx，通知其他scene goroutine 退出
		ctx.Done()
		return
	case <-ctx.Done():
		// ctx被取消，即任务被取消，做一下清理动作

		// executor.LogChan <- RunFlowLog{
		// 	EventId:   sc.SceneID,
		// 	EventName: sc.SceneName,
		// 	LogType:   "Scene_Cancel",
		// 	Message:   "接收到上游取消信号",
		// 	RunId:     execID,
		// 	RootErr:   nil,
		// }
		sc.collectLog(executor.LogChan, execID, "Scene_Cancel", "接收到上游取消信号", nil)
		return
	default:
		{
			return
		}
	}

}
