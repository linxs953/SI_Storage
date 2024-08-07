package task

import (
	"context"

	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskListLogic {
	return &GetTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskListLogic) GetTaskList(req *types.GetTaskListDto) (resp *types.GetTaskListResp, err error) {
	resp = &types.GetTaskListResp{
		Code:    0,
		Message: "success",
		Data:    nil,
	}
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskList, err := taskMod.FindAllTask(context.Background())
	if err != nil {
		resp.Code = 1
		resp.Message = "获取任务列表错误"
		return
	}
	for _, task := range taskList {
		resp.Data = append(resp.Data, types.TaskInfo{
			TaskSpec: task,
			TaskName: task.TaskName,
			TaskId:   task.TaskID,
			TaskType: "autoapi",
		})
	}
	return
}
