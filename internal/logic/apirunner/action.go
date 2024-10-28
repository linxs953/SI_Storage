package apirunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	ACTION_START = "Action_Start"
)

func (ac *Action) getActionPath() (path string) {
	path = fmt.Sprintf("https://%s%s", ac.Request.Domain, ac.Request.Path)
	if ac.Request.Params != nil && len(ac.Request.Params) > 0 {
		params := make([]string, 0)
		for k, v := range ac.Request.Params {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		path = fmt.Sprintf("%s?%s", path, strings.Join(params, "&"))
	}
	return
}

func (ac *Action) collectLog(logChan chan RunFlowLog, execId string, trigger string, actionEof bool, message string, err error, response map[string]interface{}) {
	var depends []map[string]interface{}
	for _, depend := range ac.Request.Dependency {
		de := make(map[string]interface{})
		de["ActionKey"] = depend.ActionKey
		de["DataKey"] = depend.DataKey
		de["Type"] = depend.Type
		refer := make(map[string]interface{})
		refer["DataType"] = depend.Refer.DataType
		refer["Target"] = depend.Refer.Target
		refer["Type"] = depend.Refer.Type
		de["Refer"] = refer
		depends = append(depends, de)
	}

	logEntry := RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		EventName:      ac.ActionName,
		RunId:          execId,
		SceneID:        ac.SceneID,
		TriggerNode:    trigger,
		Message:        message,
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestHeaders: ac.Request.Headers,
		RequestDepend:  depends,
		Response:       response,
		ActionIsEof:    actionEof,
		RootErr:        err,
	}

	// 使用非阻塞的方式发送日志
	select {
	case logChan <- logEntry:
		// 成功发送
		logx.Info("成功发送日志")
	default:
		// 通道满，记录警告日志并继续执行
		logx.Error("警告: 日志通道已满，丢弃当前日志条目")
		logx.Errorf("丢弃的日志内容: %+v", logEntry)
	}
}

