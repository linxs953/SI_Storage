package task

import (
	"context"

	"lexa-engine/internal/svc"

	mongo "lexa-engine/internal/model/mongo"
	trl "lexa-engine/internal/model/mongo/task_run_log"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskRunListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type GetTaskRunListResp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []*trl.TaskRunLog `json:"data"`
}

type GetTaskRunListDto struct {
	TaskId string
}

func NewGetTaskRunListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskRunListLogic {
	return &GetTaskRunListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskRunListLogic) GetRunLogList(dto GetTaskRunListDto) (resp *GetTaskRunListResp, err error) {
	resp = &GetTaskRunListResp{
		Code:    0,
		Message: "success",
	}
	murl := mongo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := trl.NewTaskRunLogModel(murl, "lct", "TaskRunLog")
	data, err := mod.FindTaskRunRecord(l.ctx, dto.TaskId)
	if err != nil {
		return
	}
	resp.Data = data
	return
}
