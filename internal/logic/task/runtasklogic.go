package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/yaml.v2"

	"lexa-engine/internal/logic/apirunner"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
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

	exector, err := buildApiExecutor("./demo.yaml")
	if err != nil {
		return nil, err
	}
	exector.Run(context.Background(), l.svcCtx.RedisClient)
	ab, _ := json.Marshal(exector.Cases[0].Actions)
	logx.Error(string(ab))
	return &types.RunTaskResp{
		Code:    0,
		Message: "success",
		Data: types.RunResp{
			Message: "已触发任务",
			RunId:   exector.ExecID,
			State:   0,
		},
	}, nil

	// 尝试解析任务
	// var scene *Scene
	// if scene, err = unMarshalTask(); err != nil {
	// 	// 增加更详细的错误日志
	// 	logx.Errorf("解析任务失败, req.TaskID: %v, err: %v", req.TaskId, err)
	// 	return
	// }

	// // 启动任务，并捕获可能的异常
	// if err = l.StartTask(scene); err != nil {
	// 	// 记录启动任务失败的日志
	// 	logx.Errorf("启动任务失败, scene.ID: %v, err: %v", scene.RequestId, err)
	// 	return
	// }

	// return &types.RunTaskResp{
	// 	Code:    0,
	// 	Message: "success",
	// 	Data: types.RunResp{
	// 		Message:  "已触发任务",
	// 		RunId:    scene.RequestId,
	// 		TaskId:   scene.SceneId,
	// 		TaskSpec: scene,
	// 		State:    0,
	// 	},
	// }, nil
}

func unMarshalTask() (*Scene, error) {
	jsonString := `{"description":"创建订单","type":"1","author":"linxs","sceneId":"100300424","requestId":"实例化id","actions":[{"actionName":"账号密码登录","relateId":100300424,"dependency":[{"type":"upstream","eventName":"订阅的事件名称","refer":[{"field":"$eventData.$token","match":"$auth","location":"payload / path / query / header"}]}],"output":[{"event":"action finish 事件","meta":[{"fieldName":"response字段名","fieldValue":"response字段值","dataType":"字段类型"}]}],"expect":{"sql":{"sql":"select * from psu_order","table":"psu_order","verify":[{"field":"表字段名, $table_alias.$field","value":"字段值"}]},"api":[{"type":"field / api","data":{"name":"statusCode","operation":"eq","type":"int","desire":0}},{"type":"field","data":{"name":"$.msg","operation":"eq","type":"string","desire":"操作成功"}}]}}]}`
	var scene Scene
	err := json.Unmarshal([]byte(jsonString), &scene)
	if err != nil {
		logx.Error("Error:", err)
		return nil, err
	}
	return &scene, nil
}

