package generateservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateExtractorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateExtractorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateExtractorLogic {
	return &GenerateExtractorLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateExtractorLogic) GenerateExtractor(in *storage.GenerateExtractorRequest) (*storage.GenerateExtractorResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.GenerateExtractorResponse{}, nil
}
