package sceneinfo

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ SceneInfoModel = (*customSceneInfoModel)(nil)

type (
	// SceneInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSceneInfoModel.
	SceneInfoModel interface {
		sceneInfoModel
		SearchWithKeyword(ctx context.Context, keyword string) ([]*SceneInfo, error)
		FindOneBySceneID(ctx context.Context, sceneID string) (*SceneInfo, error)
		FindList(ctx context.Context, page int64, pageSize int64) ([]*SceneInfo, error)
		DeletBySceneId(ctx context.Context, sceneId string) error
	}

	customSceneInfoModel struct {
		*defaultSceneInfoModel
	}
)

// NewSceneInfoModel returns a model for the mongo.
func NewSceneInfoModel(url, db, collection string) SceneInfoModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customSceneInfoModel{
		defaultSceneInfoModel: newDefaultSceneInfoModel(conn),
	}
}

func (m *customSceneInfoModel) FindOneBySceneID(ctx context.Context, sceneID string) (*SceneInfo, error) {
	var result SceneInfo
	err := m.conn.FindOne(ctx, &result, bson.M{"scene.sceneid": sceneID})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *customSceneInfoModel) FindList(ctx context.Context, page int64, pageSize int64) ([]*SceneInfo, error) {
	var err error
	var scenes []*SceneInfo
	skip := (page - 1) * pageSize
	err = m.conn.Find(ctx, &scenes, bson.D{}, options.Find().SetSkip(skip).SetLimit(pageSize))
	if err != nil {
		return nil, err
	}
	return scenes, err
}

func (m *customSceneInfoModel) DeletBySceneId(ctx context.Context, sceneId string) error {
	var err error
	if _, err = m.conn.DeleteOne(ctx, bson.M{"scene.sceneid": sceneId}); err != nil {
		return err
	}
	return nil
}

func (m *customSceneInfoModel) SearchWithKeyword(ctx context.Context, keyword string) ([]*SceneInfo, error) {
	var scenes []*SceneInfo
	var err error
	v := fmt.Sprintf(".*%s.*", keyword)
	err = m.conn.Find(ctx, &scenes, bson.M{"scene.sceneName": bson.D{{Key: "$regex", Value: v}}})
	if err != nil {
		return nil, err
	}
	return scenes, err
}
