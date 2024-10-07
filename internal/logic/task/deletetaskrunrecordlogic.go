package task

import (
	"context"
	mgo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/task_run_log"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTaskRunRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTaskRunRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskRunRecordLogic {
	return &DeleteTaskRunRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTaskRunRecordLogic) DeleteTaskRunRecord(req *types.DeleteTaskRunRecordDto) (resp *types.DeleteTaskRunRecordResp, err error) {
	resp = &types.DeleteTaskRunRecordResp{
		Code:    0,
		Message: "success",
		Data:    nil,
	}
	murl := mgo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskLogModel := task_run_log.NewTaskRunLogModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskRunLog")
	err = taskLogModel.DeleteTaskRunRecordByExecId(l.ctx, req.ExecId)
	if err != nil {
		resp.Code = 1
		resp.Message = err.Error()
		return nil, err
	}
	return
}
