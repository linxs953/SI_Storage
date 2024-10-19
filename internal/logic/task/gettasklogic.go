package task

import (
	"context"
	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskLogic {
	return &GetTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskLogic) GetTask(req *types.GetTaskDto) (resp *types.GetTaskResp, err error) {
	resp = &types.GetTaskResp{
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
	resp.Data = types.TaskInfo{
		TaskId:      taskInfo.TaskID,
		TaskName:    taskInfo.TaskName,
		Author:      taskInfo.Author,
		TaskType:    "autoapi",
		Description: taskInfo.Description,
		TaskSpec:    taskInfo.Scenes,
		CreateTime:  taskInfo.CreateAt.Format("2006-01-02 15:04:05"),
		UpdateTime:  taskInfo.UpdateAt.Format("2006-01-02 15:04:05"),
	}
	return
}
