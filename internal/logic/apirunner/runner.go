package apirunner

import (
	"context"
	"encoding/json"
	"fmt"
	"lexa-engine/internal/config"
	mongo "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/task_run_log"
	"strconv"
	"strings"
	"sync"
	"time"

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
			for dIdx, depend := range action.Request.Dependency {
				if err := runner.filledHeader(depend, dIdx, &action, rdsClient); err != nil {
					return errors.Wrap(err, "headers参数化失败")
				}
				if err := runner.filledPayload(depend, dIdx, &action, rdsClient); err != nil {
					return errors.Wrap(err, "payload参数化失败")
				}
				if err := runner.filledPath(depend, dIdx, &action, rdsClient); err != nil {
					return errors.Wrap(err, "path参数化失败")
				}
				if err := runner.filledQuery(depend, dIdx, &action, rdsClient); err != nil {
					return errors.Wrap(err, "query参数化失败")
				}
			}
		}
	}
	return err
}

func (runner *ApiExecutor) filledHeader(depend ActionDepend, dependIdx int, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "headers" {
		if action.Request.Headers == nil {
			action.Request.Headers = make(map[string]string)
		}
		if err = runner.handleActionRefer(depend, dependIdx, action, rdsClient, "headers"); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ApiExecutor) filledPayload(depend ActionDepend, dependIdx int, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "payload" {
		if err = runner.handleActionRefer(depend, dependIdx, action, rdsClient, "payload"); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ApiExecutor) filledPath(depend ActionDepend, dependIdx int, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "path" {
		if err = runner.handleActionRefer(depend, dependIdx, action, rdsClient, "path"); err != nil {
			return err
		}
	}
	return nil
}

func (runner *ApiExecutor) filledQuery(depend ActionDepend, dependIdx int, action *Action, rdsClient *redis.Redis) error {
	var err error
	if depend.Refer.Type == "params" {
		if err = runner.handleActionRefer(depend, dependIdx, action, rdsClient, "query"); err != nil {
			return err
		}
	}
	return nil
}

// 处理数据源的获取
func (runner *ApiExecutor) handleActionRefer(depend ActionDepend, dependIdx int, action *Action, rdsClient *redis.Redis, referType string) (err error) {
	// 多数据处理
	//    - 遍历DataSource,拿到里面的每个数据源定义，根据具体的定义去获取对应的数据
	//    - 拿到DsSpec，根据DsSpec的定义，填充extra模板的信息，最后写入到Output对象中
	//    - 最后拿到ActionDepend的Output对象，把Output.Value读取出来，按照Ouput.Type进行转换，赋值给字段

	actionIdx := 0
	sceneIdx := 0
	for sdx, scene := range runner.Cases {
		for idx, ac := range scene.Actions {
			if action.ActionID == ac.ActionID {
				sceneIdx = sdx
				actionIdx = idx
				break
			}
		}
	}
	if depend.IsMultiDs && depend.Extra != "" {

		// 构建DataSource映射
		dsMap := make(map[string]DependInject)
		for _, ds := range depend.DataSource {
			dsMap[ds.DependId] = ds
		}

		var multiDsVal string = depend.Extra

		// 多数据源处理
		for _, dsmap := range depend.DsSpec {
			_ = dsmap
			if dsmap.DependId == "" {
				continue
			}
			if _, ok := dsMap[dsmap.DependId]; !ok {
				continue
			}

			// dependId存在
			ds := dsMap[dsmap.DependId]
			val := runner.fetchDataSource(ds, action, rdsClient)
			if val == nil {
				continue
			}

			// 把interface{}转换成string
			valStr, err := runner.serializeData(val)
			if err != nil {
				return err
			}

			// 把获取到数据源数据填充到多数据源模板中
			multiDsVal = strings.Replace(multiDsVal, fmt.Sprintf("$$%s", dsmap.FieldName), valStr, -1)

		}

		// 模板数据填充完成后，更新到depend.Output.Value中
		runner.Cases[sceneIdx].Actions[actionIdx].Request.Dependency[dependIdx].Output.Value = multiDsVal
		runner.injectDepend(action, referType, depend.Refer.Target, sceneIdx, actionIdx, multiDsVal)
	} else {
		// 单数据源处理
		if len(depend.DataSource) == 0 {
			return nil
		}
		val := runner.fetchDataSource(depend.DataSource[0], action, rdsClient)
		if val == nil {
			return nil
		}
		valStr, err := runner.serializeData(val)
		if err != nil {
			return err
		}
		depend.Output.Value = valStr
		runner.injectDepend(action, referType, depend.Refer.Target, sceneIdx, actionIdx, valStr)
	}
	return nil
}

// 把获取的依赖注入到Request的字段中
func (runner *ApiExecutor) injectDepend(action *Action, dsType string, target string, sceneIdx, actionIdx int, val interface{}) {
	switch dsType {
	case "headers":
		{
			runner.Cases[sceneIdx].Actions[actionIdx].Request.Headers[target] = val.(string)
			break
		}
	case "payload":
		{
			newPayload, err := runner.setPayloadData(action.Request.Payload, target, val)
			if err != nil {
				return
			}
			runner.Cases[sceneIdx].Actions[actionIdx].Request.Payload = newPayload
			break
		}
	case "path":
		{
			runner.Cases[sceneIdx].Actions[actionIdx].Request.Path = strings.Replace(action.Request.Path, target, val.(string), -1)
			break
		}
	case "query":
		{
			runner.Cases[sceneIdx].Actions[actionIdx].Request.Params[target] = val.(string)
			break
		}
	}
}

// serializeData 将 interface{} 类型的数据序列化为字符串
func (runner *ApiExecutor) serializeData(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	switch v := data.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return fmt.Sprintf("%v", v), nil
	case map[string]interface{}, []interface{}:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("序列化数据失败: %w", err)
		}
		return strings.Trim(string(jsonBytes), "\""), nil
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("序列化数据失败: %w", err)
		}
		return string(jsonBytes), nil
	}
}

