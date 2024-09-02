package apirunner

import (
	"net/http"
	"sync"
	"time"

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
	PreActionsMap  map[string]string
	ActionSceneMap map[string]string
	mu             sync.RWMutex // 添加读写锁
	Result         map[string]map[string]interface{}
	LogSet         []RunFlowLog
}

type ExecutorConf struct {
	Timeout int
	Retry   int
}

type SceneConfig struct {
	SceneID     string
	Description string
	Total       int
	Author      string
	Timeout     int
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
	SqlExpect []SQLExpect
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
	ActionID string
	// 标识一个 Action 的输出唯一码
	Key string
	// 标识输出结果, 把 response 解包
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
	DataKey string
	Refer   Refer
}

type Refer struct {
	// 注入的类型, Path / Header / Payload
	Type string
	// 表达式, header.Authorization / payload.id / path.id
	Target   string
	DataType string
}

type ApiExecutorContext struct {
	ExecID   string
	Store    StoreAction
	Fetch    FetchDepend
	WriteLog WriteLogFunc
}

type RunFlowLog struct {
	// scene / action / task
	LogType string `json:"log_type"`

	// type=scene, 存储sceneID
	// type=action, 存储actionID
	// type=task, 存储taskID
	EventId string `json:"event_id"`

	// 任务实例化ID
	RunId string `json:"run_id"`

	// 日志触发节点
	TriggerNode string `json:"trigger_node"`

	// 日志的内容, 如果有错误，拿到Error外层的Message
	Message string `json:"message"`

	// 存储根因
	RootErr error `json:"rootErr;omitempty"`
}

func NewApiExecutor(scenes []*SceneConfig) (*ApiExecutor, error) {
	client := &http.Client{}
	return &ApiExecutor{
		Client: *client,
		ExecID: uuid.New().String(),
		Cases:  scenes,
		Result: make(map[string]map[string]interface{}),
		LogSet: make([]RunFlowLog, 0),
	}, nil
}
