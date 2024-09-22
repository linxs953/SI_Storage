package task_run_log

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ TaskRunLogModel = (*customTaskRunLogModel)(nil)

type (
	// TaskRunLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskRunLogModel.
	TaskRunLogModel interface {
		FindLogRecord(ctx context.Context, execID string, sceneId string, actionId string, logType string) (*TaskRunLog, error)
		FindTaskRunRecord(ctx context.Context, taskId string) ([]*TaskRunLog, error)
		FindAllSceneRecord(ctx context.Context, execId string, sceneId string) ([]*TaskRunLog, error)
		FindAllTaskRecords(ctx context.Context, taskId string) ([]*TaskRunLog, error)
		taskRunLogModel
	}

	customTaskRunLogModel struct {
		*defaultTaskRunLogModel
	}
)

// NewTaskRunLogModel returns a model for the mongo.
func NewTaskRunLogModel(url, db, collection string) TaskRunLogModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customTaskRunLogModel{
		defaultTaskRunLogModel: newDefaultTaskRunLogModel(conn),
	}
}

func (m *customTaskRunLogModel) Insert(ctx context.Context, data *TaskRunLog) error {
	return m.defaultTaskRunLogModel.Insert(ctx, data)
}

func (m *customTaskRunLogModel) FindLogRecord(ctx context.Context, execID string, sceneId string, actionId string, logType string) (*TaskRunLog, error) {
	var record TaskRunLog
	if logType == "scene" {
		if err := m.conn.FindOne(ctx, &record, bson.M{"execId": execID, "logType": "scene", "sceneDetail.sceneId": sceneId}); err != nil {
			logx.Error(err)
			return nil, err
		}
	}
	if logType == "action" {
		if err := m.conn.FindOne(ctx, &record, bson.M{"execId": execID, "logType": "action", "actionDetail.sceneId": sceneId, "actionDetail.actionId": actionId}); err != nil {
			return nil, err
		}
	}
	if logType == "task" {
		if err := m.conn.FindOne(ctx, &record, bson.M{"execId": execID, "logType": "task"}); err != nil {
			return nil, err
		}
	}

	return &record, nil
}

func (m *customTaskRunLogModel) FindAllSceneRecord(ctx context.Context, execId string, sceneId string) ([]*TaskRunLog, error) {
	var sceneRecords []*TaskRunLog
	sortOptions := options.Find()
	sortOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if err := m.conn.Find(ctx, &sceneRecords, bson.M{"execId": execId, "logType": "scene", "sceneDetail.sceneId": sceneId}, sortOptions); err != nil {
		return nil, err
	}
	return sceneRecords, nil
}

func (m *customTaskRunLogModel) FindTaskRunRecord(ctx context.Context, taskId string) ([]*TaskRunLog, error) {
	var recordList []*TaskRunLog
	sortOptions := options.Find()
	sortOptions.SetSort(bson.D{{Key: "createAt", Value: -1}})
	if err := m.conn.Find(ctx, &recordList, bson.M{"execId": taskId}); err != nil {
		return nil, err
	}
	return recordList, nil
}

func (m *customTaskRunLogModel) FindAllTaskRecords(ctx context.Context, taskId string) ([]*TaskRunLog, error) {
	var recordList []*TaskRunLog
	sortOptions := options.Find()
	sortOptions.SetSort(bson.D{{Key: "createAt", Value: -1}}) // 使用降序排列
	if err := m.conn.Find(ctx, &recordList, bson.M{"taskId": taskId, "logType": "task"}, sortOptions); err != nil {
		return nil, err
	}
	return recordList, nil
}
