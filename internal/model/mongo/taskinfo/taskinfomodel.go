package taskinfo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

var _ TaskInfoModel = (*customTaskInfoModel)(nil)

type (
	// TaskInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskInfoModel.
	TaskInfoModel interface {
		taskInfoModel
		FindAllTask(ctx context.Context) ([]*TaskInfo, error)
		FindByTaskId(ctx context.Context, taskId string) (*TaskInfo, error)
		DeleteByTaskId(ctx context.Context, taskId string) (int64, error)
	}

	customTaskInfoModel struct {
		*defaultTaskInfoModel
	}
)

// NewTaskInfoModel returns a model for the mongo.
func NewTaskInfoModel(url, db, collection string) TaskInfoModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customTaskInfoModel{
		defaultTaskInfoModel: newDefaultTaskInfoModel(conn),
	}
}

func (m *customTaskInfoModel) FindByTaskId(ctx context.Context, taskId string) (*TaskInfo, error) {
	var resp TaskInfo
	err := m.conn.FindOne(ctx, &resp, bson.M{"taskid": taskId})
	switch err {
	case nil:
		return &resp, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customTaskInfoModel) DeleteByTaskId(ctx context.Context, taskId string) (int64, error) {
	count, err := m.conn.DeleteOne(ctx, bson.M{"taskid": taskId})
	if err != nil {
		return 0, err
	}
	return count, err
}

func (m *customTaskInfoModel) FindAllTask(ctx context.Context) (tasks []*TaskInfo, err error) {
	err = m.conn.Find(ctx, &tasks, bson.M{})
	if err != nil {
		return nil, err
	}
	return
}
