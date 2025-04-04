package reportservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTaskReportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTaskReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskReportLogic {
	return &DeleteTaskReportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTaskReportLogic) DeleteTaskReport(in *storage.GetTestReportRequest) (*storage.DeleteResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.DeleteResponse{
		Header: &storage.ResponseHeader{
			Code:    0,
			Message: "delete successfully",
		},
		AffectedRows: 0,
	}, nil

}
