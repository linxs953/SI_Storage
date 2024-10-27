package task

import (
	"context"
	"fmt"
	"lexa-engine/internal/logic/apirunner"
	mgoutil "lexa-engine/internal/model/mongo"
	mongo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/task_run_log"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Action represents an action with its dependencies, output, and expected results.
type Action struct {
	ActionID     string            `json:"actionId"`
	SearchKey    string            `json:"searchKey"`
	ActionName   string            `json:"actionName"`
	ActionPath   string            `json:"actionPath"`
	ActionMethod string            `json:"actionMethod"`
	RelateId     int               `json:"relateId"`
	EnvKey       string            `json:"envKey"`
	DomainKey    string            `json:"domainKey"`
	Headers      map[string]string `json:"headers"`
	Dependency   []Dependency      `json:"dependency"`
	Output       Output            `json:"output"`
	Expect       Expect            `json:"expect"`
	Retry        int               `json:"retry"`
	Timeout      int               `json:"timeout"`
}

// Dependency represents the dependencies of an action.
type Dependency struct {
	Type      string `json:"type"`
	ActionKey string `json:"actionKey"`
	DataKey   string `json:"dataKey"`
	Refer     Refer  `json:"refer"`
}

// Refer represents references within a dependency.
type Refer struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	DataType string `json:"dataType"`
	// Field    string `json:"field"`
	// Match    string `json:"match"`
	// Location string `json:"location"`
}

// Output represents the output of an action.
type Output struct {
	// Event     string      `json:"event"`
	// EventBody []EventBody `json:"meta"`
	Key string `json:"key"`
}

// Meta represents metadata of an action output.
type EventBody struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
	DataType   string `json:"dataType"`
}

// Expect defines the expected outcomes of an action.
type Expect struct {
	Sql Sql   `json:"sql"`
	Api []Api `json:"api"`
}

// Sql represents SQL related expectations.
type Sql struct {
	Sql    string   `json:"sql"`
	Table  string   `json:"table"`
	Verify []Verify `json:"verify"`
}

// Verify represents verification details for SQL expectations.
type Verify struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// Api represents API related expectations.
type Api struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}

// Data holds the detailed data for an API expectation.
type Data struct {
	Name      string      `json:"name"`
	Operation string      `json:"operation"`
	Type      string      `json:"type"`
	Desire    interface{} `json:"desire"`
}

// Scene represents a scenario including actions, results, and webhook details.
type Scene struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Author      string `json:"author"`
	SceneId     string `json:"sceneId"`
	SearchKey   string `json:"searchKey"`
	EnvKey      string `json:"envKey"`
	// RequestId   string   `json:"requestId"`
	Actions []Action `json:"actions"`
	// State   int      `json:"state"`
	// Result      []Result `json:"result"`
	// WebHook     WebHook  `json:"webHook"`
}

// Result represents the result of an action within a scene.
type Result struct {
	ActionId int    `json:"actionId"`
	ApiId    string `json:"apiId"`
	ApiName  string `json:"apiName"`
	Request  string `json:"request"`
	Duration int    `json:"duration"`
	Error    string `json:"error"`
	Resp     string `json:"resp"`
}

// WebHook represents the webhook details within a scene.
type WebHook struct {
	HookAddr string `json:"hookAddr"`
	HookType string `json:"hookType"`
}

// 定义 kafka 推送事件 struct
type TaskEvent struct {
	EventType string   `json:"eventType"`
	EventMsg  EventMsg `json:"meta"`
}

type EventMsg struct {
	RequestID string
	TaskID    string
	Total     int
	Execute   int
	StartAt   time.Time
	FinishAt  time.Time
	Duration  time.Duration
	State     int
}

type DependencyReference struct {
	Type     string `yaml:"type"`     // payload, path, header之一
	Target   string `yaml:"target"`   // 引用的目标，如header.Authorization
	DataType string `yaml:"dataType"` // 数据类型
}

// Dependency 表示动作之间的依赖关系
type Dependency1 struct {
	Type      string                `yaml:"type"`      // 内部/外部
	ActionKey string                `yaml:"actionKey"` // 依赖的action名称
	DataKey   string                `yaml:"dataKey"`   // 依赖的action输出的key
	Refer     []DependencyReference `yaml:"refer"`     // 引用详情
}

