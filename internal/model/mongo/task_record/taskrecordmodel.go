package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ TaskRecordModel = (*customTaskRecordModel)(nil)

type (
	// TaskRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskRecordModel.
	TaskRecordModel interface {
		taskRecordModel
		IsRecordExist(ctx context.Context, sceneId string) bool
		FindAndUpdate(ctx context.Context, fileter bson.M, data bson.M) (*mongo.UpdateResult, error)
	}

	customTaskRecordModel struct {
		*defaultTaskRecordModel
	}
)

// NewTaskRecordModel returns a model for the mongo.
func NewTaskRecordModel(url, db, collection string) TaskRecordModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customTaskRecordModel{
		defaultTaskRecordModel: newDefaultTaskRecordModel(conn),
	}
}

func (ctm *customTaskRecordModel) Insert(ctx context.Context, data *TaskRecord) error {
	return ctm.defaultTaskRecordModel.Insert(ctx, data)
}

func (ctm *customTaskRecordModel) IsRecordExist(ctx context.Context, sceneId string) bool {
	var record TaskRecord
	err := ctm.conn.FindOne(ctx, &record, bson.M{"sceneId": sceneId, "state": 0})
	if err != nil && err == mongo.ErrNoDocuments {
		return false
	}
	if err != nil {
		logx.Error(err)
	}
	return true
}

func (ctm *customTaskRecordModel) FindAndUpdate(ctx context.Context, fileter bson.M, data bson.M) (*mongo.UpdateResult, error) {
	newRecord := mongo.UpdateResult{}
	err := ctm.conn.FindOneAndUpdate(ctx, &newRecord, fileter, data)
	if err != nil {
		logx.Error(err)
		return nil, err
	}
	return &newRecord, nil
}
