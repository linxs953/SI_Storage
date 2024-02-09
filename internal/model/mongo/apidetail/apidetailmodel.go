package apidetail

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ ApidetailModel = (*customApidetailModel)(nil)

type (
	// ApidetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customApidetailModel.
	ApidetailModel interface {
		apidetailModel
		FindByApiId(ctx context.Context, apiId int) (*Apidetail, error)
	}

	customApidetailModel struct {
		*defaultApidetailModel
	}
)

func (cm *customApidetailModel) FindByApiId(ctx context.Context, apiId int) (*Apidetail, error) {
	var (
		err    error
		detail Apidetail
	)
	err = cm.conn.FindOne(ctx, &detail, bson.M{"apiId": apiId})
	return &detail, err
}

func (cm *customApidetailModel) Insert(ctx context.Context, data *Apidetail) error {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}
	_, err := cm.FindByApiId(ctx, data.ApiId)
	switch err {
	case nil:
		logx.Error(fmt.Sprintf("存在 apid = [%v] 的记录", data.ApiId))
		logx.Info(fmt.Sprintf("开始更新apid = [%v] 记录.....", data.ApiId))
		_, err := cm.Update(ctx, data)
		if err != nil {
			logx.Error(fmt.Sprintf("更新apid = [%v] 记录失败", data.ApiId))
			return err
		}
		logx.Info(fmt.Sprintf("更新apid = [%v] 记录成功", data.ApiId))
		return nil
	case mon.ErrNotFound:
		_, err = cm.conn.InsertOne(ctx, data)
		if err != nil {
			logx.Error(err)
			return err
		}
		return nil
	default:
		logx.Error(err)
		return err
	}

}

// NewApidetailModel returns a model for the mongo.
func NewApidetailModel(url, db, collection string) ApidetailModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customApidetailModel{
		defaultApidetailModel: newDefaultApidetailModel(conn),
	}
}
