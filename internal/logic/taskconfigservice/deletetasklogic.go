package taskconfigservicelogic

import (
	"context"

	"Storage/internal/errors"
	model "Storage/internal/model/task"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskLogic {
	return &DeleteTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTaskLogic) DeleteTask(in *storage.DeleteTaskRequest) (*storage.DeleteResponse, error) {
	dresponse := &storage.DeleteResponse{
		Header: &storage.ResponseHeader{
			Code: int64(errors.Success),
		},
	}
	var affectedRows int64
	taskModel := model.NewTaskModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, model.TaskCollectionName)
	affectedRows, err := taskModel.DeleteOneByTaskID(l.ctx, in.TaskId)
	if err != nil {
		logx.Error(err)
		dresponse.Header.Code = int64(errors.DeleteMgoRecordError)
		dresponse.Header.Message = errors.NewWithError(err, errors.DeleteMgoRecordError).WithDetails("删除任务记录错误", err).GetMessage()
	}
	dresponse.Header.Message = "删除任务成功"
	dresponse.AffectedRows = affectedRows
	return dresponse, nil
}
