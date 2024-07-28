package syncTask

import (
	"context"
	mongo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sync_task"
	"lexa-engine/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type NewSyncRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNewSyncRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NewSyncRecordLogic {
	return &NewSyncRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NewSyncRecordLogic) NewSyncRecord() error {
	var record sync_task.Synctask
	murl := mongo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)

	mod := sync_task.NewSynctaskModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SyncTask")

	if err := mod.InsertRecord(context.Background(), &record); err != nil {
		logx.Error("创建api 同步记录失败")
		logx.Error(err)
		return err
	}

	return nil
}

func (l *NewSyncRecordLogic) FindSyncRecord() (*sync_task.Synctask, error) {
	murl := mongo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	mod := sync_task.NewSynctaskModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SyncTask")
	record, err := mod.FindRecord(context.Background())
	if err != nil {
		return nil, err
	}

	return record, err
}
