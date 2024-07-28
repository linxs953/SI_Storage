package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/mon"
)

var _ TaskRunLogModel = (*customTaskRunLogModel)(nil)

type (
	// TaskRunLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskRunLogModel.
	TaskRunLogModel interface {
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
