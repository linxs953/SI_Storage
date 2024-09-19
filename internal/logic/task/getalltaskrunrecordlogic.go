package task

import (
	"context"
	mgo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/task_run_log"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllTaskRunRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllTaskRunRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllTaskRunRecordLogic {
	return &GetAllTaskRunRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllTaskRunRecordLogic) GetAllTaskRunRecord(req *types.GetAllTaskRunRecordDto) (resp *types.GetAllTaskRunRecordResp, err error) {
	resp = &types.GetAllTaskRunRecordResp{
		Code:       0,
		Message:    "success",
		TaskName:   "",
		TaskID:     "",
		Author:     "",
		SceneCount: 0,
		CreateTime: "",
		UpdateTime: "",
		Data:       make(map[string][]map[string]interface{}),
	}
	murl := mgo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskLogModel := task_run_log.NewTaskRunLogModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskRunLog")
	taskModel := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskInfo, err := taskModel.FindByTaskId(l.ctx, req.TaskId)
	if err != nil {
		return nil, err
	}
	resp.TaskName = taskInfo.TaskName
	resp.TaskID = taskInfo.TaskID
	resp.Author = taskInfo.Author
	resp.SceneCount = len(taskInfo.Scenes)
	resp.CreateTime = taskInfo.CreateAt.Format("2006-01-02 15:04:05")
	resp.UpdateTime = taskInfo.UpdateAt.Format("2006-01-02 15:04:05")
	taskRunRecord, err := taskLogModel.FindTaskRunRecord(l.ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	// 根据 record 的 sceneDetail 来判断 task 的状态
	// 使用map来聚合相同execID的记录
	execMap := make(map[string][]map[string]interface{})

	for _, r := range taskRunRecord {
		execID := r.ExecID
		logx.Error(r)
		if r.LogType == "scene" {
			if _, exists := execMap[execID]; !exists {
				execMap[execID] = []map[string]interface{}{
					{
						"execId":        r.ExecID,
						"sceneId":       r.SceneDetail.SceneID,
						"sceneName":     r.SceneDetail.SceneName,
						"state":         r.SceneDetail.State,
						"finishCount":   r.SceneDetail.FinishCount,
						"successCount":  r.SceneDetail.SuccessCount,
						"failCount":     r.SceneDetail.FailCount,
						"duration":      r.SceneDetail.Duration,
						"actionRecords": []map[string]interface{}{},
					},
				}
			} else {
				// execID 存在
				scene := make(map[string]interface{})
				scene["sceneId"] = r.SceneDetail.SceneID
				scene["sceneName"] = r.SceneDetail.SceneName
				scene["state"] = r.SceneDetail.State
				scene["finishCount"] = r.SceneDetail.FinishCount
				scene["successCount"] = r.SceneDetail.SuccessCount
				scene["failCount"] = r.SceneDetail.FailCount
				scene["duration"] = r.SceneDetail.Duration
				scene["actionRecords"] = []map[string]interface{}{}
				execMap[execID] = append(execMap[execID], scene)
			}
		}
	}

	// 查找 action 类型的记录,并加进去 scene
	for _, r := range taskRunRecord {
		execID := r.ExecID
		if r.LogType == "action" {
			for i, scene := range execMap[execID] {
				if scene["sceneId"] == r.ActionDetail.SceneID {
					actionRecord := map[string]interface{}{
						"actionId":   r.ActionDetail.ActionID,
						"actionName": r.ActionDetail.ActionName,
						"state":      r.ActionDetail.State,
						"duration":   r.ActionDetail.Duration,
						"error":      r.ActionDetail.Error,
						"request":    r.ActionDetail.Request,
						"response":   r.ActionDetail.Response,
					}
					scene["actionRecords"] = append(scene["actionRecords"].([]map[string]interface{}), actionRecord)
					execMap[execID][i] = scene
					break
				}
			}
		}
	}

	// 将聚合结果添加到响应中
	resp.Data = execMap
	return
}
