package sceneconfigservicelogic

import (
	"context"

	"Storage/internal/errors"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSceneConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSceneConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSceneConfigLogic {
	return &GetSceneConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSceneConfigLogic) GetSceneConfig(in *storage.GetSceneConfigRequest) (*storage.SceneConfigResponse, error) {
	// 获取场景模板模型
	sceneTemplateModel, err := l.svcCtx.SceneTemplateModel()
	if err != nil {
		l.Errorf("Failed to get scene template model: %v", err)
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to get scene template model",
			},
		}, nil
	}

	// 查找场景
	scene, err := sceneTemplateModel.FindBySceneId(l.ctx, in.SceneId)
	if err != nil {
		l.Errorf("Error finding scene: %v", err)
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to find scene record",
			},
		}, nil
	}

	// 检查场景是否存在
	if scene == nil {
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.SceneNotFound),
				Message: "Scene not found",
			},
		}, nil
	}

	// 准备关联API数据
	relatedApis := make([]*storage.RelatedApi, 0)
	for _, api := range scene.RelatedApi {
		relatedApis = append(relatedApis, &storage.RelatedApi{
			ApiId:      api.ApiId,
			Name:       api.Name,
			Enabled:    api.Enabled,
			Dependency: api.Dependency,
			Expect:     api.Expect,
			Extractor:  api.Extractor,
		})
	}

	// 返回场景配置
	return &storage.SceneConfigResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "Scene found successfully",
		},
		Data: &storage.SceneConfig{
			SceneId:    scene.SceneId,
			Name:       scene.SceneName,
			Desc:       scene.SceneDesc,
			RelatedApi: relatedApis,
			Retry: &storage.RetrySetting{
				Enabled:  scene.Strategy.Retry.Enabled,
				MaxRetry: int64(scene.Strategy.Retry.MaxRetry),
				Interval: int64(scene.Strategy.Retry.Interval),
			},
			Timeout: &storage.TimeoutSetting{
				Enabled:  scene.Strategy.Timeout.Enabled,
				Duration: int64(scene.Strategy.Timeout.Duration),
			},
			CreateAt: scene.CreateAt.Format("2006-01-02 15:04:05"),
			UpdateAt: scene.UpdateAt.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
