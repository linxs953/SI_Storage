package testdataservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTestDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTestDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTestDataLogic {
	return &ListTestDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListTestDataLogic) ListTestData(in *storage.Empty) (*storage.TestDataListResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.TestDataListResponse{}, nil
}
