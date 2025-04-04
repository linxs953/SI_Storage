package api

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const ApiCollectionName = "apis" // Collection name for APIs

type ApiModel interface {
	InsertApi(ctx context.Context, data *Api) error
	UpdateApi(ctx context.Context, apiId string, data *Api) error
	FindOneByApiID(ctx context.Context, apiId string) (*Api, error)
	FindByProjectID(ctx context.Context, projectId string) ([]*Api, error)
	DeleteOneByApiID(ctx context.Context, apiId string) error
	FindAll(ctx context.Context) ([]*Api, error)
}

type defaultApiModel struct {
	conn *mon.Model
}

func NewApiModel(url, db, collection string) ApiModel {
	conn := mon.MustNewModel(url, db, collection)
	return &defaultApiModel{
		conn: conn,
	}
}

func (m *defaultApiModel) InsertApi(ctx context.Context, data *Api) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}

	data.CreateAt = time.Now()
	data.UpdateAt = time.Now()

	_, err := m.conn.InsertOne(ctx, data)
	if mongo.IsDuplicateKeyError(err) {
		// If duplicate key, try to update instead
		return m.UpdateApi(ctx, data.ApiID, data)
	}
	return err
}

func (m *defaultApiModel) UpdateApi(ctx context.Context, apiId string, data *Api) error {
	filter := bson.M{"apiId": apiId}
	data.UpdateAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":        data.Name,
			"method":      data.Method,
			"path":        data.Path,
			"description": data.Description,
			"headers":     data.Headers,
			"parameters":  data.Parameters,
			"responses":   data.Responses,
			"rawData":     data.RawData,
			"projectId":   data.ProjectID,
			"taskId":      data.TaskID,
			"updateAt":    data.UpdateAt,
		},
	}

	_, err := m.conn.UpdateOne(ctx, filter, update)
	return err
}

func (m *defaultApiModel) FindOneByApiID(ctx context.Context, apiId string) (*Api, error) {
	var api Api
	err := m.conn.FindOne(ctx, bson.M{"apiId": apiId}, &api)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &api, nil
}

func (m *defaultApiModel) FindByProjectID(ctx context.Context, projectId string) ([]*Api, error) {
	var apis []*Api
	err := m.conn.Find(ctx, bson.M{"projectId": projectId}, &apis)
	if err != nil {
		return nil, err
	}
	return apis, nil
}

func (m *defaultApiModel) DeleteOneByApiID(ctx context.Context, apiId string) error {
	_, err := m.conn.DeleteOne(ctx, bson.M{"apiId": apiId})
	return err
}

func (m *defaultApiModel) FindAll(ctx context.Context) ([]*Api, error) {
	var apis []*Api
	err := m.conn.Find(ctx, bson.M{}, &apis)
	if err != nil {
		return nil, err
	}
	return apis, nil
}
