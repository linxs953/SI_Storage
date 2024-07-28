package task

import (
	"context"
	"fmt"

	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSceneWithIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSceneWithIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSceneWithIdLogic {
	return &GetSceneWithIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSceneWithIdLogic) GetSceneWithId(req *types.GetSceneWithIdDto) (resp *types.GetSceneWithIdResp, err error) {
	if req.TaskID == "" {
		return nil, fmt.Errorf("task id is empty")
	}
	resp = &types.GetSceneWithIdResp{}
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskInfo, err := taskMod.FindByTaskId(context.Background(), req.TaskID)
	if err != nil {
		return nil, err
	}
	resp.Data.SceneInfo = map[string]interface{}{}
	for _, scene := range taskInfo.Scenes {
		// 维护一个
		resp.Data.Scenes = append(resp.Data.Scenes, scene.SceneId)

		// 构建task中的scene信息
		resp.Data.SceneInfo[scene.SceneId] = taskScene{
			SceneId:     scene.SceneId,
			SceneName:   scene.SceneName,
			Key:         scene.SceneId,
			Description: scene.SceneName,
			Author:      scene.Author,
		}
	}
	return
}

type taskScene struct {
	SceneId     string `json:"sceneId"`
	SceneName   string `json:"sceneName"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Author      string `json:"author"`
}