// OutputField 表示动作的输出
type OutputField struct {
	Key   string            `yaml:"key"`   // 输出字段名
	Value map[string]string `yaml:"value"` // 输出的具体值
}

// ExpectationAPIField 期望的API响应字段验证
type ExpectationAPIField struct {
	Type       string      `yaml:"type"`      // field 或 api
	FieldOrAPI string      `yaml:"fieldName"` // 字段名称或api整体
	Operation  string      `yaml:"operation"` // 验证操作，如eq
	DataType   string      `yaml:"dataType"`  // 数据类型
	Desire     interface{} `yaml:"desire"`    // 期望值，可以是多种类型
}

// Expectation 定义了动作执行后的预期结果
type Expectation struct {
	API []ExpectationAPIField `yaml:"api"`
}

// ActionConfig 单个动作的配置
type ActionConfig struct {
	ActionName string        `yaml:"actionName"`
	ApiID      string        `yaml:"apiID"`
	ActionID   string        `yaml:"actionID"`
	Retry      int           `yaml:"retry"`
	Timeout    int           `yaml:"timeout"` // 注意：此处的timeout单位应与yaml文件中一致，为秒
	Dependency []Dependency1 `yaml:"dependency"`
	Output     []OutputField `yaml:"output"`
	Expect     Expectation   `yaml:"expect"`
}

// SceneConfig 定义整个场景的配置
type SceneConfig struct {
	Description string         `yaml:"description"`
	Type        string         `yaml:"type"`
	Author      string         `yaml:"author"`
	SceneID     string         `yaml:"sceneId"`
	Total       int            `yaml:"total"`
	Timeout     int            `yaml:"timeout"` // 场景整体的超时时间
	Actions     []ActionConfig `yaml:"actions"`
}

func NewRunTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunTaskLogic {
	return &RunTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RunTaskLogic) RunTask(req *types.RunTaskDto) (resp *types.RunTaskResp, err error) {
	// 输入验证
	if req == nil {
		return nil, fmt.Errorf("请求参数不能为空")
	}

	// 使用有效的输入值，这里假设 "" 是不合法的输入，应视实际情况调整
	if req.TaskId == "" {
		return nil, fmt.Errorf("任务ID不能为空")
	}

	// 构建执行器
	exector, err := l.buildApiExecutorWithTaskId(req.TaskId)

	if err != nil {
		return nil, err
	}
	// exectorBts, err := json.Marshal(exector.Cases)
	// if err != nil {
	// 	logx.Error(err)
	// }
	// logx.Error(string(exectorBts))

	taskRunning, err := l.checkTaskRunning(req.TaskId)

	// 获取任务运行记录失败
	if err != nil {
		return &types.RunTaskResp{
			Code:    1,
			Message: err.Error(),
			Data: types.RunResp{
				Message: "获取任务运行状态失败",
				RunId:   "",
				State:   0,
			},
		}, err
	}

	// 有运行中的任务,返回运行中的任务
	if taskRunning != nil {
		return &types.RunTaskResp{
			Code:    0,
			Message: "success",
			Data: types.RunResp{
				Message: "存在运行中的任务",
				RunId:   taskRunning.ExecID,
				State:   taskRunning.TaskDetail.TaskState,
			},
		}, nil
	}

	// 没有运行中的任务, 创建任务执行记录
	if _, err := l.CreateTaskRunRecord(req.TaskId, exector.ExecID); err != nil {
		logx.Error(err)
		return nil, err
	}

	exector.Run(context.Background(), req.TaskId, l.svcCtx.RedisClient, l.svcCtx.Config.Database.Mongo)

	return &types.RunTaskResp{
		Code:    0,
		Message: "success",
		Data: types.RunResp{
			Message: "已触发任务",
			RunId:   exector.ExecID,
			State:   0,
		},
	}, nil
}

