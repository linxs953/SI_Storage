package scene

import (
	"context"

	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSceneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSceneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSceneLogic {
	return &GetSceneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSceneLogic) GetScene(req *types.GetSceneDto) (*sceneinfo.SceneInfo, error) {
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	smod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	var (
		err   error
		scene *sceneinfo.SceneInfo
	)
	if scene, err = smod.FindOneBySceneID(l.ctx, req.Scid); err != nil {
		return nil, err
	}
	return scene, nil
}
