// Code generated by goctl. DO NOT EDIT.
// goctl 1.7.6
// Source: Storage.proto

package executeservice

import (
	"context"

	"Storage/storage"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	ApifoxConfig               = storage.ApifoxConfig
	CreateSceneConfigRequest   = storage.CreateSceneConfigRequest
	CreateTaskRequest          = storage.CreateTaskRequest
	CreateTestDataRequest      = storage.CreateTestDataRequest
	DeleteInterfaceRequest     = storage.DeleteInterfaceRequest
	DeleteResponse             = storage.DeleteResponse
	DeleteSceneConfigRequest   = storage.DeleteSceneConfigRequest
	DeleteTaskRequest          = storage.DeleteTaskRequest
	DeleteTestDataRequest      = storage.DeleteTestDataRequest
	Dependency                 = storage.Dependency
	Empty                      = storage.Empty
	ExecuteTaskRequest         = storage.ExecuteTaskRequest
	ExecuteTaskResponse        = storage.ExecuteTaskResponse
	Expect                     = storage.Expect
	ExtractConfig              = storage.ExtractConfig
	Extractor                  = storage.Extractor
	GenerateDependencyRequest  = storage.GenerateDependencyRequest
	GenerateDependencyResponse = storage.GenerateDependencyResponse
	GenerateExpectRequest      = storage.GenerateExpectRequest
	GenerateExpectResponse     = storage.GenerateExpectResponse
	GenerateExtractorRequest   = storage.GenerateExtractorRequest
	GenerateExtractorResponse  = storage.GenerateExtractorResponse
	GetInterfaceListResponse   = storage.GetInterfaceListResponse
	GetInterfaceRequest        = storage.GetInterfaceRequest
	GetInterfaceResponse       = storage.GetInterfaceResponse
	GetSceneConfigRequest      = storage.GetSceneConfigRequest
	GetTaskReportListRequest   = storage.GetTaskReportListRequest
	GetTaskRequest             = storage.GetTaskRequest
	GetTestDataRequest         = storage.GetTestDataRequest
	GetTestReportRequest       = storage.GetTestReportRequest
	Header                     = storage.Header
	InterfaceInfo              = storage.InterfaceInfo
	ListSceneConfigsRequest    = storage.ListSceneConfigsRequest
	ListValue                  = storage.ListValue
	MongoConfig                = storage.MongoConfig
	Parameter                  = storage.Parameter
	RelatedApi                 = storage.RelatedApi
	ReportListResponse         = storage.ReportListResponse
	ResponseHeader             = storage.ResponseHeader
	RetrySetting               = storage.RetrySetting
	Scenarios                  = storage.Scenarios
	SceneConfig                = storage.SceneConfig
	SceneConfigListResponse    = storage.SceneConfigListResponse
	SceneConfigResponse        = storage.SceneConfigResponse
	Strategy                   = storage.Strategy
	Struct                     = storage.Struct
	SyncDestination            = storage.SyncDestination
	SyncInterfaceRequest       = storage.SyncInterfaceRequest
	SyncInterfaceResponse      = storage.SyncInterfaceResponse
	SyncSource                 = storage.SyncSource
	Task                       = storage.Task
	TaskAPISpec                = storage.TaskAPISpec
	TaskListResponse           = storage.TaskListResponse
	TaskListResponse_TaskItem  = storage.TaskListResponse_TaskItem
	TaskMeta                   = storage.TaskMeta
	TaskResponse               = storage.TaskResponse
	TaskSyncSpec               = storage.TaskSyncSpec
	TestData                   = storage.TestData
	TestDataListResponse       = storage.TestDataListResponse
	TestDataResponse           = storage.TestDataResponse
	TestReport                 = storage.TestReport
	TestReportResponse         = storage.TestReportResponse
	TimeoutSetting             = storage.TimeoutSetting
	Timestamp                  = storage.Timestamp
	UpdateSceneConfigRequest   = storage.UpdateSceneConfigRequest
	UpdateTaskRequest          = storage.UpdateTaskRequest
	UpdateTestDataRequest      = storage.UpdateTestDataRequest
	Value                      = storage.Value

	ExecuteService interface {
		// 任务执行
		ExecuteTask(ctx context.Context, in *ExecuteTaskRequest, opts ...grpc.CallOption) (*ExecuteTaskResponse, error)
	}

	defaultExecuteService struct {
		cli zrpc.Client
	}
)

func NewExecuteService(cli zrpc.Client) ExecuteService {
	return &defaultExecuteService{
		cli: cli,
	}
}

// 任务执行
func (m *defaultExecuteService) ExecuteTask(ctx context.Context, in *ExecuteTaskRequest, opts ...grpc.CallOption) (*ExecuteTaskResponse, error) {
	client := storage.NewExecuteServiceClient(m.cli.Conn())
	return client.ExecuteTask(ctx, in, opts...)
}