func (ac *Action) TriggerAc(ctx context.Context) error {
	var err error
	initialResponse := make(map[string]interface{})
	aec := ctx.Value("apirunner").(ApiExecutorContext)

	fetchDependency := func(key string) map[string]interface{} {
		// 重试控制
		retryCount := 0
		maxRetries := 3
		retryDelay := time.Second * 2
		logx.Infof("开始获取依赖数据,key=%s", key)

		for {
			if result, ok := aec.Result.Load(key); ok {
				return result.(map[string]interface{})
			}

			retryCount++
			if retryCount > maxRetries {
				logx.Errorf("获取依赖数据失败,已达到最大重试次数,key=%s", key)
				break
			}

			logx.Infof("正在重试获取依赖数据,key=%s,第%d次重试", key, retryCount)
			time.Sleep(retryDelay)
		}
		if result, ok := aec.Result.Load(key); ok {
			return result.(map[string]interface{})
		}
		return nil
	}

	storeResultToExecutor := func(key string, data map[string]interface{}) error {
		storeKey := fmt.Sprintf("%s.%s", aec.ExecID, key)
		logx.Infof("存储数据: %s", storeKey)
		aec.Result.Store(storeKey, data)
		return nil
	}

	ac.collectLog(aec.LogChan, aec.ExecID, "Action_Start", false, fmt.Sprintf("开始执行Action: %s", ac.ActionID), nil, initialResponse)

	// 验证Action
	if err = ac.validate(); err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_Validate", true, err.Error(), err, initialResponse)
		return err
	}

	ac.collectLog(aec.LogChan, aec.ExecID, "Action_Validate_Success", false, "Action验证成功", nil, initialResponse)

	if ac.Request.Headers == nil {
		ac.Request.Headers = make(map[string]string)
	}

	// 处理Action依赖
	for _, depend := range ac.Request.Dependency {
		if err = ac.handleActionDepend(fetchDependency, aec.ExecID, depend); err != nil {
			ac.collectLog(aec.LogChan, aec.ExecID, "Action_Process_Depend", true, err.Error(), err, initialResponse)
			return err
		}
	}
	ac.collectLog(aec.LogChan, aec.ExecID, "Action_Paramination_Success", false, "Action参数化成功", nil, initialResponse)

	// 执行Action前置hook
	if err := ac.beforeAction(); err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_Before_Hook", true, err.Error(), err, initialResponse)
		return err
	}

	// 执行Action请求
	logx.Error("开始发送请求")
	resp, err := ac.sendRequest(ctx)
	if err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_SendRequest", true, err.Error(), err, initialResponse)
		return err
	}
	defer resp.Body.Close()

	ac.collectLog(aec.LogChan, aec.ExecID, "Action_SendRequest_Success", false, "Action 请求发送成功", nil, initialResponse)

	// 读取Action响应
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)

	if err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_ReadResponse", true, err.Error(), err, initialResponse)
		return err
	}

	bodyMap := make(map[string]interface{})
	if err := json.Unmarshal(buf.Bytes(), &bodyMap); err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_ReadResp", true, "resp转换成map失败", err, initialResponse)
		return err
	}

	bodyStr := buf.String()
	logx.Infof("执行完成 %v", bodyStr)

	// 转换Action响应
	result := make(map[string]interface{})
	if err := json.Unmarshal([]byte(bodyStr), &result); err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_Transform_Response", true, err.Error(), err, bodyMap)
		return err
	}

	// 执行Action后置hook
	if err := ac.afterAction(); err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_After_Hook", true, err.Error(), err, bodyMap)
		return err
	}

	ac.collectLog(aec.LogChan, aec.ExecID, "Action_After_Success", false, "Action 后置hook 执行成功", nil, bodyMap)

	// 断言Action响应
	if err := ac.expectAction(bodyMap, fetchDependency, aec.ExecID, aec.RdsClient); err != nil {
		ac.collectLog(aec.LogChan, aec.ExecID, "Action_Expect", true, err.Error(), errors.Unwrap(err), bodyMap)
		return err
	}

	ac.collectLog(aec.LogChan, aec.ExecID, "Action_Expect_Success", false, "Action 断言成功", nil, bodyMap)

	// 存储Action运行结果
	if err = storeResultToExecutor(fmt.Sprintf("%s.%s", ac.SceneID, ac.ActionID), result); err != nil {

		ac.collectLog(aec.LogChan, aec.ExecID, "Action_Ouput_Store", true, err.Error(), err, bodyMap)
		return err
	}

	ac.collectLog(aec.LogChan, aec.ExecID, "Action_Ouput_Store_Success", true, "Action 运行结果存储成功", nil, bodyMap)

	return nil
}

func (ac *Action) validate() error {
	if ac.Request.Domain == "" {
		return errors.New("domain is required")
	}
	if ac.Request.Path == "" {
		return errors.New("path is required")
	}
	if ac.Request.Method == "" {
		return errors.New("method is required")
	}
	if ac.Request.Headers == nil {
		ac.Request.Headers = make(map[string]string)
		for _, depend := range ac.Request.Dependency {
			if depend.Refer.Type != "headers" {
				continue
			}
			if depend.Type != "3" {
				continue
			}
			ac.Request.Headers[depend.Refer.Target] = depend.DataKey
		}
		return nil
	}
	return nil
}

func setPayloadField(payload map[string]interface{}, key string, value interface{}) (map[string]interface{}, error) {
	current := payload
	for pidx, part := range strings.Split(key, ".") {
		if pidx == len(strings.Split(key, "."))-1 {
			current[part] = value
		} else {
			if next, ok := current[part]; ok {
				if nextMap, ok := next.(map[string]interface{}); ok {
					current = nextMap
				} else {
					return nil, fmt.Errorf("key ['%s'] already exists but is not a map", key)
				}
			} else {
				return nil, fmt.Errorf("key ['%s'] not exist", key)
			}
		}
	}
	newPayload := current
	return newPayload, nil
}

