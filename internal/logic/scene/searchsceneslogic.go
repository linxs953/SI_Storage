package scene

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type SearchScenesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchScenesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchScenesLogic {
	return &SearchScenesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchScenesLogic) SearchScenes(req *types.SearchScenesDto) (resp *types.SearchScentVo, err error) {
	resp = &types.SearchScentVo{
		Code:    0,
		Message: "搜索场景成功",
		Data:    nil,
	}
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	smod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	scenes, err := smod.SearchWithKeyword(context.Background(), req.Keyword)
	if err != nil {
		resp.Code = 1
		resp.Message = "搜索场景失败"
		return
	}
	for _, scene := range scenes {
		bts, err1 := json.Marshal(scene)
		if err1 != nil {
			resp.Code = 2
			resp.Message = "序列化场景失败"
			err = err1
			return
		}
		sc := make(map[string]interface{})
		err1 = json.Unmarshal(bts, &sc)
		if err1 != nil {
			resp.Code = 2
			resp.Message = "场景数据转换失败"
			err = err1
			return
		}
		resp.Data = append(resp.Data, sc)
	}
	return
}
