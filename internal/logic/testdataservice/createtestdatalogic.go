package testdataservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTestDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTestDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTestDataLogic {
	return &CreateTestDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 测试数据CRUD
func (l *CreateTestDataLogic) CreateTestData(in *storage.CreateTestDataRequest) (*storage.TestDataResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.TestDataResponse{}, nil
}
