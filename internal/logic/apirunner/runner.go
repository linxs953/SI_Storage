package apirunner

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"lexa-engine/internal/config"
	mongo "lexa-engine/internal/model/mongo"

	"lexa-engine/internal/model/mongo/task_run_log"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func (runner *ApiExecutor) Initialize(rdsClient *redis.Redis) error {
	if err := runner.parametrization(rdsClient); err != nil {
		return err
	}
	return nil
}

// 处理每个Action的参数化
func (runner *ApiExecutor) parametrization(rdsClient *redis.Redis) error {
	var err error

	for _, scene := range runner.Cases {
		for _, action := range scene.Actions {
			for _, depend := range action.Request.Dependency {
				if err := runner.filledHeader(depend, &action, rdsClient); err != nil {
					return errors.Wrap(err, "headers参数化失败")
				}
				if err := runner.filledPayload(depend, &action, rdsClient); err != nil {
					return errors.Wrap(err, "payload参数化失败")
				}
				if err := runner.filledPath(depend, &action, rdsClient); err != nil {
					return errors.Wrap(err, "path参数化失败")
				}
				if err := runner.filledQuery(depend, &action, rdsClient); err != nil {
					return errors.Wrap(err, "query参数化失败")
				}
			}
		}
	}
	return err
}

func (runner *ApiExecutor) filledHeader(depend ActionDepend, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "headers" {
		if action.Request.Headers == nil {
			action.Request.Headers = make(map[string]string)
		}
		if err = runner.handleActionRefer(depend, action, rdsClient, "headers"); err != nil {
			return err
		}
	}
	logx.Error(action.Request.Headers)
	return nil
}

