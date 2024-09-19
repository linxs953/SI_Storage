package apirunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
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

func (ac *Action) TriggerAc(ctx context.Context) error {
	var err error
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

	aec.LogChan <- RunFlowLog{
		LogType:     "ACTION",
		EventId:     ac.ActionID,
		EventName:   ac.ActionName,
		RunId:       aec.ExecID,
		SceneID:     ac.SceneID,
		TriggerNode: "Action_Start",
		Message:     fmt.Sprintf("开始执行Action: %s", ac.ActionID),
		RootErr:     nil,
	}

	if err = ac.validate(); err != nil {
		logx.Error(err)
		aec.LogChan <- RunFlowLog{
			LogType:     "ACTION",
			EventId:     ac.ActionID,
			EventName:   ac.ActionName,
			RunId:       aec.ExecID,
			SceneID:     ac.SceneID,
			TriggerNode: "Action_Validate",
			Message:     err.Error(),
			RootErr:     err,
			ActionIsEof: true,
		}
		return err
	}

	aec.LogChan <- RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		EventName:      ac.ActionName,
		RunId:          aec.ExecID,
		SceneID:        ac.SceneID,
		TriggerNode:    "Action_Validate_Success",
		Message:        "Action验证成功",
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestHeaders: ac.Request.Headers,
		RootErr:        nil,
	}

	if ac.Request.Headers == nil {
		ac.Request.Headers = make(map[string]string)
	}

	for _, depend := range ac.Request.Dependency {
		if depend.Type != "1" {
			continue
		}
		if err = ac.handleActionDepend(fetchDependency, fmt.Sprintf("%s.%s", aec.ExecID, depend.ActionKey)); err != nil {
			aec.LogChan <- RunFlowLog{
				LogType:     "ACTION",
				EventId:     ac.ActionID,
				EventName:   ac.ActionName,
				RunId:       aec.ExecID,
				SceneID:     ac.SceneID,
				TriggerNode: "Action_Process_Depend",
				Message:     err.Error(),
				RootErr:     err,
				ActionIsEof: true,
			}
			return err
		}
	}

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

	aec.LogChan <- RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		EventName:      ac.ActionName,
		RunId:          aec.ExecID,
		SceneID:        ac.SceneID,
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestDepend:  depends,
		RequestHeaders: ac.Request.Headers,
		TriggerNode:    "Action_Paramination_Success",
		Message:        "Action参数化成功",
		RootErr:        nil,
	}

	if err := ac.beforeAction(); err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			TriggerNode:    "Action_Before_Hook",
			Message:        err.Error(),
			RootErr:        err,
			ActionIsEof:    true,
		}
		return err
	}

	logx.Error("开始发送请求")
	resp, err := ac.sendRequest(ctx)
	if err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_SendRequest",
			Message:        err.Error(),
			RootErr:        err,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			ActionIsEof:    true,
		}
		return err
	}
	defer resp.Body.Close()

	aec.LogChan <- RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		RunId:          aec.ExecID,
		EventName:      ac.ActionName,
		SceneID:        ac.SceneID,
		TriggerNode:    "Action_SendRequest_Success",
		Message:        "Action 请求发送成功",
		RootErr:        nil,
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestDepend:  depends,
		RequestHeaders: ac.Request.Headers,
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)

	if err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_ReadResponse",
			Message:        err.Error(),
			RootErr:        err,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			ActionIsEof:    true,
		}
		return err
	}

	bodyMap := make(map[string]interface{})
	if err := json.Unmarshal(buf.Bytes(), &bodyMap); err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_ReadResp",
			Message:        "resp转换成map失败",
			RootErr:        err,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			ActionIsEof:    true,
		}
		return err
	}

	bodyStr := buf.String()
	logx.Infof("执行完成 %v", bodyStr)

	result := make(map[string]interface{})
	if err := json.Unmarshal([]byte(bodyStr), &result); err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_Transform_Response",
			Message:        err.Error(),
			RootErr:        err,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			Response:       bodyMap,
			ActionIsEof:    true,
		}
		return err
	}

	if err := ac.afterAction(); err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_After_Hook",
			Message:        err.Error(),
			RootErr:        err,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			Response:       bodyMap,
			ActionIsEof:    true,
		}
		return err
	}

	aec.LogChan <- RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		EventName:      ac.ActionName,
		RunId:          aec.ExecID,
		SceneID:        ac.SceneID,
		TriggerNode:    "Action_After_Success",
		Message:        "Action 后置hook 执行成功",
		RootErr:        nil,
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestDepend:  depends,
		RequestHeaders: ac.Request.Headers,
		Response:       bodyMap,
	}

	if err := ac.expectAction(bodyMap); err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_Expect",
			Message:        err.Error(),
			RootErr:        errors.Unwrap(err),
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			Response:       bodyMap,
			ActionIsEof:    true,
		}
		return err
	}

	aec.LogChan <- RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		EventName:      ac.ActionName,
		RunId:          aec.ExecID,
		SceneID:        ac.SceneID,
		TriggerNode:    "Action_Expect_Success",
		Message:        "Action 断言成功",
		RootErr:        nil,
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestDepend:  depends,
		RequestHeaders: ac.Request.Headers,
		Response:       bodyMap,
	}

	if err = storeResultToExecutor(fmt.Sprintf("%s.%s", ac.SceneID, ac.ActionID), result); err != nil {
		aec.LogChan <- RunFlowLog{
			LogType:        "ACTION",
			EventId:        ac.ActionID,
			EventName:      ac.ActionName,
			RunId:          aec.ExecID,
			SceneID:        ac.SceneID,
			TriggerNode:    "Action_Ouput_Store",
			Message:        err.Error(),
			RootErr:        err,
			RequestURL:     ac.getActionPath(),
			RequestMethod:  ac.Request.Method,
			RequestPayload: ac.Request.Payload,
			RequestDepend:  depends,
			RequestHeaders: ac.Request.Headers,
			Response:       bodyMap,
			ActionIsEof:    true,
		}
		return err
	}

	aec.LogChan <- RunFlowLog{
		LogType:        "ACTION",
		EventId:        ac.ActionID,
		EventName:      ac.ActionName,
		RunId:          aec.ExecID,
		SceneID:        ac.SceneID,
		TriggerNode:    "Action_Ouput_Store_Success",
		Message:        "Action 运行结果存储成功",
		RootErr:        nil,
		RequestURL:     ac.getActionPath(),
		RequestMethod:  ac.Request.Method,
		RequestPayload: ac.Request.Payload,
		RequestDepend:  depends,
		RequestHeaders: ac.Request.Headers,
		Response:       bodyMap,
		ActionIsEof:    true,
	}

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

