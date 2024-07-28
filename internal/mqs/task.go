package mqs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"

	"lexa-engine/internal/logic/task"
	mong "lexa-engine/internal/model/mongo"
	model "lexa-engine/internal/model/mongo/task_record"
	task_run_model "lexa-engine/internal/model/mongo/task_run_log"
	"lexa-engine/internal/svc"
)

type TaskSuccess struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTaskSuccess(ctx context.Context, svcCtx *svc.ServiceContext) *TaskSuccess {
	return &TaskSuccess{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TaskSuccess) Consume(key, val string) error {
	logx.Infof("Receive message %v", val)
	var event task.TaskEvent
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	if err := json.Unmarshal([]byte(val), &event); err != nil {
		logx.Error("[消费者] 解析消费事件失败", err)
		return err
	}
	switch event.EventType {
	case "task_start":
		{
			record := &model.TaskRecord{
				TaskType:        "api_auto",
				SceneID:         event.EventMsg.TaskID,
				RequestID:       event.EventMsg.RequestID,
				Total:           event.EventMsg.Total,
				SuccessCount:    0,
				FailedCount:     0,
				State:           event.EventMsg.State,
				ActionRunDetail: []model.ActionRunDetail{},
				CreateAt:        event.EventMsg.StartAt,
				UpdateAt:        time.Now(),
			}
			err := insertTaskRunRecord(murl, l.svcCtx.Config.Database.Mongo.UseDb, record)
			if err != nil {
				return err
			}
			runLog := &task_run_model.TaskRunLog{
				Type:      "scene_create",
				RequestID: event.EventMsg.RequestID,
				SceneID:   event.EventMsg.TaskID,
				StartTime: event.EventMsg.StartAt,
				Total:     event.EventMsg.Total,
				UpdateAt:  time.Now(),
				CreateAt:  time.Now(),
			}

			if err = insertTaskRunLog(murl, l.svcCtx.Config.Database.Mongo.UseDb, runLog); err != nil {
				return err
			}
			return nil
		}
	case "task_update":
		{
			return nil
		}
	case "action_start":
		{
			return nil
		}
	case "action_update":
		{
			return nil
		}
	case "task_finish":
		{
			record := &model.TaskRecord{
				TaskType:  "api_auto",
				SceneID:   event.EventMsg.TaskID,
				RequestID: event.EventMsg.RequestID,
				Total:     event.EventMsg.Total,
				State:     event.EventMsg.State,
				Duration:  time.Since(event.EventMsg.StartAt),
				FinishAt:  event.EventMsg.FinishAt,
				CreateAt:  event.EventMsg.StartAt,
				UpdateAt:  time.Now(),
			}
			err := updateTaskRecord(murl, l.svcCtx.Config.Database.Mongo.UseDb, record)
			if err != nil {
				return err
			}
			runLog := &task_run_model.TaskRunLog{
				Type:       "scene_finish",
				RequestID:  event.EventMsg.RequestID,
				SceneID:    event.EventMsg.TaskID,
				Total:      event.EventMsg.Total,
				FinishTime: event.EventMsg.FinishAt,
				UpdateAt:   time.Now(),
				CreateAt:   time.Now(),
			}

			if err = insertTaskRunLog(murl, l.svcCtx.Config.Database.Mongo.UseDb, runLog); err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("未定义的事件类型")
}

func insertTaskRunRecord(mongoUrl string, useDb string, data *model.TaskRecord) error {
	mod := model.NewTaskRecordModel(mongoUrl, useDb, "TaskRunRecord")
	var err error
	if mod.IsRecordExist(context.Background(), data.SceneID) {
		logx.Errorf("TaskRunRecord [SceneID=%v] 存在进行中的任务,过滤", data.SceneID)
		return nil
	}
	if err = mod.Insert(context.Background(), data); err != nil {
		return err
	}
	return err
}

func updateTaskRecord(mongoUrl string, useDb string, data *model.TaskRecord) error {
	mod := model.NewTaskRecordModel(mongoUrl, useDb, "TaskRunRecord")
	var err error
	_, err = mod.FindAndUpdate(context.Background(),
		bson.M{"sceneId": data.SceneID, "requestId": data.RequestID},
		bson.M{"$set": bson.M{"state": data.State, "duration": data.Duration, "finishAt": data.FinishAt}})
	if err != nil {
		return err
	}
	return nil
}

func insertTaskRunLog(mongoUrl string, useDb string, data *task_run_model.TaskRunLog) error {
	var err error
	mod := task_run_model.NewTaskRunLogModel(mongoUrl, useDb, "TaskRunLog")
	if err = mod.Insert(context.Background(), data); err != nil {
		return err
	}
	return err
}
