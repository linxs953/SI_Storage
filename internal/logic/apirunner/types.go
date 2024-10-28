package apirunner

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type StoreAction func(actionKey string, respFields map[string]interface{}) error

type WriteLogFunc func(logType string, eventId string, trigger_node string, message string, err error) error

type FetchDepend func(key string) map[string]interface{}

type ApiExecutor struct {
	Client         http.Client
	RDSClient      *redis.Redis
	Conf           ExecutorConf
	ExecID         string              `json:"execId"`
	Cases          []*SceneConfig      `json:"cases"`
	SceneMap       map[string]string   `json:"sceneMap"`
	ActionMap      map[string]string   `json:"actionMap"`
	PreActionsMap  map[string][]string `json:"preActionsMap"`
	ActionSceneMap map[string]string   `json:"actionSceneMap"`
	Result         sync.Map            `json:"result"`
	LogSet         []RunFlowLog        `json:"logSet"`
	LogChan        chan RunFlowLog     `json:"logChan"`
}

type ExecutorConf struct {
	Timeout int `json:"timeout"`
	Retry   int `json:"retry"`
}

type SceneConfig struct {
	SceneID     string   `json:"sceneId"`
	SceneName   string   `json:"sceneName"`
	Description string   `json:"description"`
	Total       int      `json:"total"`
	Author      string   `json:"author"`
	Timeout     int      `json:"timeout"`
	Retry       int      `json:"retry"`
	Actions     []Action `json:"actions"`
}

type Action struct {
	SceneID      string        `json:"sceneId"`
	ApiID        string        `json:"apiId"`
	ActionID     string        `json:"actionId"`
	ActionName   string        `json:"actionName"`
	CurrentRefer string        `json:"currentRefer"`
	Conf         ActionConf    `json:"conf"`
	StartTime    time.Time     `json:"startTime"`
	FinishTime   time.Time     `json:"finishTime"`
	Duration     int           `json:"duration"`
	Request      ActionRequest `json:"request"`
	Output       ActionOutput  `json:"output"`
	Expect       ActionExpect  `json:"expect"`
	Before       []Hook        `json:"before"`
	After        []Hook        `json:"after"`
}

type ActionExpect struct {
	ActionID  string      `json:"actionId"`
	ApiExpect []ApiExpect `json:"apiExpect"`
	SqlExpect SQLExpect   `json:"sqlExpect"`
}

type ApiExpect struct {
	Type      string        `json:"type"`
	FieldName string        `json:"fieldName"`
	Operation string        `json:"operation"`
	DataType  string        `json:"dataType"`
	Desire    DesireSetting `json:"desire"`
}

type DesireSetting struct {
	DataSource  []DependInject   `json:"dataSource"`
	Extra       string           `json:"extra"`
	DsSpec      []DataSourceSpec `json:"dsSpec"`
	Output      OutputSpec       `json:"output"`
	IsMultiDs   bool             `json:"isMultiDs"`
	Mode        string           `json:"mode"`
	ReferTarget string           `json:"referTarget"`
	ReferType   string           `json:"referType"`
}

type OutputSpec struct {
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

type SQLExpect struct {
	SQLClase string   `json:"sqlClase"`
	Table    string   `json:"table"`
	Verify   []Verify `json:"verify"`
}

type Verify struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
}

type ActionOutput struct {
	ExecID   string                 `json:"execId"`
	SceneID  string                 `json:"sceneId"`
	ActionID string                 `json:"actionId"`
	Value    map[string]interface{} `json:"value"`
}

type ActionConf struct {
	Retry   int `json:"retry"`
	Timeout int `json:"timeout"`
}

type ActionRequest struct {
	Domain     string                 `json:"domain"`
	Path       string                 `json:"path"`
	Method     string                 `json:"method"`
	Headers    map[string]string      `json:"headers"`
	Params     map[string]string      `json:"params"`
	Payload    map[string]interface{} `json:"payload"`
	HasRetry   int                    `json:"hasRetry"`
	Dependency []ActionDepend         `json:"dependency"`
}

type ActionDepend struct {
	Type       string           `json:"type"`
	ActionKey  string           `json:"actionKey"`
	DataKey    string           `json:"dataKey"`
	DataSource []DependInject   `json:"dataSource"`
	IsMultiDs  bool             `json:"isMultiDs"`
	Mode       string           `json:"mode"`
	Extra      string           `json:"extra"`
	DsSpec     []DataSourceSpec `json:"dsSpec"`
	Output     Output           `json:"output"`
	Refer      Refer            `json:"refer"`
}

// 添加到模板的数据源定义
type DataSourceSpec struct {
	FieldName string `json:"fieldName"`
	DependId  string `json:"dependId"`

	// 写入到 extra 里面的字段的数据类型
	DataType string `json:"dataType"`
}

type Output struct {
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

type DependInject struct {
	DependId      string       `json:"dependId"`
	Type          string       `json:"type"`
	DataKey       string       `json:"dataKey"`
	ActionKey     string       `json:"actionKey"`
	SearchCondArr []SearchCond `json:"searchCondArr"`
}

type SearchCond struct {
	// 条件字段
	CondFiled string `json:"condFiled"`
	// 条件值
	CondValue string `json:"condValue"`
	// 条件类型, eq / neq / gt / gte / lt / lte / like / in / nin / nin
	CondOperation string `json:"condOperation"`
}

type Refer struct {
	// 注入的类型, Path / Header / Payload
	Type string `json:"type"`
	// 表达式, header.Authorization / payload.id / path.id
	Target   string `json:"target"`
	DataType string `json:"dataType"`
}

type ApiExecutorContext struct {
	RdsClient *redis.Redis    `json:"rdsClient"`
	ExecID    string          `json:"execId"`
	Result    *sync.Map       `json:"result"`
	LogChan   chan RunFlowLog `json:"logChan"`
	WriteLog  WriteLogFunc    `json:"writeLog"`
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
	RootErr error `json:"rootErr,omitempty"`
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

		sceneMap[scene.SceneID] = scene.Description
		sceneIterMap[scene.SceneID] = ""
		for aidx, action := range scene.Actions {
			scenes[sidx].Actions[aidx].Output.ExecID = execID
			preActionMap[action.ActionID] = strings.Split(sceneIterMap[scene.SceneID], ",")
			if sceneIterMap[scene.SceneID] == "" {
				sceneIterMap[scene.SceneID] = action.ActionID
			} else {
				sceneIterMap[scene.SceneID] += fmt.Sprintf(",%s", action.ActionID)
			}
			sceneActionMap[action.ActionID] = scene.SceneID
			actionMap[action.ActionID] = action.ActionName
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
