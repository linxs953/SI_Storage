package sceneconfigservicelogic

import (
	"context"

	"Storage/internal/errors"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSceneConfigsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSceneConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSceneConfigsLogic {
	return &ListSceneConfigsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSceneConfigsLogic) ListSceneConfigs(in *storage.ListSceneConfigsRequest) (*storage.SceneConfigListResponse, error) {
	// 获取场景模板模型
	sceneTemplateModel, err := l.svcCtx.SceneTemplateModel()
	if err != nil {
		l.Errorf("Failed to get scene template model: %v", err)
		return &storage.SceneConfigListResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to get scene template model",
			},
		}, nil
	}

	// 设置默认分页参数
	page := int(in.Page)
	pageSize := int(in.PageSize)

	// 查询场景列表
	scenes, _, err := sceneTemplateModel.FindAll(l.ctx, page, pageSize)
	if err != nil {
		l.Errorf("Failed to list scenes: %v", err)
		return &storage.SceneConfigListResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to list scenes",
			},
		}, nil
	}

	// 准备响应数据
	configs := make([]*storage.SceneConfig, 0)
	for _, scene := range scenes {
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
		configs = append(configs, &storage.SceneConfig{
			SceneId: scene.SceneId,
			Name:    scene.SceneName,
			Desc:    scene.SceneDesc,
			Retry: &storage.RetrySetting{
				Enabled:  scene.Strategy.Retry.Enabled,
				MaxRetry: int64(scene.Strategy.Retry.MaxRetry),
				Interval: int64(scene.Strategy.Retry.Interval),
			},
			Timeout: &storage.TimeoutSetting{
				Enabled:  scene.Strategy.Timeout.Enabled,
				Duration: int64(scene.Strategy.Timeout.Duration),
			},
			RelatedApi: relatedApis,
			CreateAt:   scene.CreateAt.Format("2006-01-02 15:04:05"),
			UpdateAt:   scene.UpdateAt.Format("2006-01-02 15:04:05"),
		})
	}

	// 返回响应
	return &storage.SceneConfigListResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "Scenes listed successfully",
		},
		Data: configs,
	}, nil
}
