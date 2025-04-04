package reportservicelogic

import (
	"context"

	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTestReportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTestReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTestReportLogic {
	return &GetTestReportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查看测试报告
func (l *GetTestReportLogic) GetTestReport(in *storage.GetTestReportRequest) (*storage.TestReportResponse, error) {
	// todo: add your logic here and delete this line

	return &storage.TestReportResponse{}, nil
}
