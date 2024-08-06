package sync

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"lexa-engine/internal/logic/sync/apitest"
	synchronizer "lexa-engine/internal/logic/sync/synchronizer"
	"lexa-engine/internal/model/mongo/apidetail"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type SyncapiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncapiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncapiLogic {
	return &SyncapiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncapiLogic) Syncapi(req *types.StartDto, syncRecordId primitive.ObjectID) (resp *types.StartResp, err error) {
	logx.Info("启动 job")
	switch req.JobType {
	case "sync":
		{
			return StartSyncJob(l.svcCtx, req.JobSpec, syncRecordId)
		}
	case "apitest":
		{
			return StartJob(l.svcCtx, req.JobSpec, req.JobType)
		}
	}
	return
}

func StartSyncJob(svcCtx *svc.ServiceContext, jobSpec string, recordId primitive.ObjectID) (resp *types.StartResp, err error) {

	logx.Info("解析同步器 job")
	var synchronizer synchronizer.SyncSpec
	if err = JobUnmarshal([]byte(jobSpec), &synchronizer); err != nil {
		logx.Error("解析同步器 spec 失败")
		return
	}
	logx.Info("解析同步器 job 成功")
	if err = synchronizer.Sync(svcCtx, recordId); err != nil {
		return
	}
	resp = &types.StartResp{
		Code:    0,
		Message: "操作成功",
		Data:    nil,
	}
	return
}

func JobUnmarshal(spec []byte, targetType any) (err error) {
	if err = json.Unmarshal(spec, targetType); err != nil {
		return err
	}
	return
}

func StartJob(svcCtx *svc.ServiceContext, jobSpec string, jobType string) (resp *types.StartResp, err error) {
	switch jobType {
	case "apitest":
		{
			var acJob apitest.ApiTestJob
			if err = JobUnmarshal([]byte(jobSpec), &acJob); err != nil {
				logx.Error("解析apitest Job spec 失败")
				return
			}
			logx.Info("解析apitest job 成功")
			acJob.Build(svcCtx)
			// acJob.Run()
		}
	}
	return
}

// 获取所有 api 的记录

func (l *SyncapiLogic) FetchAllApiInfo() (apis []apidetail.Apidetail, err error) {
	return
}
