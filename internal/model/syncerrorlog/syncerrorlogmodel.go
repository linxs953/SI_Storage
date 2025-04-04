package syncerrorlog

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const SyncErrorLogCollectionName = "sync_error_log" // Collection name for sync error logs

type SyncErrorLogModel interface {
	InsertErrorLog(ctx context.Context, data *SyncErrorLog) error
	FindByTaskID(ctx context.Context, taskId string) ([]*SyncErrorLog, error)
	FindByProjectID(ctx context.Context, projectId string) ([]*SyncErrorLog, error)
	FindByTimeRange(ctx context.Context, start, end time.Time) ([]*SyncErrorLog, error)
}

type defaultSyncErrorLogModel struct {
	conn *mon.Model
}

func NewSyncErrorLogModel(url, db, collection string) SyncErrorLogModel {
	conn := mon.MustNewModel(url, db, collection)
	return &defaultSyncErrorLogModel{
		conn: conn,
	}
}

func (m *defaultSyncErrorLogModel) InsertErrorLog(ctx context.Context, data *SyncErrorLog) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
	}

	data.CreateAt = time.Now()

	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultSyncErrorLogModel) FindByTaskID(ctx context.Context, taskId string) ([]*SyncErrorLog, error) {
	var logs []*SyncErrorLog
	err := m.conn.Find(ctx, bson.M{"taskId": taskId}, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (m *defaultSyncErrorLogModel) FindByProjectID(ctx context.Context, projectId string) ([]*SyncErrorLog, error) {
	var logs []*SyncErrorLog
	err := m.conn.Find(ctx, bson.M{"projectId": projectId}, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (m *defaultSyncErrorLogModel) FindByTimeRange(ctx context.Context, start, end time.Time) ([]*SyncErrorLog, error) {
	var logs []*SyncErrorLog
	filter := bson.M{
		"createAt": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	err := m.conn.Find(ctx, filter, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