func (runner *ApiExecutor) filledPayload(depend ActionDepend, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "payload" {
		if err = runner.handleActionRefer(depend, action, rdsClient, "payload"); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ApiExecutor) filledPath(depend ActionDepend, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "path" {
		if err = runner.handleActionRefer(depend, action, rdsClient, "path"); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ApiExecutor) filledQuery(depend ActionDepend, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "params" {
		if err = runner.handleActionRefer(depend, action, rdsClient, "query"); err != nil {
			return err
		}
	}
	return nil
}

// 处理数据源的获取
func (runner *ApiExecutor) handleActionRefer(depend ActionDepend, action *Action, rdsClient *redis.Redis, referType string) (err error) {
	setPayloadData := func(payload map[string]interface{}, key string, value interface{}) (map[string]interface{}, error) {
		parts := strings.Split(key, ".")
		data := payload
		for idx, part := range parts {
			if idx == len(parts)-1 {
				data[part] = value
			} else {
				if next, ok := data[part]; ok {
					if nextMap, ok := next.(map[string]interface{}); ok {
						data = nextMap
					} else {
						return nil, fmt.Errorf("key '%s' already exists but is not a map", key)
					}
				} else {
					return nil, fmt.Errorf("key '%s' not exist", key)

				}
			}
		}
		newPayload := payload
		return newPayload, nil
	}

	if depend.Type == "1" {
		// 数据源=场景，验证依赖引用关系是否合法
		err = runner.handleSceneRefer(depend, action)
		if err != nil {
			return err
		}
	}

	if depend.Type == "2" {
		// 数据源=基础数据(Redis)，获取redis数据进行填充
		val, err := runner.handleBasicRefer(rdsClient, depend)
		if err != nil {
			return err
		}

		switch referType {
		case "headers":
			{
				action.Request.Headers[depend.Refer.Target] = val
				break
			}
		case "payload":
			{
				newPayload, err := setPayloadData(action.Request.Payload, depend.Refer.Target, val)
				if err != nil {
					return err
				}
				action.Request.Payload = newPayload
				break
			}
		case "path":
			{
				action.Request.Path = strings.Replace(action.Request.Path, depend.Refer.Target, val, -1)
				break
			}
		case "query":
			{
				action.Request.Params[depend.Refer.Target] = val
				break
			}
		}
	}

	if depend.Type == "3" {
		// 数据源=自定义, 直接覆盖原有值
		switch referType {
		case "headers":
			{
				action.Request.Headers[depend.Refer.Target] = depend.DataKey
				break
			}
		case "payload":
			{
				newPayload, err := setPayloadData(action.Request.Payload, depend.Refer.Target, depend.DataKey)
				if err != nil {
					return err
				}
				action.Request.Payload = newPayload
				break
			}
		case "path":
			{
				action.Request.Path = strings.Replace(action.Request.Path, depend.Refer.Target, depend.DataKey, -1)
				break
			}
		case "query":
			{
				action.Request.Params[depend.Refer.Target] = depend.DataKey
				break
			}
		}
	}

	if depend.Type == "4" {
		// 数据源=事件
		_, err = runner.handleEventRefer(depend)
		if err != nil {
			return err
		}
	}
	return nil
}

func (runner *ApiExecutor) handleSceneRefer(depend ActionDepend, action *Action) (err error) {
	// 在任务创建或者编辑之后，场景依赖关系已经生成，这个方法主要是验证引用关系是否正常
	// 1. 场景引用表达式： [sceneInstanceId].[actionInstanceId]
	// 2. dataKey照旧，不做处理
	if depend.Type != "1" {
		return errors.New("依赖的数据源不是场景")
	}

	relateData := strings.Split(depend.ActionKey, ".")

	if len(relateData) == 0 || len(relateData) == 1 {
		return errors.New("ActionKey invaild")
	}

	relateScene := relateData[0]
	if relateScene == "" {
		return errors.New("依赖关联的场景为空")
	}

	relateAction := relateData[1]
	if relateAction == "" {
		return errors.New("依赖关联的Action为空")
	}

	// 检测对应的scene 和 action是否存在
	if _, ok := runner.SceneMap[relateScene]; !ok {
		return errors.New("依赖关联的场景不存在")
	}

	if _, ok := runner.ActionMap[relateAction]; !ok {
		return errors.New("依赖关联的Action不存在")
	}

	if runner.ActionSceneMap[relateAction] == runner.ActionSceneMap[action.ActionID] {
		// 引用的action和当前action同属同个场景
		preAction := runner.PreActionsMap[action.ActionID]
		for pidx, pre := range preAction {
			if pre != relateAction && pidx == len(preAction)-1 {
				return errors.New("依赖Action 不在当前Action的前置列表中")
			}
		}
	}

	// 引用关系验证成功
	return
}

func (runner *ApiExecutor) handleBasicRefer(rdsClient *redis.Redis, depend ActionDepend) (value string, err error) {
	// 数据源不是基础数据，不处理
	if depend.Type != "2" {
		return
	}
	logx.Error(depend)
	value, err = runner.fetchDataWithRedis(rdsClient, "list", depend.ActionKey, depend.DataKey)
	return
}

func (runner *ApiExecutor) fetchDataWithRedis(rdsClient *redis.Redis, resultType, key string, dataKey string) (string, error) {
	/*
		仅支持获取list和string类型的基础数据
		1. 根据dataKey获取数据，第一个字段是数字，获取list类型数据
		2. 不是数字，直接通过key获取string类型数据
		3. 遍历剩下的dataKey引用字段
		   - 把1和2获取的数据解析成map
		   - 根据dataKey引用字段，从map中获取对应的值
	*/
	if key == "" {
		// 键名为空，直接返回
		return "", errors.New("redis key为空")
	}

	if dataKey == "" {
		// 说明不是数组，直接按字符串获取
		value, err := rdsClient.GetCtx(context.Background(), key)
		return value, err
	}

	var value interface{}

	dataKeyParts := strings.Split(dataKey, ".")

	// 判断第一个元素是否是数字
	if i, err := strconv.Atoi(dataKeyParts[0]); err == nil {
		// 是数字，说明是数组，直接取数组中的值
		value, err = rdsClient.Lindex(key, int64(i))
		if err != nil {
			return "", err
		}
	} else {
		value, err = rdsClient.Get(key)
	}

	// 尝试转换value成map[string]string
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(value.(string)), &data); err != nil {
		return "", err
	}

	// 遍历dataParts，从data中获取对应的值
	for idx, key := range dataKeyParts {
		if idx == 0 {
			continue
		}
		if _, ok := data[key]; !ok {
			return "", errors.New("数据不存在")
		}
		newDataMap := make(map[string]interface{})
		if idx < len(dataKeyParts)-1 {
			// 需要继续提取，需要继续构造map
			if err := json.Unmarshal([]byte(data[key].(string)), &newDataMap); err != nil {
				return "", err
			}
			data = newDataMap
			continue
		}
		// 能运行到这里，表明是最后一个key
		value = data[key]
	}

	return value.(string), nil
}

func (runner *ApiExecutor) handleEventRefer(depend ActionDepend) (value string, err error) {
	if depend.Type != "4" {
		return
	}
	return
}

func (runner *ApiExecutor) Run(ctx context.Context, rdsClient *redis.Redis, mgoConfig config.MongoConfig) {
	if err := runner.Initialize(rdsClient); err != nil {
		logx.Error(err)
		return
	}
	go runner.ScheduleTask(ctx, mgoConfig)
}

func (runner *ApiExecutor) ScheduleTask(ctx context.Context, mgoConfig config.MongoConfig) {
	var wg sync.WaitGroup
	ctx = context.WithValue(ctx, "apirunner", ApiExecutorContext{
		ExecID:  runner.ExecID,
		Result:  &runner.Result,
		LogChan: runner.LogChan,
	})
	logx.Info("执行器初始化完成")
	logx.Infof("开始执行任务 [%s]", runner.ExecID)

	murl := mongo.GetMongoUrl(mgoConfig)
	mod := task_run_log.NewTaskRunLogModel(murl, "lct", "TaskRunLog")

	wg.Add(1)
	go func(mod task_run_log.TaskRunLogModel) {
		defer func() {
			close(runner.LogChan)
			wg.Done()
		}()
		for log := range runner.LogChan {
			if log.LogType == "scene" {
				taskLog, err := mod.FindLogRecord(ctx, log.RunId, log.EventId, "", "scene")
				if err != nil {
					logx.Error(err)
				}
				if taskLog != nil && err == nil {
					updateSceneRecord(log, mod)
					continue
				}
				createSceneRecord(log, mod)
			}

			if log.LogType == "action" {
				taskLog, err := mod.FindLogRecord(ctx, log.RunId, log.SceneID, log.EventId, "action")
				if err != nil {
					logx.Error(err)
				}
				if taskLog != nil && err == nil {
					updateActionRecord(log, mod)
					continue
				}
				createActionRecord(log, mod)
			}
		}
	}(mod)

	for _, scene := range runner.Cases {
		wg.Add(1)
		logx.Infof("开始执行场景 %s--%s", scene.SceneID, scene.Description)
		go func(scene SceneConfig) {
			defer wg.Done()
			scene.Execute(ctx, runner)
		}(*scene)
	}
	wg.Wait()
	logx.Infof("执行任务 [%s] 结束", runner.ExecID)

	// 这个时候所有场景都已经执行完成
	close(runner.LogChan)
}

func updateActionRecord(log RunFlowLog, mod task_run_log.TaskRunLogModel) {
	// 1. 接受ActionFinish事件，更新ActionRecord，同时更新SceneRecord的数量
	// 2. 接受ActionUpdate事件，更新ActionRecord，同时更新SceneRecord的数量
}

func createActionRecord(log RunFlowLog, mod task_run_log.TaskRunLogModel) {
	// 接受ActionStart事件，创建ActionRecord
}

func createSceneRecord(log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	// 接受SceneStart事件，创建SceneRecord
	if err := mod.Insert(context.Background(), &task_run_log.TaskRunLog{
		ExecID:   log.RunId,
		LogType:  "scene",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		SceneDetail: task_run_log.SceneLog{
			SceneID:      log.EventId,
			Events:       []task_run_log.EventMeta{},
			FinishCount:  0,
			SuccessCount: 0,
			FailCount:    0,
			Duration:     0,
			State:        0,
			Error:        log.RootErr,
		},
	}); err != nil {
		return err
	}
	return nil
}

func updateSceneRecord(log RunFlowLog, mod task_run_log.TaskRunLogModel) {

}
