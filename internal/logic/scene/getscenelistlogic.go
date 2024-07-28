package scene

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	mgoutil "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type GetSceneListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSceneListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSceneListLogic {
	return &GetSceneListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSceneListLogic) GetSceneList(req *types.GetSceneListDto) (resp *types.GetSceneListVO, err error) {
	resp = &types.GetSceneListVO{
		Code:    0,
		Message: "获取场景列表成功",
	}
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	scenes, err := mod.FindList(context.Background(), int64(req.Page), int64(req.PageSize))
	if err != nil {
		resp.Code = 1
		resp.Message = "获取场景列表失败"
		resp.Data = []map[string]interface{}{}
		return
	}
	var scenesMap []map[string]interface{}
	bts, err := json.Marshal(scenes)
	if err != nil {
		resp.Code = 2
		resp.Message = err.Error()
		resp.Data = []map[string]interface{}{}
		return
	}
	err = json.Unmarshal(bts, &scenesMap)
	if err != nil {
		resp.Code = 2
		resp.Message = err.Error()
		resp.Data = []map[string]interface{}{}
		return
	}
	for idx, scene := range scenesMap {
		createTime, err := time.Parse(time.RFC3339, scene["createAt"].(string))
		if err != nil {
			resp.Code = 3
			resp.Message = err.Error()
			resp.Data = []map[string]interface{}{}
			return resp, err
		}
		scenesMap[idx]["createAt"] = createTime.Format("2006-01-02 15:04:05")
		updateTime, err := time.Parse(time.RFC3339, scene["updateAt"].(string))
		if err != nil {
			resp.Code = 3
			resp.Message = err.Error()
			resp.Data = []map[string]interface{}{}
			return resp, err
		}
		scenesMap[idx]["updateAt"] = updateTime.Format("2006-01-02 15:04:05")
		scenesMap[idx]["actionCounts"] = len(scene["actions"].([]interface{}))
	}
	resp.Data = scenesMap
	return
}
