package testdataservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTestDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTestDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTestDataLogic {
	return &GetTestDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTestDataLogic) GetTestData(in *storage.GetTestDataRequest) (*storage.TestDataResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.TestDataResponse{}, nil
}
