package task

import (
	"context"
	"encoding/json"
	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"math"

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
	total, err := taskMod.GetTotalCount(context.Background())
	if err != nil {
		resp.Code = 1
		resp.Message = "获取任务列表错误"
		return
	}
	if total == 0 {
		resp.Data = []map[string]interface{}{}
		resp.TotalNum = 0
		resp.CurrentPage = 0
		return
	}
	taskList, err := taskMod.FindAllTask(context.Background(), req.PageNum, req.PageSize)
	if err != nil {
		resp.Code = 1
		resp.Message = "获取任务列表错误"
		return
	}

	taskBts, err := json.Marshal(taskList)
	if err != nil {
		resp.Code = 2
		resp.Message = "序列化任务列表失败"
		return
	}
	var taskListMap []map[string]interface{}
	err = json.Unmarshal(taskBts, &taskListMap)
	if err != nil {
		resp.Code = 2
		resp.Message = "映射任务列表失败"
		return
	}
	resp.Data = taskListMap
	resp.CurrentPage = req.PageNum
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	resp.TotalPage = totalPages
	resp.TotalNum = int(total)
	return
}
