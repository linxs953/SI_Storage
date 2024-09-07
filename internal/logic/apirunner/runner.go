package apirunner

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
	// for sdx, scene := range runner.Cases {
	// 	for adx, action := range scene.Actions {
	// 		currentRefer := action.CurrentRefer
	// 		// 处理payload
	// 		for field, value := range action.Request.Payload {
	// 			valueStr := value.(string)
	// 			referExpressList, data := runner.proccessRefer(rdsClient, `\$[^ ]*`, currentRefer, valueStr)
	// 			if len(referExpressList) > 0 {
	// 				for _, referExpress := range referExpressList {
	// 					runner.Cases[sdx].Actions[adx].Request.Payload[field] = data[referExpress]
	// 				}
	// 			}
	// 		}

	// 		//  处理path
	// 		referExpressList, data := runner.proccessRefer(rdsClient, `\$(\w+(\.\w+|\:\w+)+)`, action.Request.Path, currentRefer)
	// 		if len(referExpressList) > 0 {
	// 			for _, referExpress := range referExpressList {
	// 				apiPath := runner.Cases[sdx].Actions[adx].Request.Path
	// 				runner.Cases[sdx].Actions[adx].Request.Path = strings.Replace(apiPath, referExpress, data[referExpress], -1)
	// 			}
	// 		}

	// 		// 处理query
	// 		for field, value := range action.Request.Params {
	// 			referExpressList, data := runner.proccessRefer(rdsClient, `\$(\w+(\.\w+|\:\w+)+)`, value, currentRefer)
	// 			if len(referExpressList) > 0 {
	// 				for _, referExpress := range referExpressList {
	// 					fieldValue := runner.Cases[sdx].Actions[adx].Request.Params[field]
	// 					runner.Cases[sdx].Actions[adx].Request.Params[field] = strings.Replace(fieldValue, referExpress, data[referExpress], -1)
	// 				}
	// 			}
	// 		}

	// 		// 处理headers
	// 		for key, value := range action.Request.Headers {
	// 			referExpressList, data := runner.proccessRefer(rdsClient, `\$(\w+(\.\w+|\:\w+)+)`, currentRefer, value)
	// 			if len(referExpressList) > 0 {
	// 				for _, referExpress := range referExpressList {
	// 					headervalue := runner.Cases[sdx].Actions[adx].Request.Headers[key]
	// 					logx.Errorf("header value: %s", headervalue)
	// 					runner.Cases[sdx].Actions[adx].Request.Headers[key] = strings.Replace(headervalue, referExpress, data[referExpress], -1)
	// 					logx.Errorf("header value: %s", runner.Cases[sdx].Actions[adx].Request.Headers[key])
	// 				}
	// 			}
	// 		}
	// 	}
	// }

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
		if err = runner.handleActionRefer(depend, action, rdsClient, "headers"); err != nil {
			return err
		}
	}
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

func (runner *ApiExecutor) Run(ctx context.Context, rdsClient *redis.Redis) {
	if err := runner.Initialize(rdsClient); err != nil {
		logx.Error(err)
		causeErr := errors.Cause(err)
		runner.WriteLog("Executor_Initialize", runner.ExecID, "Executor_Initialize", causeErr.Error(), err)
		return
	}
	ctx = context.WithValue(ctx, "apirunner", ApiExecutorContext{
		ExecID:   uuid.New().String(),
		Store:    runner.StoreActionResult,
		Fetch:    runner.FetchDependency,
		WriteLog: runner.WriteLog,
	})
	logx.Info("执行器初始化完成")
	// for _, scene := range runner.Cases {
	// 	logx.Infof("开始执行场景 --%s", scene.Description)
	// 	logx.Error(scene)
	// 	go scene.Execute(ctx, runner)
	// }
}

