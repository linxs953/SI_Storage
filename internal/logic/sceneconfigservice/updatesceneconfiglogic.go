package sceneconfigservicelogic

import (
	"context"
	"time"

	"Storage/internal/errors"
	"Storage/internal/model/scene"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSceneConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSceneConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSceneConfigLogic {
	return &UpdateSceneConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateSceneConfigLogic) UpdateSceneConfig(in *storage.UpdateSceneConfigRequest) (*storage.SceneConfigResponse, error) {
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

	// 检查场景是否存在
	existingScene, err := sceneTemplateModel.FindBySceneId(l.ctx, in.SceneId)
	if err != nil {
		l.Errorf("Error finding scene: %v", err)
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to find scene record",
			},
		}, nil
	}
	if existingScene == nil {
		// 场景不存在
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.SceneNotFound),
				Message: "Scene not found",
			},
		}, nil
	}

	// 准备更新的场景数据
	relatedApi := make([]*scene.RelatedApi, 0)
	for _, api := range in.RelatedApi {
		relatedApi = append(relatedApi, &scene.RelatedApi{
			ApiId:      api.ApiId,
			Name:       api.Name,
			Enabled:    api.Enabled,
			Dependency: api.Dependency,
			Expect:     api.Expect,
			Extractor:  api.Extractor,
		})
	}

	// 更新场景模板
	updatedScene := &scene.Scenetempmodel{
		SceneName:  in.Name,
		SceneDesc:  in.Desc,
		RelatedApi: relatedApi,
		Strategy: &scene.SceneStrategy{
			Timeout: &scene.SceneTimeoutSetting{
				Duration: int(in.Timeout.Duration),
				Enabled:  in.Timeout.Enabled,
			},
			Retry: &scene.SceneRetrySetting{
				Enabled:  in.Retry.Enabled,
				MaxRetry: int(in.Retry.MaxRetry),
				Interval: int(in.Retry.Interval),
			},
		},
		UpdateAt: time.Now(),
	}

	// 执行更新
	err = sceneTemplateModel.UpdateBySceneId(l.ctx, in.SceneId, updatedScene)
	if err != nil {
		l.Errorf("Failed to update scene: %v", err)
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to update scene",
			},
		}, nil
	}

	// 返回成功响应
	return &storage.SceneConfigResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "Scene updated successfully",
		},
		Data: &storage.SceneConfig{
			SceneId: in.SceneId,
			Name:    updatedScene.SceneName,
		},
	}, nil
}
