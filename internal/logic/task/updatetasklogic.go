package task

import (
	"context"
	"encoding/json"
	"lexa-engine/internal/logic"
	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskLogic {
	return &UpdateTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskLogic) UpdateTask(req *types.UpdateTaskDto) (resp *types.UpdateTaskResp, err error) {
	resp = &types.UpdateTaskResp{
		Code:    0,
		Data:    types.TaskInfo{},
		Message: "success",
	}

	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskInfo, err := taskMod.FindByTaskId(context.Background(), req.TaskId)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return
	}
	if taskInfo == nil {
		resp.Code = 2
		resp.Message = "task not found"
		return
	}

	// todo: 重新设计入参dto， 更新逻辑再更新
	taskInfo.TaskName = req.TaskName
	taskInfo.Description = req.Description

	taskInfo.UpdateAt = time.Now()

	// 辅助函数用于转换 Actions
	convertActions := func(actions []map[string]interface{}) []logic.Action {
		result := make([]logic.Action, len(actions))
		bts, err := json.Marshal(actions)
		if err != nil {
			logx.Error(err)
			return result
		}
		err = json.Unmarshal(bts, &result)
		if err != nil {
			logx.Error(err)
			return result
		}
		return result
	}

	// 将 req.TaskSpec 转换为 []logic.Scene 类型
	scenes := make([]logic.Scene, len(req.TaskSpec))
	for i, scene := range req.TaskSpec {
		scenes[i] = logic.Scene{
			SceneName:   scene.SceneName,
			SceneId:     scene.SceneId,
			Description: scene.Description,
			Author:      scene.Author,
			Retry:       scene.Retry,
			Timeout:     scene.Timeout,
			SearchKey:   scene.SearchKey,
			EnvKey:      scene.EnvKey,
			Actions:     convertActions(scene.Actions),
		}
	}
	taskInfo.Scenes = scenes

	_, err = taskMod.Update(context.Background(), taskInfo)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return
	}

	taskInfo, err = taskMod.FindByTaskId(context.Background(), req.TaskId)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return
	}

	resp.Data = types.TaskInfo{
		TaskId:     taskInfo.TaskID,
		TaskName:   taskInfo.TaskName,
		Author:     taskInfo.Author,
		CreateTime: taskInfo.CreateAt.Format("2006-01-02 15:04:05"),
		UpdateTime: taskInfo.UpdateAt.Format("2006-01-02 15:04:05"),
		TaskType:   "autoapi",
		TaskSpec:   taskInfo.Scenes,
	}
	return
}