func (runner *ApiExecutor) FetchDependency(key string) map[string]interface{} {
	// 确保在函数开始就锁定，并在函数结束时释放，即使发生panic
	defer runner.mu.RUnlock()
	runner.mu.RLock()
	if result, ok := runner.Result[key]; ok {
		return result
	}

	// 使用context来控制超时，避免死锁
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // 假设最大超时时间为10秒
	defer cancel()

	ticker := time.NewTicker(5 * time.Second) // 初始化ticker，每次间隔5秒
	defer ticker.Stop()

	retryCount := 0
	const initialDelay = 1 * time.Second // 初始重试延迟
	const maxDelay = 5 * time.Second     // 重试延迟的最大值

	for {
		select {
		default:
			// 为了减少锁的竞争，仅在需要时加锁
			runner.mu.RLock()
			defer runner.mu.RUnlock()

			if result, ok := runner.Result[key]; ok {
				// 如果找到依赖项，返回结果
				return result
			}

			// 重试逻辑
			retryCount++
			if retryCount > runner.Conf.Retry {
				// 达到最大重试次数，记录日志
				logx.Errorf("FetchDependency: Max retries reached, dependency with key %s not fetched.", key)
				return make(map[string]interface{})
			}

			delay := initialDelay << uint(retryCount-1)
			if delay > maxDelay {
				delay = maxDelay
			}
			time.Sleep(delay)

		case <-ctx.Done():
			// 优化错误日志，包含更多上下文信息
			logx.Errorf("FetchDependency: Timeout reached, dependency with key %s not fetched.", key)
			return make(map[string]interface{}) // 返回空map而不是nil
		}
	}
}

func (runner *ApiExecutor) StoreActionResult(actionKey string, respFields map[string]interface{}) error {
	if actionKey == "" || respFields == nil {
		return fmt.Errorf("actionKey cannot be empty and respFields must not be nil")
	}
	key := fmt.Sprintf("%s.%s", runner.ExecID, actionKey)
	runner.mu.Lock()
	defer func() {
		// 确保在发生 panic 时解锁
		defer runner.mu.Unlock()

		if r := recover(); r != nil {
			fmt.Printf("Recovered in StoreActionResult: %v\n", r)
			return
		}
	}()
	runner.Result[key] = respFields
	return nil
}

func (runner *ApiExecutor) WriteLog(logType, eventId, trigger_node, message string, err error) error {
	logObject := RunFlowLog{
		RunId:       runner.ExecID,
		EventId:     eventId,
		LogType:     logType,
		TriggerNode: trigger_node,
		Message:     message,
		RootErr:     errors.Unwrap(err),
	}
	runner.LogSet = append(runner.LogSet, logObject)
	return nil
}

