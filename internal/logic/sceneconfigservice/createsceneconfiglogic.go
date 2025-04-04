package sceneconfigservicelogic

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"Storage/internal/errors"
	"Storage/internal/model/scene"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSceneConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSceneConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSceneConfigLogic {
	return &CreateSceneConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSceneConfigLogic) CreateSceneConfig(in *storage.CreateSceneConfigRequest) (*storage.SceneConfigResponse, error) {
	// 准备场景模板数据

	relatedApi := make([]*scene.RelatedApi, 0)

	if in.RelatedApi == nil {
		return &storage.SceneConfigResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.CreateSceneConfigError),
				Message: "Invalid related api",
			},
			Data: nil,
		}, nil
	}
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

	sceneTemplate := &scene.Scenetempmodel{
		SceneId:    generateSceneId(in),
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
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}

	// 调用场景模板模型的创建方法
	sceneTemplateModel, err := l.svcCtx.SceneTemplateModel()
	if err != nil {
		l.Errorf("Failed to get scene template model: %v", err)
		return nil, err
	}

	err = sceneTemplateModel.Create(l.ctx, sceneTemplate)
	if err != nil {
		l.Errorf("Failed to create scene template: %v", err)
		return nil, err
	}

	// 构造响应
	return &storage.SceneConfigResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "Success",
		},
		Data: &storage.SceneConfig{
			SceneId: sceneTemplate.SceneId,
			Name:    sceneTemplate.SceneName,
		},
	}, nil
}

func generateSceneId(in *storage.CreateSceneConfigRequest) string {
	// 使用更安全的唯一ID生成方法
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d-%s",
		in.Name,
		time.Now().UnixNano(),
		uuid.New().String(),
	)))

	// 取哈希的前16个字符作为ID
	return fmt.Sprintf("scene-%x", hash[:8])
}
