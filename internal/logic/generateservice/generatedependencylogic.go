package generateservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateDependencyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateDependencyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateDependencyLogic {
	return &GenerateDependencyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据ApiInfo生成依赖、提取器、预期
func (l *GenerateDependencyLogic) GenerateDependency(in *storage.GenerateDependencyRequest) (*storage.GenerateDependencyResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.GenerateDependencyResponse{}, nil
}
