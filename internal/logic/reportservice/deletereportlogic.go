package reportservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteReportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteReportLogic {
	return &DeleteReportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteReportLogic) DeleteReport(in *storage.GetTestReportRequest) (*storage.DeleteResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.DeleteResponse{}, nil
}
