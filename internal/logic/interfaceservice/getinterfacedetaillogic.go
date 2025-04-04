package interfaceservicelogic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Storage/internal/errors"
	"Storage/internal/model/api"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInterfaceDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInterfaceDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInterfaceDetailLogic {
	return &GetInterfaceDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetInterfaceDetailLogic) GetInterfaceDetail(in *storage.GetInterfaceRequest) (*storage.GetInterfaceResponse, error) {
	// 入参校验
	if in.InterfaceId == "" {
		return nil, errors.NewWithError(
			fmt.Errorf("interfaceId不能为空"),
			errors.GetInterfaceDetailError,
		)
	}

	// 日志记录
	l.Logger.Infof("开始获取接口详情: interfaceId=%s", in.InterfaceId)

	apiModel := api.NewApiModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, api.ApiCollectionName)
	apiDetail, err := apiModel.FindOneByApiID(l.ctx, in.InterfaceId)
	if err != nil {
		return nil, errors.NewWithError(err, errors.GetInterfaceDetailError).WithDetails("获取接口详情失败", err)
	}

	// 处理未找到的情况
	if apiDetail == nil {
		return nil, errors.NewWithError(
			fmt.Errorf("接口 %s 不存在", in.InterfaceId),
			errors.GetInterfaceDetailError,
		)
	}

	// RawData 转换优化
	rawData := ""
	if apiDetail.RawData != nil {
		bts, err := json.MarshalIndent(apiDetail.RawData, "", "  ")
		if err != nil {
			l.Logger.Errorf("RawData 序列化失败: %v", err)
			rawData = "{}"
		} else {
			rawData = string(bts)
		}
	}

	// 参数转换优化
	headers := make([]*storage.Header, 0, len(apiDetail.Headers))
	for _, h := range apiDetail.Headers {
		headers = append(headers, &storage.Header{
			Name:  h.Name,
			Value: h.Value,
		})
	}

	parameters := make([]*storage.Parameter, 0, len(apiDetail.Parameters))
	for _, p := range apiDetail.Parameters {
		parameters = append(parameters, &storage.Parameter{
			Name:  p.Name,
			Value: "", // 这个在组装apipipeline的时候会被替换掉，所以返回空字符串没什么关系
		})
	}

	// 日志记录处理结果
	l.Logger.Infof("成功获取接口详情: interfaceId=%s, apiId=%s", in.InterfaceId, apiDetail.ApiID)

	return &storage.GetInterfaceResponse{
		Header: &storage.ResponseHeader{
			Code: int64(errors.Success),
		},
		Detail: &storage.InterfaceInfo{
			ApiId:       apiDetail.ApiID,
			Name:        apiDetail.Name,
			Method:      apiDetail.Method,
			Path:        apiDetail.Path,
			Description: apiDetail.Description,
			Headers:     headers,
			Parameters:  parameters,
			RawData:     rawData,
			ProjectId:   apiDetail.ProjectID,
			TaskId:      apiDetail.TaskID,
			CreateAt:    apiDetail.CreateAt.Format(time.RFC3339),
			UpdateAt:    apiDetail.UpdateAt.Format(time.RFC3339),
		},
	}, nil
}
