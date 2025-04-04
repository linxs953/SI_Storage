package executeservicelogic

import (
	"context"
	"net/http"
	"strconv"

	"Storage/internal/errors"
	"Storage/internal/logic/pipelines"
	"Storage/internal/logic/tools"
	"Storage/internal/logic/workflows/core"
	model "Storage/internal/model/task"
	"Storage/internal/svc"
	"Storage/storage"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// parsePort 将端口字符串转换为整数
func parsePort(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 27017 // 默认 MongoDB 端口
	}
	return port
}

func NewExecuteTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteTaskLogic {
	return &ExecuteTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务执行
func (l *ExecuteTaskLogic) ExecuteTask(in *storage.ExecuteTaskRequest) (*storage.ExecuteTaskResponse, error) {
	// todo: add your logic here and delete this line
	// 构建 ApiFoxSyncPipeline 实例

	taskModel := model.NewTaskModel(l.svcCtx.GetMongoURI(), l.svcCtx.Config.Database.Mongo.UseDb, model.TaskCollectionName)
	task, err := taskModel.FindOneByTaskID(l.ctx, in.TaskId)
	if err != nil {
		return &storage.ExecuteTaskResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InternalError),
				Message: "获取任务信息失败: " + err.Error(),
			},
		}, nil
	}
	if task == nil {
		return &storage.ExecuteTaskResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.DBNotFound),
				Message: "任务不存在",
			},
		}, nil
	}

	if !task.Enable {
		return &storage.ExecuteTaskResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InvalidParameter),
				Message: "任务未启用",
			},
		}, nil
	}

	if task.Type != int32(1) {
		return &storage.ExecuteTaskResponse{
			Header: &storage.ResponseHeader{
				Code:    int64(errors.InvalidParameter),
				Message: "不是同步类型任务",
			},
		}, nil
	}

	// 遍历所有数据源
	for _, source := range task.SyncSpec.Source {
		// 构建 MongoDB 配置列表
		var mongoConfigs []tools.MongoConfig
		for _, dest := range task.SyncSpec.Destination {
			if dest.DestType == "mongodb" {
				mongoConfigs = append(mongoConfigs, tools.MongoConfig{
					MongoHost:   dest.MongoConfig.Host,
					MongoPort:   parsePort(dest.MongoConfig.Port),
					MongoUser:   dest.MongoConfig.Username,
					MongoPasswd: dest.MongoConfig.Password,
					UseDb:       dest.MongoConfig.Dbname[0], // 使用第一个数据库
				})
			}
		}

		// 初始化 pipeline
		syncPipeline := &pipelines.ApiFoxSyncPipeline{
			BasePipeline: &core.BasePipeline{},
			Config: pipelines.ApiFoxSyncConfig{
				ProjectID:   source.Apifox.ProjectId,
				SharedDocID: source.Apifox.ProjectId,
				Username:    source.Apifox.Username,
				Password:    source.Apifox.Password,
				Mongo:       mongoConfigs,
			},
			Client:  &pipelines.ApiClient{Client: &http.Client{}},
			BaseURL: source.Apifox.Base,
		}

		// 执行 ApiFox 同步
		go func(pipeline *pipelines.ApiFoxSyncPipeline) {
			err := pipeline.Execute(l.ctx)
			if err != nil {
				logx.Errorf("执行 ApiFox 同步失败: %v", err)
			}
		}(syncPipeline)
	}

	return &storage.ExecuteTaskResponse{
		Header: &storage.ResponseHeader{
			Code:    int64(errors.Success),
			Message: "ApiFox 同步执行成功",
		},
	}, nil
}
