package task

import (
	"context"
	mgo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/task_run_log"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"math"
	"strconv"

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
		Code:     0,
		Message:  "success",
		TaskMeta: types.TaskMeta{},
		TaskRun:  []types.TaskRecord{},
	}
	murl := mgo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskLogModel := task_run_log.NewTaskRunLogModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskRunLog")
	taskModel := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")

	// 获取数据总量
	totalCount, err := taskLogModel.CountTaskRecords(l.ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	taskInfo, err := taskModel.FindByTaskId(l.ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	pageNum, err := strconv.Atoi(req.PageNum)
	if err != nil {
		return nil, err
	}
	pageSize, err := strconv.Atoi(req.PageSize)
	if err != nil {
		return nil, err
	}

	recordList, err := getAllTasks(l.ctx, taskLogModel, req.TaskId, pageNum, pageSize)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	for _, r := range recordList {
		sceneDetails, err := getSceneDetail(l.ctx, taskLogModel, r.ExecID)
		if err != nil {
			return nil, err
		}

		data = append(data, map[string]interface{}{
			"taskName":     r.TaskName,
			"execId":       r.ExecID,
			"taskId":       taskInfo.TaskID,
			"sceneCount":   r.SceneCount,
			"state":        r.State,
			"sceneDetails": sceneDetails,
			"createTime":   r.CreateTime,
			"updateTime":   r.UpdateTime,
		})
	}

	// 将 data 中的每个 map 转换为 types.TaskRecord
	var taskRecords []types.TaskRecord
	for _, item := range data {
		state, err := strconv.Atoi(item["state"].(string))
		if err != nil {
			return nil, err
		}
		finishTime := item["updateTime"].(string)
		if state == 0 {
			finishTime = ""
		}
		taskRecord := types.TaskRecord{
			RunId:        item["execId"].(string),
			TaskId:       item["taskId"].(string),
			State:        state,
			SceneRecords: item["sceneDetails"].([]map[string]interface{}),
			CreateTime:   item["createTime"].(string),
			FinishTime:   finishTime,
		}
		taskRecords = append(taskRecords, taskRecord)
	}

	resp.TaskMeta = types.TaskMeta{
		TaskName:   taskInfo.TaskName,
		TaskID:     taskInfo.TaskID,
		Author:     taskInfo.Author,
		SceneCount: len(taskInfo.Scenes),
		CreateTime: taskInfo.CreateAt.Format("2006-01-02 15:04:05"),
		UpdateTime: taskInfo.UpdateAt.Format("2006-01-02 15:04:05"),
	}

	resp.TaskRun = taskRecords
	resp.CurrentPage = pageNum
	resp.TotalPage = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	resp.TotalNum = int(totalCount)
	return
}

type RecordMeta struct {
	TaskId     string `json:"taskId"`
	ExecID     string `json:"execId"`
	TaskName   string `json:"taskName"`
	SceneCount int    `json:"sceneCount"`
	State      string `json:"state"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}

func getSceneDetail(ctx context.Context, taskLogModel task_run_log.TaskRunLogModel, execId string) ([]map[string]interface{}, error) {
	taskRunRecord, err := taskLogModel.FindTaskRunRecord(ctx, execId)
	if err != nil {
		return nil, err
	}

	// 根据 record 的 sceneDetail 来判断 task 的状态
	// 使用map来聚合相同execID的记录
	execMap := make(map[string]string)
	var sceneDetails []map[string]interface{}

	for _, r := range taskRunRecord {
		if r.LogType == "scene" {
			if _, exists := execMap[r.SceneDetail.SceneID]; !exists {
				sceneDetails = append(sceneDetails, map[string]interface{}{
					"execId":        r.ExecID,
					"sceneId":       r.SceneDetail.SceneID,
					"sceneName":     r.SceneDetail.SceneName,
					"state":         r.SceneDetail.State,
					"finishCount":   r.SceneDetail.FinishCount,
					"successCount":  r.SceneDetail.SuccessCount,
					"failCount":     r.SceneDetail.FailCount,
					"duration":      r.SceneDetail.Duration,
					"actionRecords": []map[string]interface{}{},
				})
				execMap[r.SceneDetail.SceneID] = ""
			}
		}
	}

	// 查找 action 类型的记录,并加进去 scene
	for _, r := range taskRunRecord {
		if r.LogType == "action" {
			for i, scene := range sceneDetails {
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
					sceneDetails[i]["actionRecords"] = append(sceneDetails[i]["actionRecords"].([]map[string]interface{}), actionRecord)
					break
				}
			}
		}
	}
	return sceneDetails, nil
}

func getAllTasks(ctx context.Context, taskLogModel task_run_log.TaskRunLogModel, taskId string, pageNum, pageSize int) ([]*RecordMeta, error) {
	var recordList []*RecordMeta
	taskRunRecords, err := taskLogModel.FindAllTaskRecords(ctx, taskId, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	for _, r := range taskRunRecords {
		recordList = append(recordList, &RecordMeta{
			TaskId:     r.TaskID,
			ExecID:     r.ExecID,
			TaskName:   r.TaskDetail.TaskName,
			SceneCount: r.TaskDetail.TaskSceneCount,
			State:      strconv.Itoa(r.TaskDetail.TaskState),
			CreateTime: r.CreateAt.Format("2006-01-02 15:04:05"),
			UpdateTime: r.UpdateAt.Format("2006-01-02 15:04:05"),
		})
	}

	return recordList, nil
}
