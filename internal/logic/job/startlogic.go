package job

import (
	"context"
	"encoding/json"
	synchronizer "lexa-engine/internal/logic/job/synchronizer"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartLogic {
	return &StartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StartLogic) Start(req *types.StartDto) (resp *types.StartResp, err error) {
	logx.Info("启动 job")
	switch req.JobType {
	case "sync":
		{
			return StartSyncJob(l.svcCtx, req.JobSpec)
		}
	}
	return
}

func StartSyncJob(svcCtx *svc.ServiceContext, jobSpec string) (resp *types.StartResp, err error) {
	logx.Info("解析同步器 job")
	var synchronizer synchronizer.SyncSpec
	if err = JobUnmarshal([]byte(jobSpec), &synchronizer); err != nil {
		logx.Error("解析同步器 spec 失败")
		return
	}
	logx.Info("解析同步器 job 成功")
	if err = synchronizer.Sync(svcCtx); err != nil {
		return
	}
	return
}

func JobUnmarshal(spec []byte, targetType any) (err error) {
	if err = json.Unmarshal(spec, targetType); err != nil {
		return err
	}
	return
}
