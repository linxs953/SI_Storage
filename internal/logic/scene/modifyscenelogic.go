package scene

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"lexa-engine/internal/logic"
	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type ModifySceneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifySceneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifySceneLogic {
	return &ModifySceneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifySceneLogic) ModifyScene(req *types.SceneUpdate, scendId string) (*sceneinfo.SceneInfo, error) {
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	smod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")

	// 更新sceneinfo
	var existScene *sceneinfo.SceneInfo
	var err error
	if existScene, err = smod.FindOneBySceneID(l.ctx, scendId); err != nil {
		_ = existScene
		return nil, err
	}

	actBytes, err := json.Marshal(req.Actions)
	if err != nil {
		return nil, err
	}
	var actions []logic.Action
	if err = json.Unmarshal(actBytes, &actions); err != nil {
		return nil, err
	}
	if req.Description != "" {
		existScene.Description = req.Description
	}
	if req.Scname != "" {
		existScene.SceneName = req.Scname
	}
	if req.Retry > 0 {
		existScene.Retry = req.Retry
	}
	if req.Timeout > 0 {
		existScene.Timeout = req.Timeout
	}
	// if req.Key != "" {
	// 	existScene.SearchKey = req.Key
	// }
	// if req.Author != "" {
	// 	existScene.Author = req.Author
	// }
	// if req.Env != "" {
	// 	existScene.EnvKey = req.Env
	// }
	if len(req.Actions) > 0 {
		existScene.Actions = actions
	}
	if _, err = smod.Update(context.Background(), existScene); err != nil {
		return nil, err
	}
	return existScene, nil
}