// Action执行阶段，只有场景依赖数据是需要动态获取的，所以这个时候只需要处理depend.type=1的数据即可
func (ac *Action) handleActionDepend(fetch FetchDepend, key string) error {
	var err error

	// 根据获取的场景依赖数据，设置payload
	setPayloadField := func(payload map[string]interface{}, key string, value interface{}) (map[string]interface{}, error) {
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

	// 获取依赖之后，根据dataKey提取具体的值
	extractFromResp := func(resp map[string]interface{}, key string) (interface{}, error) {
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

	for _, depend := range ac.Request.Dependency {
		// logx.Error(depend)
		switch depend.Refer.Type {
		case "headers":
			{
				if depend.Type == "1" {
					actionResp := fetch(key)
					v, err := extractFromResp(actionResp, depend.DataKey)
					if err != nil {
						return err
					}
					data, ok := v.(string)
					if !ok {
						return fmt.Errorf("获取依赖[%s], 断言string类型错误", fmt.Sprintf("%s", key))
					}
					if depend.Refer.Target == "Authorization" {
						ac.Request.Headers["Authorization"] = fmt.Sprintf("Bearer %s", data)
					} else {
						ac.Request.Headers[depend.Refer.Target] = data
					}
				}
				break
			}
		case "payload":
			{
				if depend.Type == "1" {
					actionResp := fetch(key)
					v, err := extractFromResp(actionResp, depend.DataKey)
					if err != nil {
						return err
					}
					data, ok := v.(string)
					if !ok {
						return fmt.Errorf("获取依赖[%s], 断言string类型错误", fmt.Sprintf("%s", key))
					}

					// 这里需要处理payload字段是多级的情况
					newPayload, err := setPayloadField(ac.Request.Payload, depend.Refer.Target, data)
					if err != nil {
						return err
					} else {
						ac.Request.Payload = newPayload
					}
				}
				break
			}
		case "path":
			{
				// path大概率是没有多级嵌套参数的情况
				if depend.Type == "1" {
					actionResp := fetch(key)
					v, err := extractFromResp(actionResp, depend.DataKey)
					if err != nil {
						return err
					}
					data, ok := v.(string)
					if !ok {
						return fmt.Errorf("获取依赖[%s], 断言string类型错误", fmt.Sprintf("%s", key))
					}
					ac.Request.Path = strings.ReplaceAll(ac.Request.Path, depend.Refer.Target, data)
				}
				break
			}
		case "query":
			{
				// url参数大概率不会有多级嵌套参数的情况
				if depend.Type == "1" {
					actionResp := fetch(key)
					v, err := extractFromResp(actionResp, depend.DataKey)
					if err != nil {
						return err
					}
					data, ok := v.(string)
					if !ok {
						return fmt.Errorf("获取依赖[%s], 断言string类型错误", fmt.Sprintf("%s", key))
					}
					ac.Request.Params[depend.Refer.Target] = data
				}
				break
			}
		}
	}
	return err
}

// 断言整个action
func (ac *Action) expectResp(respFields map[string]interface{}) error {
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

			assertOk, err := assert(v, ae.Desire, ae.DataType, ae.Operation)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("断言 [%s] %s [%s] 发生错误", ae.FieldName, ae.Operation, ae.Desire))
			}
			if !assertOk {
				rootErr := fmt.Errorf("断言 [%s] %s [%s] 失败", ae.FieldName, ae.Operation, ae.Desire)
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
		logx.Error(err)
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
		logx.Error(err)
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

func (ac *Action) expectAction(respFields map[string]interface{}) error {
	if err := ac.expectResp(respFields); err != nil {
		return err
	}
	return nil
}
