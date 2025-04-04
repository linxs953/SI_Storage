// Code generated by goctl. DO NOT EDIT.
// goctl 1.7.6

package scene

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type scenetempmodelModel interface {
	Insert(ctx context.Context, data *Scenetempmodel) error
	FindOne(ctx context.Context, id string) (*Scenetempmodel, error)
	Update(ctx context.Context, data *Scenetempmodel) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, id string) (int64, error)
	Count(ctx context.Context) (int64, error)
}

type defaultScenetempmodelModel struct {
	conn *mon.Model
}

func newDefaultScenetempmodelModel(conn *mon.Model) *defaultScenetempmodelModel {
	return &defaultScenetempmodelModel{conn: conn}
}

func (m *defaultScenetempmodelModel) Insert(ctx context.Context, data *Scenetempmodel) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultScenetempmodelModel) FindOne(ctx context.Context, id string) (*Scenetempmodel, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data Scenetempmodel

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

func (m *defaultScenetempmodelModel) Update(ctx context.Context, data *Scenetempmodel) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *defaultScenetempmodelModel) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, ErrInvalidObjectId
	}

	res, err := m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	return res, err
}


func (m *defaultScenetempmodelModel) Count(ctx context.Context) (int64, error) {
	return m.conn.CountDocuments(ctx, bson.M{})
}