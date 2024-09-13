package task_run_log

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

var _ TaskRunLogModel = (*customTaskRunLogModel)(nil)

type (
	// TaskRunLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskRunLogModel.
	TaskRunLogModel interface {
		FindLogRecord(ctx context.Context, execID string, sceneId string, actionId string, logType string) (*TaskRunLog, error)
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
			return nil, err
		}
	}
	if logType == "action" {
		if err := m.conn.FindOne(ctx, &record, bson.M{"execId": execID, "logType": "action", "actionDetail.sceneId": sceneId, "actionDetail.actionId": actionId}); err != nil {
			return nil, err
		}
	}

	return &record, nil
}