func (l *RunTaskLogic) CreateTaskRunRecord(taskId string, execId string) (*task_run_log.TaskRunLog, error) {
	murl := mongo.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskInfoMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	mod := task_run_log.NewTaskRunLogModel(murl, "lct", "TaskRunLog")
	taskMeta, err := taskInfoMod.FindByTaskId(context.Background(), taskId)
	if err != nil {
		return nil, err
	}
	data := &task_run_log.TaskRunLog{
		TaskID:  taskId,
		ExecID:  execId,
		LogType: "task",
		TaskDetail: &task_run_log.TaskLog{
			TaskID:         taskId,
			ExecID:         execId,
			TaskName:       taskMeta.TaskName,
			TaskSceneCount: len(taskMeta.Scenes),
			TaskState:      0,
		},
	}
	if err := mod.Insert(context.Background(), data); err != nil {
		return nil, err
	}
	return data, nil
}

func (l *RunTaskLogic) buildApiExecutorWithTaskId(taskId string) (*apirunner.ApiExecutor, error) {
	var executor *apirunner.ApiExecutor
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskInfo, err := taskMod.FindByTaskId(context.Background(), taskId)
	if err != nil {
		return nil, err
	}
	// 转换taskInfo.Scenes类型为 SceneConfig
	scenes := make([]*apirunner.SceneConfig, 0)
	for _, scene := range taskInfo.Scenes {
		sceneConfig := apirunner.SceneConfig{
			SceneName:   scene.SceneName,
			Description: scene.Description,
			Author:      scene.Author,
			SceneID:     scene.SceneId,
			Total:       len(scene.Actions),
			Timeout:     scene.Timeout,
			Retry:       scene.Retry,
		}

		// 构建动作配置
		actions := make([]apirunner.Action, 0)
		for _, action := range scene.Actions {
			// 构建Action dependency
			actionDepends := make([]apirunner.ActionDepend, 0)
			for _, dep := range action.Dependency {
				ad := apirunner.ActionDepend{
					Type:      dep.Type,
					ActionKey: dep.ActionKey,
					DataKey:   dep.DataKey,
					Refer: apirunner.Refer{
						DataType: dep.Refer.DataType,
						Target:   dep.Refer.Target,
						Type:     dep.Refer.Type,
					},
					DataSource: []apirunner.DependInject{},
					DsSpec:     []apirunner.DataSourceSpec{},
					Mode:       dep.Mode,
					Extra:      dep.Extra,
					IsMultiDs:  dep.IsMultiDs,
				}
				for _, dss := range dep.DsSpec {
					ad.DsSpec = append(ad.DsSpec, apirunner.DataSourceSpec{
						DependId:  dss.DependId,
						FieldName: dss.FieldName,
						DataType:  dss.DataType,
					})
				}
				for _, ds := range dep.DataSource {
					searchCondArr := make([]apirunner.SearchCond, 0)
					for _, sc := range ds.SearchCondArr {
						searchCondArr = append(searchCondArr, apirunner.SearchCond{
							CondFiled:     sc.CondFiled,
							CondValue:     sc.CondValue,
							CondOperation: sc.CondOperation,
						})
					}
					// dependId := utils.EncodeToBase36(utils.GenerateId())
					dependId := ds.DependId
					ad.DataSource = append(ad.DataSource, apirunner.DependInject{
						// DependId:      fmt.Sprintf("%s_%s", "DEPEND", strings.ToUpper(dependId)),
						DependId:      dependId,
						Type:          ds.Type,
						DataKey:       ds.DataKey,
						ActionKey:     ds.ActionKey,
						SearchCondArr: searchCondArr,
					})
				}
				actionDepends = append(actionDepends, ad)
			}

			// 根据dependency构建Params
			actionParams := make(map[string]string)

			// 根据dependency 构建 payload
			actionPayload := make(map[string]interface{})

			for _, dep := range actionDepends {
				var valRefer string
				switch dep.Type {
				case "1":
					{
						// 数据源=场景
						valRefer = fmt.Sprintf("$%s.%s", dep.ActionKey, dep.DataKey)
						break
					}
				case "2":
					{
						// 数据源=基础数据
						if dep.ActionKey != "" {
							valRefer = fmt.Sprintf("%s.%s", dep.ActionKey, dep.DataKey)
						} else {
							valRefer = fmt.Sprintf(dep.DataKey)
						}
						break
					}
				case "3":
					{
						// 数据源=自定义值
						valRefer = dep.DataKey
						break
					}
				case "4":
					{
						// 数据源=事件
						valRefer = dep.DataKey
						break
					}
				default:
					{
						break
					}
				}
				// 考虑到字符串是一个对象，在前端配置的时候，传过来的是一个具体参数化的值
				// e.g
				// 1. 传递的值=$scid.acid.data.token, 后续执行读取scid.acid的返回值，解析出token字段进行赋值
				// 2. 传递的值={"goods":"$scid.acid.data.token"}, 执行时读取scid.acid的返回值，解析出token字段的值，给字符串中的某个字段赋值
				if dep.Refer.Type == "params" {
					actionParams[dep.Refer.Target] = valRefer
				}
				if dep.Refer.Type == "payload" {
					actionPayload[dep.Refer.Target] = valRefer
				}

				if dep.Refer.Type == "header" {
					action.Headers[dep.Refer.Target] = valRefer
				}

				if dep.Refer.Type == "path" {
					pathId := dep.Refer.Target
					pathValue := valRefer
					// 替换actionPath 中的参数
					action.ActionPath = strings.Replace(action.ActionPath, fmt.Sprintf("{%s}", pathId), pathValue, -1)
				}
			}

			// 构建Action Expect
			apiExpects := make([]apirunner.ApiExpect, 0)
			for _, expect := range action.Expect.Api {
				apiExpects = append(apiExpects, apirunner.ApiExpect{
					DataType:  expect.Data.Type,
					Desire:    expect.Data.Desire,
					FieldName: expect.Data.Name,
					Operation: expect.Data.Operation,
					Type:      expect.Type,
				})
			}

			sqlVerify := make([]apirunner.Verify, 0)
			for _, expect := range action.Expect.Sql.Verify {
				sqlVerify = append(sqlVerify, apirunner.Verify{
					FieldName:  expect.Field,
					FieldValue: expect.Value,
				})
			}
			sqlExpects := apirunner.SQLExpect{
				SQLClase: action.Expect.Sql.Sql,
				Table:    action.Expect.Sql.Table,
				Verify:   sqlVerify,
			}

			actionConfig := apirunner.Action{
				SceneID:    scene.SceneId,
				ApiID:      fmt.Sprintf("%d", action.RelateId),
				ActionID:   action.ActionID,
				ActionName: action.ActionName,
				Conf: apirunner.ActionConf{
					Retry:   action.Retry,
					Timeout: action.Timeout,
				},
				Request: apirunner.ActionRequest{
					Dependency: actionDepends,
					Domain:     action.DomainKey,
					Method:     action.ActionMethod,
					Path:       action.ActionPath,
					Payload:    actionPayload,
					Params:     actionParams,
					Headers:    action.Headers,
					HasRetry:   0,
				},
				Output: apirunner.ActionOutput{
					ExecID:   "",
					SceneID:  scene.SceneId,
					ActionID: action.ActionID,
					Value:    map[string]interface{}{},
				},
				Expect: apirunner.ActionExpect{
					ApiExpect: apiExpects,
					SqlExpect: sqlExpects,
				},
			}
			actions = append(actions, actionConfig)
		}
		sceneConfig.Actions = actions
		scenes = append(scenes, &sceneConfig)
	}
	executor, err = apirunner.NewApiExecutor(scenes)
	return executor, err
}

func (l *RunTaskLogic) checkTaskRunning(taskId string) (*task_run_log.TaskRunLog, error) {
	murl := mgoutil.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	taskMod := task_run_log.NewTaskRunLogModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskRunLog")
	taskAllRun, err := taskMod.FindAllTaskRecords(l.ctx, taskId, -1, -1)
	if err != nil {
		return nil, err
	}

	if len(taskAllRun) == 0 {
		return nil, nil
	}

	if len(taskAllRun) > 0 {
		for _, taskRun := range taskAllRun {
			if taskRun.TaskDetail.TaskState == 0 {
				return taskRun, nil
			}
		}
	}
	return nil, nil
}
