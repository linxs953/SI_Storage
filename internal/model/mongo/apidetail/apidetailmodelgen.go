package apidetail

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type apidetailModel interface {
	Insert(ctx context.Context, data *Apidetail) error
	FindOne(ctx context.Context, id string) (*Apidetail, error)
	Update(ctx context.Context, data *Apidetail) (*mongo.UpdateResult, error)
	Delete(ctx context.Context, id string) (int64, error)
}

type defaultApidetailModel struct {
	conn *mon.Model
}

func newDefaultApidetailModel(conn *mon.Model) *defaultApidetailModel {
	return &defaultApidetailModel{conn: conn}
}

func (m *defaultApidetailModel) Insert(ctx context.Context, data *Apidetail) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}
	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultApidetailModel) FindOne(ctx context.Context, id string) (*Apidetail, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data Apidetail

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

func (m *defaultApidetailModel) Update(ctx context.Context, data *Apidetail) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *defaultApidetailModel) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, ErrInvalidObjectId
	}

	res, err := m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	return res, err
}