func buildApiExecutor(filename string) (*apirunner.ApiExecutor, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		logx.Errorf("Error reading YAML file: %v", err)
	}

	var sceneConfig Scene
	err = yaml.Unmarshal(yamlFile, &sceneConfig)
	if err != nil {
		logx.Errorf("Error unmarshaling YAML: %v", err)
		return nil, err
	}

	var scenes []*apirunner.SceneConfig
	var actions []apirunner.Action

	// 通过action sk 找到 scene sk
	actionSceneMap := make(map[string]string)

	// 通过action sk 找到 当前场景下的所有前置actionid
	preActionsMap := make(map[string]string)

	// 通过scene sk 找到sceneid
	sceneMap := make(map[string]string)

	// 通过action sk 找到 actionid
	actionMap := make(map[string]string)

	sceneId := uuid.New().String()

	sceneMap[sceneConfig.SearchKey] = sceneId
	canUseScene := ""
	for _, a := range sceneConfig.Actions {
		actionId := uuid.New().String()
		logx.Errorf("searchkey, %s", a.SearchKey)
		actionMap[a.SearchKey] = actionId
		preActionsMap[a.SearchKey] = canUseScene
		actionSceneMap[a.SearchKey] = sceneConfig.SearchKey

		// 处理 action headers
		apiheaders := make(map[string]string)
		for hn, hv := range a.Headers {
			apiheaders[hn] = hv
		}
		for _, depend := range a.Dependency {
			if depend.Refer.Type != "header" {
				continue
			}
			if depend.Type == "1" {
				apiheaders[depend.Refer.Target] = fmt.Sprintf("Bearer %s.%s", depend.ActionKey, depend.DataKey)
			}
			if depend.Type == "2" {
				apiheaders[depend.Refer.Target] = fmt.Sprintf("Bearer %s.%s", depend.ActionKey, depend.DataKey)
			}

			if depend.Type == "3" {
				apiheaders[depend.Refer.Target] = depend.DataKey
			}
		}

		// 处理 action path
		actionPath := a.ActionPath
		pathParams := regexp.MustCompile(`\{.*?\}`).FindAllString(actionPath, -1)
		if len(pathParams) > 0 {
			// for _, param := range pathParams {
			for _, depend := range a.Dependency {
				if depend.Refer.Type != "path" {
					continue
				}

				if depend.Type == "1" {
					for _, param := range pathParams {
						pathKey := strings.Split(depend.Refer.Target, ".")
						if len(pathKey) != 2 {
							logx.Errorf("填充url path表达式时，depenency个是错误，%s", depend.Refer.Target)
							continue
						}
						if strings.Contains(param, pathKey[1]) {
							actionPath = strings.ReplaceAll(actionPath, param, fmt.Sprintf("%s.%s", depend.ActionKey, depend.DataKey))
						}
						break
					}

				}

				if depend.Type == "2" {
					for _, param := range pathParams {
						pathKey := strings.Split(depend.Refer.Target, ".")
						if len(pathKey) != 2 {
							logx.Errorf("填充url path表达式时，depenency个是错误，%s", depend.Refer.Target)
							continue
						}
						if strings.Contains(param, pathKey[1]) {
							actionPath = strings.ReplaceAll(actionPath, param, fmt.Sprintf("%s.%s", depend.ActionKey, depend.DataKey))
							logx.Errorf("new path : %s, %s, %s", param, depend.ActionKey, depend.DataKey)
							break
						}
					}
				}

				if depend.Type == "3" {
					for _, param := range pathParams {
						pathKey := strings.Split(depend.Refer.Target, ".")
						if len(pathKey) != 2 {
							logx.Errorf("填充url path表达式时，depenency个是错误，%s", depend.Refer.Target)
							continue
						}
						if strings.Contains(param, pathKey[1]) {
							actionPath = strings.ReplaceAll(actionPath, param, depend.DataKey)
						}
						break
					}
				}
			}
			// }
		}

		// 处理 action query
		apiparams := make(map[string]string)
		for _, depend := range a.Dependency {
			if depend.Refer.Type != "query" {
				continue
			}
			queryParts := strings.Split(depend.Refer.Target, ".")
			if len(queryParts) < 2 || len(queryParts) > 2 {
				logx.Error("depend.refer.target格式错误, 示例: query.$field")
				continue
			}
			queryName := queryParts[1]

			if depend.Type == "1" {
				apiparams[queryName] = fmt.Sprintf("%s.%s", depend.ActionKey, depend.DataKey)
			}
			if depend.Type == "2" {
				apiparams[queryName] = fmt.Sprintf("%s.%s", depend.ActionKey, depend.DataKey)
			}
			if depend.Type == "3" {
				apiparams[queryName] = depend.DataKey
			}
		}

		// 处理 action payload
		payload := make(map[string]interface{})
		// post / put 需要去设置payload
		if strings.Contains("POSTPUT", a.ActionMethod) {
			for _, pl := range a.Dependency {
				if pl.Refer.Type != "payload" {
					continue
				}
				key := strings.Split(pl.Refer.Target, ".")[1]
				if pl.Type == "1" {
					payload[key] = fmt.Sprintf("%s.%s", pl.ActionKey, pl.DataKey)
				}
				if pl.Type == "2" {
					payload[key] = fmt.Sprintf("%s.%s", pl.ActionKey, pl.DataKey)
				}
				if pl.Type == "3" {
					payload[key] = pl.DataKey
				}
			}
		}

		// 处理 action 预期结果
		expects := apirunner.ActionExpect{
			ActionID: a.ActionID,
		}
		for _, e := range a.Expect.Api {
			expect := apirunner.ApiExpect{
				Type:      e.Type,
				FieldName: e.Data.Name,
				Operation: e.Data.Operation,
				DataType:  e.Data.Type,
				Desire:    e.Data.Desire,
			}
			expects.ApiExpect = append(expects.ApiExpect, expect)
		}

		// 处理 action 输出
		output := apirunner.ActionOutput{
			ActionID: a.ActionID,
			Key:      fmt.Sprintf("%s.%s", sceneId, a.ActionID),
		}

		action := apirunner.Action{
			ActionID: actionId,
			SceneID:  sceneId,

			CurrentRefer: fmt.Sprintf("%s.%s", sceneConfig.SearchKey, a.SearchKey),
			ApiID:        fmt.Sprintf("%v", a.RelateId),
			ActionName:   a.ActionName,
			Conf: apirunner.ActionConf{
				Timeout: a.Timeout,
				Retry:   a.Retry,
			},
			Request: apirunner.ActionRequest{
				Path:    actionPath,
				Method:  a.ActionMethod,
				Headers: apiheaders,
				Params:  apiparams,
				Payload: payload,
				Domain:  a.DomainKey,
			},
			Expect: expects,
			Output: output,
		}
		actions = append(actions, action)
		if canUseScene == "" {
			canUseScene = actionId
		} else {
			canUseScene = fmt.Sprintf("%s,%s", canUseScene, actionId)
		}
	}
	scene := &apirunner.SceneConfig{
		Description: sceneConfig.Description,
		Total:       len(sceneConfig.Actions),
		Author:      sceneConfig.Author,
		Timeout:     30,
		Actions:     actions,
		SceneID:     sceneId,
	}
	scenes = append(scenes, scene)
	executor, err := apirunner.NewApiExecutor(scenes)
	if err != nil {
		return nil, err
	}
	executor.SceneMap = sceneMap
	executor.ActionSceneMap = actionSceneMap
	executor.PreActionsMap = preActionsMap
	executor.ActionMap = actionMap
	return executor, nil
}

