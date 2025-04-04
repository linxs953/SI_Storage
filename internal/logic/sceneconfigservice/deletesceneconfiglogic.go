package sceneconfigservicelogic

import (
	"context"

	"Storage/internal/errors"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSceneConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSceneConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSceneConfigLogic {
	return &DeleteSceneConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSceneConfigLogic) DeleteSceneConfig(in *storage.DeleteSceneConfigRequest) (*storage.DeleteResponse, error) {
	// 获取场景模板模型
	sceneTemplateModel, err := l.svcCtx.SceneTemplateModel()
	if err != nil {
		l.Errorf("Failed to get scene template model: %v", err)
		return &storage.DeleteResponse{
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
		return &storage.DeleteResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to find scene record",
			},
		}, nil
	}
	if existingScene == nil {
		// 场景不存在
		return &storage.DeleteResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.SceneNotFound),
				Message: "Scene not found",
			},
		}, nil
	}

	// 删除场景
	err = sceneTemplateModel.DeleteBySceneId(l.ctx, in.SceneId)
	if err != nil {
		l.Errorf("Failed to delete scene: %v", err)
		return &storage.DeleteResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "Failed to delete scene",
			},
		}, nil
	}

	// 返回成功响应
	return &storage.DeleteResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "Scene deleted successfully",
		},
	}, nil
}
