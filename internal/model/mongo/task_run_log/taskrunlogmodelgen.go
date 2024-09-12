// Code generated by goctl. DO NOT EDIT.
package task_run_log

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type taskRunLogModel interface {
	Insert(ctx context.Context, data *TaskRunLog) error
	FindOne(ctx context.Context, id string) (*TaskRunLog, error)
	Update(ctx context.Context, data *TaskRunLog) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, id string) (int64, error)
}

type defaultTaskRunLogModel struct {
	conn *mon.Model
}

func newDefaultTaskRunLogModel(conn *mon.Model) *defaultTaskRunLogModel {
	return &defaultTaskRunLogModel{conn: conn}
}

func (m *defaultTaskRunLogModel) Insert(ctx context.Context, data *TaskRunLog) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultTaskRunLogModel) FindOne(ctx context.Context, id string) (*TaskRunLog, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data TaskRunLog

	err = m.conn.FindOne(ctx, &data, bson.M{"_id": oid})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultTaskRunLogModel) Update(ctx context.Context, data *TaskRunLog) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *defaultTaskRunLogModel) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, ErrInvalidObjectId
	}

	res, err := m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	return res, err
}
