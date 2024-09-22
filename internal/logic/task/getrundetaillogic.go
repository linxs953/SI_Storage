package task

import (
	"context"
	mgo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/task_run_log"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRunDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRunDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRunDetailLogic {
	return &GetRunDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRunDetailLogic) GetRunDetail(req *types.GetRunDetailDto) (resp *types.GetRunDetailResp, err error) {
	resp = &types.GetRunDetailResp{
		Code:     0,
		Message:  "success",
		TaskMeta: types.TaskMeta{},
		TaskRun:  types.TaskRecord{},
	}
	murl := mgo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskLogModel := task_run_log.NewTaskRunLogModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskRunLog")
	taskModel := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskInfo, err := taskModel.FindByTaskId(l.ctx, req.TaskId)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return
	}
	resp.TaskMeta = types.TaskMeta{
		TaskID:     taskInfo.TaskID,
		TaskName:   taskInfo.TaskName,
		Author:     taskInfo.Author,
		SceneCount: len(taskInfo.Scenes),
		CreateTime: taskInfo.CreateAt.Format("2006-01-02 15:04:05"),
		UpdateTime: taskInfo.UpdateAt.Format("2006-01-02 15:04:05"),
	}

	sceneRunList, err := getSceneDetailWithTaskd(l.ctx, taskLogModel, req.ExecId)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return
	}

	taskRunRecords, err := getAllTasks(l.ctx, taskLogModel, req.TaskId)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return
	}

	for _, record := range taskRunRecords {
		if record.ExecID == req.ExecId {
			state, err := strconv.Atoi(record.State)
			if err != nil {
				resp.Code = 1
				resp.Message = "状态转换错误"
				return resp, err
			}
			finishTime := ""
			if state == 0 {
				finishTime = ""
			} else {
				finishTime = record.UpdateTime
			}
			resp.TaskRun = types.TaskRecord{
				RunId:      record.ExecID,
				TaskId:     record.TaskId,
				State:      state,
				CreateTime: record.CreateTime,
				FinishTime: finishTime,
			}
			break
		}
	}

	if resp.TaskRun.RunId == "" {
		resp.Code = 1
		resp.Message = "未找到指定的运行记录"
		return
	}
	resp.TaskRun.SceneRecords = sceneRunList

	return
}

func getSceneDetailWithTaskd(ctx context.Context, taskLogModel task_run_log.TaskRunLogModel, execId string) ([]map[string]interface{}, error) {

	// 获取所有任务记录
	scenes, err := getSceneDetail(ctx, taskLogModel, execId)
	if err != nil {
		return nil, err
	}

	return scenes, nil
}
