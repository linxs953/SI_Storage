package apirunner

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	// "github.com/zeromicro/go-zero/core/logx"
	"github.com/google/uuid"
)

type StoreAction func(actionKey string, respFields map[string]interface{}) error

type WriteLogFunc func(logType string, eventId string, trigger_node string, message string, err error) error

type FetchDepend func(key string) map[string]interface{}

type ApiExecutor struct {
	Client         http.Client
	Conf           ExecutorConf
	ExecID         string
	Cases          []*SceneConfig
	SceneMap       map[string]string
	ActionMap      map[string]string
	PreActionsMap  map[string][]string
	ActionSceneMap map[string]string
	// mu             sync.RWMutex // 添加读写锁
	// Result         map[string]map[string]interface{}
	Result  sync.Map
	LogSet  []RunFlowLog
	LogChan chan RunFlowLog
}

type ExecutorConf struct {
	Timeout int
	Retry   int
}

type SceneConfig struct {
	SceneID     string
	SceneName   string
	Description string
	Total       int
	Author      string
	Timeout     int
	Retry       int
	Actions     []Action
}

type Action struct {
	SceneID      string
	ApiID        string
	ActionID     string
	ActionName   string
	CurrentRefer string
	Conf         ActionConf
	StartTime    time.Time
	FinishTime   time.Time
	Duration     int
	Request      ActionRequest
	Output       ActionOutput
	Expect       ActionExpect
	Before       []Hook
	After        []Hook
}

type ActionExpect struct {
	ActionID  string
	ApiExpect []ApiExpect
	SqlExpect SQLExpect
}

type ApiExpect struct {
	Type      string
	FieldName string
	Operation string
	DataType  string
	Desire    interface{}
}

type SQLExpect struct {
	SQLClase string
	Table    string
	Verify   []Verify
}

type Verify struct {
	FieldName  string
	FieldValue string
}

type ActionOutput struct {
	ExecID   string
	SceneID  string
	ActionID string

	// 存储action的结果
	Value map[string]interface{}
}

type ActionConf struct {
	Retry   int
	Timeout int
}

type ActionRequest struct {
	Domain     string
	Path       string
	Method     string
	Headers    map[string]string
	Params     map[string]string
	Payload    map[string]interface{}
	HasRetry   int
	Dependency []ActionDepend
}

type ActionDepend struct {
	// 当前依赖的类型, 内部 / 外部
	Type      string
	ActionKey string

	// 表达式,从 response 中读取具体字段的值
	DataKey    string
	DataSource []DependInject

	IsMultiDs bool

	// 1 表示最终的数据类型是 string, 包括常规的string 和序列化对象后的 string
	Mode string

	// 存储最终赋值给字段的模版, 执行的时候把数据源注入, 拼接成真实的数据
	Extra string

	DsSpec []DataSourceSpec

	// 最终的数据值
	Output Output

	Refer Refer
}

// 添加到模板的数据源定义
type DataSourceSpec struct {
	FieldName string
	DependId  string

	// 写入到 extra 里面的字段的数据类型
	DataType string
}

type Output struct {
	Value interface{}
	Type  string
}

type DependInject struct {
	DependId      string
	Type          string
	DataKey       string
	ActionKey     string
	SearchCondArr []SearchCond
}

type SearchCond struct {
	// 条件字段
	CondFiled string
	// 条件值
	CondValue string
	// 条件类型, eq / neq / gt / gte / lt / lte / like / in / nin / nin
	CondOperation string
}

type Refer struct {
	// 注入的类型, Path / Header / Payload
	Type string
	// 表达式, header.Authorization / payload.id / path.id
	Target   string
	DataType string
}

type ApiExecutorContext struct {
	ExecID string
	// Store    StoreAction
	// Fetch    FetchDepend
	Result   *sync.Map
	LogChan  chan RunFlowLog
	WriteLog WriteLogFunc
}

type RunFlowLog struct {
	// scene / action / task
	LogType string `json:"log_type"`

	// type=scene, 存储sceneID
	// type=action, 存储actionID
	// type=task, 存储taskID
	EventId   string `json:"event_id"`
	EventName string `json:"eventName"`

	// 任务实例化ID
	RunId string `json:"run_id"`

	// 日志触发节点
	TriggerNode string `json:"trigger_node"`

	// scene级别字段
	FinishCount  int  `json:"finish_count"`
	SuccessCount int  `json:"success_count"`
	FailCount    int  `json:"fail_count"`
	SceneIsEof   bool `json:"scene_is_eof"`

	// action级别字段
	SceneID        string                   `json:"scene_id"`
	RequestURL     string                   `json:"request_url"`
	RequestMethod  string                   `json:"request_method"`
	RequestHeaders map[string]string        `json:"request_headers"`
	RequestPayload interface{}              `json:"request_payload"`
	RequestDepend  []map[string]interface{} `json:"request_depend"`
	Response       map[string]interface{}   `json:"response"`
	ActionIsEof    bool                     `json:"action_is_eof"`

	// 日志的内容, 如果有错误，拿到Error外层的Message
	Message string `json:"message"`

	// 存储根因
	RootErr error `json:"rootErr;omitempty"`
}

func NewApiExecutor(scenes []*SceneConfig) (*ApiExecutor, error) {
	client := &http.Client{}
	execID := uuid.New().String()
	preActionMap := make(map[string][]string)
	sceneActionMap := make(map[string]string)
	sceneMap := make(map[string]string)
	actionMap := make(map[string]string)
	sceneIterMap := make(map[string]string)
	for sidx, scene := range scenes {
		// var preActions []string
		// var preActionsStr string
		sceneMap[scene.SceneID] = scene.Description
		sceneIterMap[scene.SceneID] = ""
		for aidx, action := range scene.Actions {
			// logx.Errorf("当前Action, %s", action.ActionName)
			scenes[sidx].Actions[aidx].Output.ExecID = execID
			preActionMap[action.ActionID] = strings.Split(sceneIterMap[scene.SceneID], ",")
			if sceneIterMap[scene.SceneID] == "" {
				sceneIterMap[scene.SceneID] = action.ActionID
			} else {
				sceneIterMap[scene.SceneID] += fmt.Sprintf(",%s", action.ActionID)
			}
			sceneActionMap[action.ActionID] = scene.SceneID
			actionMap[action.ActionID] = action.ActionName
			// logx.Errorf("当前Action的前置, %v", preActionMap[action.ActionID])

		}
	}
	return &ApiExecutor{
		Client: *client,
		ExecID: execID,
		Cases:  scenes,
		// Result:         make(map[string]map[string]interface{}),
		Result:         sync.Map{},
		LogSet:         make([]RunFlowLog, 0),
		LogChan:        make(chan RunFlowLog, len(scenes)*25),
		ActionSceneMap: sceneActionMap,
		PreActionsMap:  preActionMap,
		SceneMap:       sceneMap,
		ActionMap:      actionMap,
	}, nil
}