func extractFromResp(resp map[string]interface{}, dataKey string) (interface{}, error) {
	// current := resp
	var current interface{}
	current = resp
	for _, part := range strings.Split(dataKey, ".") {
		if index, err := strconv.Atoi(part); err == nil {
			// 处理数组索引
			if arr, ok := current.([]interface{}); ok {
				if index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil, fmt.Errorf("数组索引 ['%d'] 越界", index)
				}
			} else {
				return nil, fmt.Errorf("无法处理类型 %T，期望数组", current)
			}
		} else {
			// 处理map键
			if m, ok := current.(map[string]interface{}); ok {
				if value, exists := m[part]; exists {
					current = value
				} else {
					return nil, fmt.Errorf("键 ['%s'] 不存在, value: %v", dataKey, m)
				}
			} else {
				return nil, fmt.Errorf("无法处理类型 %T，期望map", current)
			}
		}
	}
	return current, nil
}

// Action执行阶段，只有场景依赖数据是需要动态获取的，所以这个时候只需要处理depend.type=1的数据即可
func (ac *Action) handleActionDepend(fetch FetchDepend, key string, depend ActionDepend) error {
	return ac.getAndInjectDataToField(fetch, key, depend)
}

func (ac *Action) getAndInjectDataToField(fetch FetchDepend, execId string, depend ActionDepend) error {
	if !depend.IsMultiDs {
		// 单数据源
		if len(depend.DataSource) == 0 {
			return nil
		}
		if depend.DataSource[0].Type != "1" {
			return nil
		}
		val, err := ac.fetchDataFromScene(fetch, depend.DataSource[0].DataKey, fmt.Sprintf("%s.%s", execId, depend.DataSource[0].ActionKey))
		if err != nil {
			return err
		}

		// 单数据源，直接获取第一个元素，获取到数据源后，直接替换到Output.Value中
		depend.Output.Value = val

	} else {
		// 多数据源
		dataSourceMap := make(map[string]DependInject)
		for _, ds := range depend.DataSource {
			dataSourceMap[ds.DependId] = ds
		}

		for _, dsSpec := range depend.DsSpec {
			dataSource := dataSourceMap[dsSpec.DependId]
			if dataSource.Type != "1" {
				continue
			}
			logx.Errorf("获取上游的引用: %v", fmt.Sprintf("%s.%s", execId, dataSource.ActionKey))
			val, err := ac.fetchDataFromScene(fetch, dataSource.DataKey, fmt.Sprintf("%s.%s", execId, dataSource.ActionKey))
			if err != nil {
				return err
			}

			logx.Errorf("获取数据源: %v", val)

			// 读取旧值
			if depend.Output.Value == nil {
				depend.Output.Value = ""
			}
			oldVal, ok := depend.Output.Value.(string)
			if !ok {
				return fmt.Errorf("依赖 [%s] 断言string类型错误", depend.Refer.Target)
			}

			// 序列化新值
			newVal, err := ac.serializeData(val)
			if err != nil {
				return err
			}
			depend.Output.Value = strings.Replace(oldVal, fmt.Sprintf("$$%s", dsSpec.FieldName), newVal, -1)
		}
	}

	// 将数据源写入到具体字段中
	if err := ac.injectDepend(depend); err != nil {
		return err
	}
	return nil
}

