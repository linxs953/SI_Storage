package task

import (
	"context"

	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

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
		TaskId:   taskInfo.TaskID,
		TaskName: taskInfo.TaskName,
		TaskType: "autoapi",
		TaskSpec: taskInfo,
	}
	return
}
