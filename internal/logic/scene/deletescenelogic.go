package scene

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type DeleteSceneLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSceneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSceneLogic {
	return &DeleteSceneLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSceneLogic) DeleteScene(req *types.DeleteSceneDto) (*types.DeleteSceneVO, error) {
	resp := &types.DeleteSceneVO{
		Code:    0,
		Message: "删除场景成功",
		Data:    nil,
	}
	var err error
	if req.Scid == "" {
		resp.Code = 1
		resp.Message = "场景id不能为空"
		return resp, err
	}
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	if err = mod.DeletBySceneId(context.Background(), req.Scid); err != nil {
		resp.Code = 2
		resp.Message = err.Error()
		return resp, err
	}
	return resp, err
}