func (ac *Action) serializeData(data interface{}) (string, error) {
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

func (ac *Action) fetchDataWithRedis(rdsClient *redis.Redis, resultType, key string, dataKey string) (string, error) {
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

func (ac *Action) fetchDataFromScene(fetch FetchDepend, dataKey string, key string) (interface{}, error) {
	actionResp := fetch(key)
	v, err := extractFromResp(actionResp, dataKey)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (ac *Action) injectDepend(depend ActionDepend) error {
	var err error
	var value interface{}

	// 通用的类型转换逻辑
	switch depend.Output.Type {
	case "string":
		value, err = ConvertToType[string](depend.Output.Value)
	case "int":
		value, err = ConvertToType[int](depend.Output.Value)
	case "float64":
		value, err = ConvertToType[float64](depend.Output.Value)
	case "bool":
		value, err = ConvertToType[bool](depend.Output.Value)
	case "array":
		value, err = ConvertToType[[]interface{}](depend.Output.Value)
	case "json":
		value, err = ConvertToType[map[string]interface{}](depend.Output.Value)
	default:
		// return fmt.Errorf("不支持的类型: %s", depend.Output.Type)
		logx.Errorf("不支持的类型: %s", depend.Output.Type)
		value, err = ConvertToType[string](depend.Output.Value)
	}

	if err != nil {
		return fmt.Errorf("类型转换错误 [%s]: %w", depend.Refer.Target, err)
	}

	// 根据不同的Refer.Type注入值
	switch depend.Refer.Type {
	case "headers":
		{
			if strVal, ok := value.(string); ok {
				ac.Request.Headers[depend.Refer.Target] = strVal
				logx.Errorf("headers: %v", ac.Request.Headers)
			} else {
				return fmt.Errorf("headers只支持string类型, 但得到了 %T", value)
			}
			break
		}
	case "payload":
		{
			newPayload, err := setPayloadField(ac.Request.Payload, depend.Refer.Target, value)
			if err != nil {
				return err
			}
			ac.Request.Payload = newPayload
			break
		}
	case "path":
		{
			if strVal, ok := value.(string); ok {
				ac.Request.Path = strings.ReplaceAll(ac.Request.Path, "{"+depend.Refer.Target+"}", strVal)
			} else {
				return fmt.Errorf("path只支持string类型, 但得到了 %T", value)
			}
			break
		}
	case "query":
		{
			if strVal, ok := value.(string); ok {
				ac.Request.Params[depend.Refer.Target] = strVal
			} else {
				return fmt.Errorf("query只支持string类型, 但得到了 %T", value)
			}
			break
		}
	default:
		return fmt.Errorf("不支持的Refer.Type: %s", depend.Refer.Type)
	}

	return nil
}

// 断言整个action
func (ac *Action) expectResp(respFields map[string]interface{}, fetch FetchDepend, execId string, rdsClient *redis.Redis) error {
	getFiledValue := func(resp map[string]interface{}, key string) (interface{}, error) {
		if !strings.Contains(key, ".") {
			return resp[key], nil
		}
		current := resp
		for pidx, part := range strings.Split(key, ".") {
			if pidx == len(strings.Split(key, "."))-1 {
				return current[part], nil
			} else {
				if next, ok := current[part]; ok {
					if nextMap, ok := next.(map[string]interface{}); ok {
						current = nextMap
					} else {
						return nil, fmt.Errorf("key ['%s'] already exists but is not a map", key)
					}
				} else {
					return nil, fmt.Errorf("key ['%s'] not exist", key)
				}
			}
		}
		return nil, fmt.Errorf("key ['%s'] not exist", key)
	}
	for _, ae := range ac.Expect.ApiExpect {
		if ae.Type == "api" {
			v, err := getFiledValue(respFields, ae.FieldName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("获取响应字段 [%s] 异常", ae.FieldName))
			}

			// 这里通过数据源获取desire,再传给assert断言
			var desireValue interface{}
			if ae.Desire.IsMultiDs && ae.Desire.Extra != "" {
				desireValue = ae.Desire.Extra
				dataSourceMap := make(map[string]DependInject)
				for _, ds := range ae.Desire.DataSource {
					dataSourceMap[ds.DependId] = ds
				}
				for _, dsSpec := range ae.Desire.DsSpec {
					var dsVal interface{}
					dataSource, ok := dataSourceMap[dsSpec.DependId]
					if !ok {
						return fmt.Errorf("数据源 [%s] 不存在", dsSpec.DependId)
					}

					if dataSource.Type == "1" {
						dsVal, err = ac.fetchDataFromScene(fetch, dataSource.DataKey, fmt.Sprintf("%s.%s", execId, dataSource.ActionKey))
						if err != nil {
							logx.Errorf("获取数据源 [%s] 失败: %v", fmt.Sprintf("%s.%s", execId, dataSource.ActionKey), err)
							return err
						}
					}

					if dataSource.Type == "2" {
						dsVal, err = ac.fetchDataWithRedis(rdsClient, "list", dataSource.ActionKey, dataSource.DataKey)
						if err != nil {
							logx.Errorf("获取数据源 [%s] 失败: %v", fmt.Sprintf("%s.%s", execId, dataSource.ActionKey), err)
							return err
						}
					}
					if dataSource.Type == "3" {
						dsVal = dataSource.DataKey
					}
					if dataSource.Type == "4" {
						dsVal = dataSource.DataKey
					}

					desireValue = strings.Replace(desireValue.(string), fmt.Sprintf("$$%s", dsSpec.FieldName), dsVal.(string), -1)
				}
			} else {
				if len(ae.Desire.DataSource) > 0 {
					dataSource := ae.Desire.DataSource[0]
					var err error
					if dataSource.Type == "1" {
						desireValue, err = ac.fetchDataFromScene(fetch, dataSource.DataKey, fmt.Sprintf("%s.%s", execId, dataSource.ActionKey))
						if err != nil {
							logx.Errorf("获取数据源 [%s] 失败: %v", fmt.Sprintf("%s.%s", execId, dataSource.ActionKey), err)
							return err
						}
					}
					if dataSource.Type == "2" {
						desireValue, err = ac.fetchDataWithRedis(rdsClient, "list", dataSource.ActionKey, dataSource.DataKey)
						if err != nil {
							logx.Errorf("获取数据源 [%s] 失败: %v", fmt.Sprintf("%s.%s", execId, dataSource.ActionKey), err)
							return err
						}
					}
					if dataSource.Type == "3" {
						desireValue = dataSource.DataKey
					}
					if dataSource.Type == "4" {
						desireValue = dataSource.DataKey
					}
				} else {
					desireValue = ""
				}

			}
			if ae.DataType != ae.Desire.Output.Type &&
				!(ae.DataType == "integer" && (ae.Desire.Output.Type == "number" ||
					ae.Desire.Output.Type == "int" ||
					ae.Desire.Output.Type == "int32" ||
					ae.Desire.Output.Type == "int64" ||
					ae.Desire.Output.Type == "float" ||
					ae.Desire.Output.Type == "float32" ||
					ae.Desire.Output.Type == "float64")) {
				// 数据源输出的数据类型和字段比较类型不一致
				return errors.Wrap(fmt.Errorf("预期结果和实际结果类型不一致, 预期结果类型: %s, 实际结果类型: %s", ae.Desire.Output.Type, ae.DataType), "断言失败")
			}
			assertOk, err := assert(v, desireValue, ae.DataType, ae.Operation)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("断言 [%s] %s [%s] 发生错误", ae.FieldName, ae.Operation, desireValue))
			}
			if !assertOk {
				rootErr := fmt.Errorf("断言 [%s] %s [%s] 失败", ae.FieldName, ae.Operation, desireValue)
				return errors.Wrap(rootErr, "断言失败")
			}
		}
	}
	return nil
}

