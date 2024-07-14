package sync_task

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ SynctaskModel = (*customSynctaskModel)(nil)

type (
	// SynctaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSynctaskModel.
	SynctaskModel interface {
		synctaskModel
		FindRecord(ctx context.Context) (st *Synctask, err error)
		InsertRecord(ctx context.Context, data *Synctask) (err error)
		UpdateRecord(ctx context.Context, data *Synctask) (*mongo.UpdateResult, error)
	}

	customSynctaskModel struct {
		*defaultSynctaskModel
	}
)

func (csm *customSynctaskModel) FindRecord(ctx context.Context) (st *Synctask, err error) {
	var record Synctask
	if err = csm.conn.FindOne(ctx, &record, bson.M{"state": 0}); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logx.Error("查找同步任务记录失败", err)
			return
		}
		err = nil
		st = nil
		return
	}
	st = &record
	return
}

func (csm *customSynctaskModel) InsertRecord(ctx context.Context, data *Synctask) (err error) {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}
	// 进行中的状态是 0
	data.State = 0

	// 1-同步 api
	data.SyncType = 1

	data.TimeStamp = time.Now().Unix()
	if err = csm.Insert(ctx, data); err != nil {
		logx.Error("插入同步任务记录失败", err)
		return
	}
	return
}

func (csm *customSynctaskModel) UpdateRecord(ctx context.Context, data *Synctask) (*mongo.UpdateResult, error) {
	// result, err := csm.Update(ctx, data)
	result, err := csm.conn.UpdateOne(ctx, bson.M{"_id": data.ID}, bson.M{"$set": data})
	if err != nil {
		logx.Error("更新同步任务记录失败", err)
		return nil, err
	}

	return result, err
}

// NewSynctaskModel returns a model for the mongo.
func NewSynctaskModel(url, db, collection string) SynctaskModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customSynctaskModel{
		defaultSynctaskModel: newDefaultSynctaskModel(conn),
	}
}