// 根据DependInfo的定义，获取具体的数据
func (runner *ApiExecutor) fetchDataSource(depend DependInject, action *Action, rdsClient *redis.Redis) interface{} {
	switch depend.Type {
	case "1":
		{
			// 数据源=场景，验证依赖引用关系是否合法
			err := runner.handleSceneRefer(depend, action)
			if err != nil {
				logx.Errorf("场景依赖引用关系不合法: %s", err)
				return nil
			}
			break
		}
	case "2":
		{
			// 数据源=基础数据(Redis)，获取redis数据进行填充
			val, err := runner.handleBasicRefer(rdsClient, depend)
			if err != nil {
				logx.Errorf("获取基础数据失败: %s", err)
				return nil
			}
			return val
		}
	case "3":
		{
			// 数据源=自定义, 直接覆盖原有值
			return depend.DataKey
		}
	case "4":
		{
			// 数据源=事件
			_, err := runner.handleEventRefer(ActionDepend{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 生成一个新的Request Payload
func (runner *ApiExecutor) setPayloadData(payload map[string]interface{}, key string, value interface{}) (map[string]interface{}, error) {
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

// 处理场景依赖： 用于验证引用关系是否合法
func (runner *ApiExecutor) handleSceneRefer(depend DependInject, action *Action) (err error) {
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
			if pre == relateAction {
				break
			}
			if pre != relateAction && pidx == len(preAction)-1 {
				return errors.New("依赖Action 不在当前Action的前置列表中")
			}
		}
	}

	return
}

// 处理基础数据依赖： 用于获取redis数据进行填充
func (runner *ApiExecutor) handleBasicRefer(rdsClient *redis.Redis, depend DependInject) (value string, err error) {
	// 数据源不是基础数据，不处理
	if depend.Type != "2" {
		return
	}
	value, err = runner.fetchDataWithRedis(rdsClient, "list", depend.ActionKey, depend.DataKey)
	return
}

// 获取redis数据
func (runner *ApiExecutor) fetchDataWithRedis(rdsClient *redis.Redis, resultType, key string, dataKey string) (string, error) {
	/*
		仅支持获取list和string类型的基础数据
		1. 根据dataKey获取数据，第一个字段是数字，获取list类型数据
		2. 不是数字，直接通过key获取string类型数据
		3. 遍历剩下的dataKey引用字段
		   - 把1和2获取的数据解析成map
		   - 根据dataKey引用字段，从map中获取对应的值
	*/
	_ = resultType
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

	logx.Errorf("%s,%s, %v", key, dataKey, value)

	return value.(string), nil
}

// 处理事件依赖： 用于获取事件数据进行填充
func (runner *ApiExecutor) handleEventRefer(depend ActionDepend) (value string, err error) {
	if depend.Type != "4" {
		return
	}
	return
}

func (runner *ApiExecutor) Run(ctx context.Context, taskId string, rdsClient *redis.Redis, mgoConfig config.MongoConfig) {
	if err := runner.Initialize(rdsClient); err != nil {
		logx.Error(err)
		return
	}

	// 创建task记录
	go runner.ScheduleTask(ctx, taskId, mgoConfig, rdsClient)
}

func (runner *ApiExecutor) ScheduleTask(ctx context.Context, taskId string, mgoConfig config.MongoConfig, rdsClient *redis.Redis) {
	var wg sync.WaitGroup
	ctx = context.WithValue(ctx, "apirunner", ApiExecutorContext{
		ExecID:    runner.ExecID,
		Result:    &runner.Result,
		LogChan:   runner.LogChan,
		RdsClient: rdsClient,
	})
	logx.Info("执行器初始化完成")
	logx.Infof("开始执行任务 [%s]", runner.ExecID)

	murl := mongo.GetMongoUrl(mgoConfig)
	mod := task_run_log.NewTaskRunLogModel(murl, "lct", "TaskRunLog")

	finishChan := make(chan RunFlowLog)
	go func(taskId string, mod task_run_log.TaskRunLogModel) {
		for log := range runner.LogChan {
			if log.LogType == "SCENE" {
				sceneLog, err := mod.FindLogRecord(ctx, log.RunId, log.EventId, "", "scene")
				if err != nil {
					logx.Error(err)
				}
				if sceneLog != nil && err == nil {
					updateSceneRecord(sceneLog, log, mod)
					continue
				}
				createSceneRecord(taskId, log, mod)
			}
			if log.LogType == "ACTION" {
				taskLog, err := mod.FindLogRecord(ctx, log.RunId, log.SceneID, log.EventId, "action")
				if err != nil {
					logx.Error(err)
				}
				sceneLog, err := mod.FindLogRecord(ctx, log.RunId, log.SceneID, "", "scene")
				if err != nil {
					logx.Error(err)
				}
				if taskLog != nil && err == nil {
					updateActionRecord(taskLog, log, mod)
					syncSceneRecord(sceneLog, log, mod)
					continue
				}
				createActionRecord(taskId, log, mod)
			}
		}

		// 读取 Finish 事件
		finishSign := <-finishChan
		if finishSign.LogType == "TASK" {
			taskLog, err := mod.FindLogRecord(ctx, finishSign.RunId, "", "", "task")
			if err != nil {
				logx.Error(err)
				return
			}
			// 查找所有的scene记录，看看是否有error，有的话更新状态=2，否则状态=1
			allRunScenes, err := mod.FindAllSceneRecord(ctx, finishSign.RunId, finishSign.SceneID)
			if err != nil {
				logx.Error(err)
				return
			}

			syncTaskRecord(taskLog, allRunScenes, finishSign, mod)
		}
		close(finishChan)
	}(taskId, mod)

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

	// 发送task_finish事件
	finishChan <- RunFlowLog{
		LogType:     "TASK",
		TriggerNode: "TASK_FINISH",
		RunId:       runner.ExecID,
		Message:     "任务执行完成",
	}
}

func updateActionRecord(record *task_run_log.TaskRunLog, log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	// 1. 接受ActionFinish事件，更新ActionRecord，同时更新SceneRecord的数量
	// 2. 接受ActionUpdate事件，更新ActionRecord，同时更新SceneRecord的数量
	if record == nil {
		return fmt.Errorf("record is nil")
	}
	errStr := ""
	if log.RootErr != nil {
		errStr = log.RootErr.Error()
	}

	event := task_run_log.EventMeta{
		EventName:   log.TriggerNode,
		Message:     log.Message,
		TriggerTime: time.Now(),
		Error:       errStr,
	}
	record.ActionDetail.Error = errStr
	record.ActionDetail.ActionName = log.EventName
	record.ActionDetail.Events = append(record.ActionDetail.Events, event)
	record.ActionDetail.Request = &task_run_log.RequestMeta{
		Method:     log.RequestMethod,
		URL:        log.RequestURL,
		Headers:    log.RequestHeaders,
		Payload:    log.RequestPayload,
		Dependency: log.RequestDepend,
	}
	if log.RootErr != nil {
		record.ActionDetail.State = 2
	} else {
		record.ActionDetail.State = 1
	}
	record.ActionDetail.Response = log.Response
	record.ActionDetail.Duration = int(time.Since(record.CreateAt).Milliseconds())
	record.UpdateAt = time.Now()
	if _, err := mod.Update(context.Background(), record); err != nil {
		return err
	}
	return nil
}

func syncSceneRecord(record *task_run_log.TaskRunLog, log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	if record == nil {
		return fmt.Errorf("record is nil")
	}

	if log.ActionIsEof {
		record.SceneDetail.FinishCount++
		if log.RootErr != nil {
			record.SceneDetail.FailCount++
		} else {
			record.SceneDetail.SuccessCount++
		}
	}
	if _, err := mod.Update(context.Background(), record); err != nil {
		logx.Error(err)
		return err
	}
	return nil
}

func createActionRecord(taskId string, log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	// 接受ActionStart事件，创建ActionRecord
	if err := mod.Insert(context.Background(), &task_run_log.TaskRunLog{
		ExecID:   log.RunId,
		TaskID:   taskId,
		LogType:  "action",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		ActionDetail: &task_run_log.ActionLog{
			SceneID:    log.SceneID,
			ActionID:   log.EventId,
			ActionName: log.EventName,
			Events:     []task_run_log.EventMeta{},
			Request: &task_run_log.RequestMeta{
				Method:     log.RequestMethod,
				URL:        log.RequestURL,
				Headers:    log.RequestHeaders,
				Payload:    log.RequestPayload,
				Dependency: log.RequestDepend,
			},
			Response: nil,
			State:    0,
			Duration: 0,
			Error:    "",
		},
	}); err != nil {
		return err
	}
	return nil
}

func createSceneRecord(taskId string, log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	// 接受SceneStart事件，创建SceneRecord
	errStr := ""
	if log.RootErr != nil {
		errStr = log.RootErr.Error()
	}
	if err := mod.Insert(context.Background(), &task_run_log.TaskRunLog{
		ExecID:   log.RunId,
		TaskID:   taskId,
		LogType:  "scene",
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		SceneDetail: &task_run_log.SceneLog{
			SceneID:   log.EventId,
			SceneName: log.EventName,
			Events: []task_run_log.EventMeta{
				{
					EventName:   log.TriggerNode,
					Message:     log.Message,
					TriggerTime: time.Now(),
					Error:       errStr,
				},
			},
			FinishCount:  0,
			SuccessCount: 0,
			FailCount:    0,
			Duration:     0,
			State:        0,
			Error:        errStr,
		},
	}); err != nil {
		return err
	}
	return nil
}

func updateSceneRecord(record *task_run_log.TaskRunLog, log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	if record == nil {
		return fmt.Errorf("record is nil")
	}
	errStr := ""
	if log.RootErr != nil {
		errStr = log.RootErr.Error()
	}
	event := task_run_log.EventMeta{
		EventName:   log.TriggerNode,
		Message:     log.Message,
		TriggerTime: time.Now(),
		Error:       errStr,
	}
	record.SceneDetail.Error = errStr
	if log.RootErr != nil {
		record.SceneDetail.State = 2
	} else {
		record.SceneDetail.State = 1
	}
	record.SceneDetail.SceneName = log.EventName
	record.SceneDetail.Events = append(record.SceneDetail.Events, event)
	record.UpdateAt = time.Now()
	// 计算时间戳
	record.SceneDetail.Duration = int(time.Since(record.CreateAt).Milliseconds())
	if _, err := mod.Update(context.Background(), record); err != nil {
		return err
	}
	return nil
}

func syncTaskRecord(record *task_run_log.TaskRunLog, allRunScene []*task_run_log.TaskRunLog, log RunFlowLog, mod task_run_log.TaskRunLogModel) error {
	if record == nil {
		return fmt.Errorf("record is nil")
	}
	if len(allRunScene) == 0 {
		return fmt.Errorf("找不到运行中的场景")
	}

	isSuccess := true
	for _, scene := range allRunScene {
		if scene == nil {
			continue
		}
		if scene.SceneDetail.State == 2 {
			isSuccess = false
		}
	}

	if isSuccess {
		record.TaskDetail.TaskState = 1
	} else {
		record.TaskDetail.TaskState = 2
	}

	if _, err := mod.Update(context.Background(), record); err != nil {
		return err
	}
	return nil
}