// 开始执行任务
func (l *RunTaskLogic) StartTask(task *Scene) (err error) {
	taskRunID := uuid.New().String()
	taskStartEvent := &TaskEvent{
		EventMsg: EventMsg{
			RequestID: taskRunID,
			TaskID:    task.SceneId,
			Total:     len(task.Actions),
			Execute:   0,
			State:     0,
			StartAt:   time.Now(),
			FinishAt:  time.Time{},
		},
		EventType: "task_start",
	}
	eventBytes, err := json.Marshal(taskStartEvent)
	if err != nil {
		return err
	}
	if err = l.svcCtx.TaskPushClient.Push(string(eventBytes)); err != nil {
		return err
	}
	// for i, action := range task.Actions {

	// }
	taskFinishEvent := &TaskEvent{
		EventMsg: EventMsg{
			RequestID: taskRunID,
			TaskID:    task.SceneId,
			State:     1,
			Duration:  100.0,
			StartAt:   time.Now(),
			FinishAt:  time.Now(),
		},
		EventType: "task_finish",
	}
	finishEventBytes, err := json.Marshal(taskFinishEvent)
	if err != nil {
		return err
	}

	if err = l.svcCtx.TaskPushClient.Push(string(finishEventBytes)); err != nil {
		return err
	}
	return
}
