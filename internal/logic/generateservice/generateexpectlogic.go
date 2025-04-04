package generateservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateExpectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateExpectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateExpectLogic {
	return &GenerateExpectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GenerateExpectLogic) GenerateExpect(in *storage.GenerateExpectRequest) (*storage.GenerateExpectResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.GenerateExpectResponse{}, nil
}