/*
暂时废弃

func (runner *ApiExecutor) proccessRefer(rdsClient *redis.Redis, expression string, currentRefer string, value string) ([]string, map[string]string) {
	re := regexp.MustCompile(expression)
	matches := re.FindAllStringSubmatch(value, -1)
	var results map[string]string = make(map[string]string)
	var keys []string
	if len(matches) > 0 {
		for _, match := range matches {
			if len(match) < 1 {
				continue
			}
			data, err := runner.fetchRefer(rdsClient, match[0], currentRefer)
			if err != nil {
				logx.Error(err)
				continue
			}
			logx.Error(data)
			results[match[0]] = data
		}
		for key, _ := range results {
			keys = append(keys, key)
		}
		return keys, results
	}
	return nil, nil
}

func (runner *ApiExecutor) fetchRefer(rdsClient *redis.Redis, fetchRefer string, currentRefer string) (string, error) {
	referExpress := strings.TrimPrefix(fetchRefer, "$")
	actionKeyParts := strings.Split(referExpress, ".")
	dsType := actionKeyParts[0]
	switch dsType {
	case "rds":
		return runner.getReferRedis(rdsClient, fetchRefer)
	case "sc":
		return runner.getReferAction(currentRefer, fetchRefer)
	}
	return "", fmt.Errorf("不支持的数据源类型, %s", dsType)
}

func (runner *ApiExecutor) getReferRedis(rdsClient *redis.Redis, fetchRefer string) (string, error) {
	if fetchRefer == "" {
		return "", fmt.Errorf("fetch refer is empty")
	}
	if !strings.Contains(fetchRefer, ".") {
		return "", fmt.Errorf("fetch refer is invalid")
	}
	if rdsClient != nil {
		// 获取redis数据源
		referExpress := strings.TrimPrefix(fetchRefer, "$")
		actionKeyParts := strings.Split(referExpress, ".")
		dsType, acKey := actionKeyParts[0], actionKeyParts[1]

		if dsType == "rds" {
			// redis的数据依赖
			// dk := fmt.Sprintf("%s", strings.Join(actionKeyParts[2:], "."))
			dk := strings.Join(actionKeyParts[2:], ".")
			logx.Infof("获取REDIS数据源, %s", referExpress)
			data, err := runner.fetchWithRedis(rdsClient, acKey, dk)
			if err != nil {
				logx.Error(err)
				return "", err
			}
			return data, nil
		}
	}
	return "", fmt.Errorf("redis client is nil")
}

// 传入一个action key 的表达式，转换为actionid的表达式
func (runner *ApiExecutor) getReferAction(currentActRefer string, fetchActionRefer string) (string, error) {
	logx.Infof("%s, %s", currentActRefer, fetchActionRefer)
	if currentActRefer == "" {
		return "", fmt.Errorf("current action refer is empty")
	}
	if fetchActionRefer == "" {
		return "", fmt.Errorf("fetch action refer is empty")
	}
	if currentActRefer == fetchActionRefer {
		return "", fmt.Errorf("current action refer is equal to fetch action refer")
	}
	if !strings.Contains(fetchActionRefer, ".") {
		return "", fmt.Errorf("fetch action refer is invalid")
	}
	if !strings.Contains(currentActRefer, ".") {
		return "", fmt.Errorf("current action refer is invalid")
	}
	result := ""
	referExpress := strings.TrimPrefix(fetchActionRefer, "$sc.")

	currentActParts := strings.Split(currentActRefer, ".")
	fetchActParts := strings.Split(referExpress, ".")
	if currentActParts[0] == fetchActParts[0] {
		// 引用的是同个场景
		logx.Error("引用同个场景")
		preActions := runner.PreActionsMap[currentActParts[1]]
		for _, pa := range strings.Split(preActions, ",") {
			if pa == runner.ActionMap[fetchActParts[1]] {
				result = fmt.Sprintf("$sc.%s.%s.%s", runner.SceneMap[currentActParts[0]], pa, strings.Join(fetchActParts[2:], "."))
				return result, nil
			}
		}
		// 引用的场景不在当前场景的前置场景中
		return "", fmt.Errorf("引用的action [%s] 不在前置actions中", fetchActionRefer)
	}

	// 引用的是其他场景，获取对应场景sceneid和actionid 返回
	targetSceneId := runner.SceneMap[currentActParts[0]]
	targetActionId := runner.ActionMap[fetchActParts[1]]
	// 加上action 后面引用的东西
	result = fmt.Sprintf("$sc.%s.%s.%s", targetSceneId, targetActionId, strings.Join(fetchActParts[2:], "."))
	return result, nil
}

func (runner *ApiExecutor) fetchWithRedis(rdsClient *redis.Redis, key string, dataKey string) (string, error) {
	// todo： 目前只获取list类型的key
	var result interface{}
	dkParts := strings.Split(dataKey, ".")
	if len(dkParts) >= 3 {
		return "", fmt.Errorf("dataKey格式错误, 超过3级,  格式: field1.field2 / index/field1, datakey: %s", dataKey)
	}
	if len(dkParts) == 0 {
		return "", fmt.Errorf("dataKey格式错误,没有包含., 格式: field1.field2 / index/field1, datakey: %s", dataKey)
	}
	if len(dkParts) == 1 {
		return "", fmt.Errorf("dataKey格式错误, 只有1级, 格式: field1.field2 / index/field1, datakey: %s", dataKey)
	}
	arrIdx, err := strconv.ParseInt(dkParts[0], 10, 64)
	if err != nil {
		return "", err
	}
	result, err = rdsClient.Lrange(key, int(arrIdx), int(arrIdx))
	if err != nil {
		return "", err
	}
	if len(result.([]string)) == 0 {
		return "", fmt.Errorf("根据key=%v 获取redis list为空", dataKey)
	}
	ele := result.([]string)[0]
	eleMap := make(map[string]string)
	if err = json.Unmarshal([]byte(ele), &eleMap); err != nil {
		return "", err
	}
	return eleMap[dkParts[1]], err
}

*/
