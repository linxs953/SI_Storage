package testdataservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTestDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTestDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTestDataLogic {
	return &DeleteTestDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTestDataLogic) DeleteTestData(in *storage.DeleteTestDataRequest) (*storage.DeleteResponse, error) {
	return &storage.DeleteResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "delete successfully",
		},
		AffectedRows: 0,
	}, nil
}
