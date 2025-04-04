package scene

import (
	"Storage/internal/errors"
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ ScenetempmodelModel = (*customScenetempmodelModel)(nil)

type (
	// ScenetempmodelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customScenetempmodelModel.
	ScenetempmodelModel interface {
		scenetempmodelModel
		FindBySceneId(ctx context.Context, sceneId string) (*Scenetempmodel, error)
		FindAll(ctx context.Context, page, pageSize int) ([]*Scenetempmodel, int64, error)
		Create(ctx context.Context, data *Scenetempmodel) error
		UpdateBySceneId(ctx context.Context, sceneId string, data *Scenetempmodel) error
		DeleteBySceneId(ctx context.Context, sceneId string) error
	}

	customScenetempmodelModel struct {
		*defaultScenetempmodelModel
	}
)

// NewScenetempmodelModel returns a model for the mongo.
func NewScenetempmodelModel(url, db, collection string) ScenetempmodelModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customScenetempmodelModel{
		defaultScenetempmodelModel: newDefaultScenetempmodelModel(conn),
	}
}

// FindBySceneId retrieves a scene template by its scene ID
func (m *customScenetempmodelModel) FindBySceneId(ctx context.Context, sceneId string) (*Scenetempmodel, error) {
	result, err := m.FindOne(ctx, sceneId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// FindAll retrieves all scene templates with pagination
func (m *customScenetempmodelModel) FindAll(ctx context.Context, page, pageSize int) ([]*Scenetempmodel, int64, error) {
	// Count total documents
	total, err := m.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Set pagination options
	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSort(bson.D{{"createAt", -1}})

	// Find documentsd
	var results []*Scenetempmodel
	err = m.conn.Find(ctx, &results, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// Create adds a new scene template
func (m *customScenetempmodelModel) Create(ctx context.Context, data *Scenetempmodel) error {
	if data == nil {
		return nil
	}

	if data.SceneId == "" {
		return errors.New(errors.InvalidSceneIdError)
	}

	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}

	if data.RelatedApi == nil || len(data.RelatedApi) == 0 {
		return errors.New(errors.InvalidRelatedApiError)
	}

	// Set timestamps
	data.CreateAt = time.Now()
	data.UpdateAt = time.Now()

	// Insert document
	return m.Insert(ctx, data)
}

// UpdateBySceneId updates an existing scene template by scene ID
func (m *customScenetempmodelModel) UpdateBySceneId(ctx context.Context, sceneId string, data *Scenetempmodel) error {
	// Find the document first to get its ID
	template, err := m.FindBySceneId(ctx, sceneId)
	if err != nil || template == nil {
		return errors.New(errors.InvalidSceneIdError)
	}

	// Update timestamps
	data.UpdateAt = time.Now()

	// Set ID
	data.ID = template.ID

	// Update document
	_, err = m.Update(ctx, data)
	return err
}

// DeleteBySceneId removes a scene template by its scene ID
func (m *customScenetempmodelModel) DeleteBySceneId(ctx context.Context, sceneId string) error {
	// Find the document first to get its ID
	template, err := m.FindBySceneId(ctx, sceneId)
	if err != nil || template == nil {
		return err
	}

	id := template.ID.Hex()

	// Delete document
	_, err = m.Delete(ctx, id)
	return err
}