func (ac *Action) beforeAction() error {
	for _, hook := range ac.Before {
		if err := hook.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (ac *Action) afterAction() error {
	for _, hook := range ac.After {
		if err := hook.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (ac *Action) sendRequest(ctx context.Context) (*http.Response, error) {
	client := &http.Client{
		// Timeout: time.Duration(ac.Conf.Timeout),
	}
	// 输入验证
	if ac.SceneID == "" || ac.ApiID == "" || ac.ActionID == "" {
		return nil, errors.New("missing required field(s)")
	}

	// ActionRequest 有效性检查（假设）
	if err := ac.validate(); err != nil {
		return nil, err
	}

	// 构建HTTP请求（伪代码）
	url := fmt.Sprintf("https://%s%s?", ac.Request.Domain, ac.Request.Path)
	if len(ac.Request.Params) == 0 {
		url = strings.TrimRight(url, "?")
	}
	for key, value := range ac.Request.Params {
		url += fmt.Sprintf("%s=%s&", key, value)
	}
	url = strings.TrimRight(url, "&")

	logx.Infof("SendRequest Url: %v", url)

	var payloadStr string
	for k, v := range ac.Request.Payload {
		payloadStr += fmt.Sprintf("%s=%v&", k, v)
	}
	payloadStr = strings.TrimRight(payloadStr, "&")
	logx.Infof("SendRequest Payload: %v", payloadStr)

	req, err := http.NewRequest(ac.Request.Method, url, strings.NewReader(payloadStr))
	if err != nil {
		return nil, err
	}

	for key, value := range ac.Request.Headers {
		req.Header.Set(key, value)
	}
	logx.Infof("SendRequest Headers: %v", req.Header)

	ac.StartTime = time.Now()

	// 发送请求并处理响应（伪代码）
	resp, err := client.Do(req)
	if err != nil {
		for ac.Request.HasRetry < ac.Conf.Retry {
			ac.Request.HasRetry++
			time.After(time.Second * 1)
			resp, err = client.Do(req)
			if err != nil {
				if ac.Request.HasRetry >= ac.Conf.Retry {
					return nil, err
				}
				continue
			}
			break
		}
	}
	ac.FinishTime = time.Now()
	ac.Duration = int(ac.FinishTime.Sub(ac.StartTime).Milliseconds())
	if resp.StatusCode != http.StatusOK {
		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logx.Error("读取响应体失败:", err)
			return nil, fmt.Errorf("读取响应体失败: %v", err)
		}

		// 记录响应状态码和响应体
		logx.Infof("响应状态码: %d", resp.StatusCode)
		logx.Infof("响应体: %s", string(body))

		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	select {
	case <-ctx.Done():
		{
			logx.Error("ctx done")
			return nil, fmt.Errorf("上下文取消, 取消执行")
		}
	case <-time.After(time.Duration(ac.Conf.Timeout) * time.Second):
		{
			logx.Error("请求超时了")
			return nil, fmt.Errorf("发送请求超时")
		}
	default:
		{
			logx.Error("触发default")
			return resp, nil
		}
	}
}

func processResponse(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()
	var respFields = make(map[string]interface{})
	respBodyMap := make(map[string]interface{})
	bytes, err := json.Marshal(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(bytes, &respBodyMap); err != nil {
		return nil, err
	}
	getMapFields(respBodyMap, "", respFields)
	return respFields, nil
}

func getMapFields(data map[string]interface{}, parentKey string, fields map[string]interface{}) map[string]interface{} {
	for key, value := range data {
		if parentKey != "" {
			key = fmt.Sprintf("%s.%s", parentKey, key)
		} else {
			key = fmt.Sprintf("%s", key)
		}
		fields[key] = value
		if _, ok := value.(map[string]interface{}); ok {
			getMapFields(value.(map[string]interface{}), key, fields)
		}
	}
	return fields
}

func (ac *Action) expectAction(respFields map[string]interface{}, fetch FetchDepend, execId string, rdsClient *redis.Redis) error {
	if err := ac.expectResp(respFields, fetch, execId, rdsClient); err != nil {
		return err
	}
	return nil
}

// ConvertToType 将 interface{} 类型的数据转换为指定的类型 T
func ConvertToType[T any](data interface{}) (T, error) {
	var result T

	// 获取目标类型
	targetType := reflect.TypeOf(result)

	// 获取输入数据的值
	value := reflect.ValueOf(data)

	// 如果输入数据是指针，获取其元素
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// 检查是否可以转换
	if !value.Type().ConvertibleTo(targetType) {
		return result, fmt.Errorf("无法将类型 %v 转换为 %v", value.Type(), targetType)
	}

	// 进行转换
	convertedValue := value.Convert(targetType)

	// 将转换后的值赋给结果
	result = convertedValue.Interface().(T)

	return result, nil
}
