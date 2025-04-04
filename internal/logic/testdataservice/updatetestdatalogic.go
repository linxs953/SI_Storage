package testdataservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTestDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTestDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTestDataLogic {
	return &UpdateTestDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTestDataLogic) UpdateTestData(in *storage.UpdateTestDataRequest) (*storage.TestDataResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.TestDataResponse{}, nil
}
