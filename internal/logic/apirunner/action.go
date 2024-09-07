package apirunner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (ac *Action) TriggerAc(ctx context.Context) error {
	var err error
	aec := ctx.Value("apirunner").(ApiExecutorContext)
	writeLogFunc := aec.WriteLog
	storeResultToExecutor := aec.Store
	fetchDependency := aec.Fetch

	if err = ac.validate(); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_Validate", err.Error(), err)
		return err
	}

	if err = ac.handleActionDepend(fetchDependency, fmt.Sprintf("%s.%s.%s", aec.ExecID, ac.SceneID, ac.ActionID)); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_Process_Depend", err.Error(), err)
		return err
	}

	if err := ac.beforeAction(); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_Before_Hook", err.Error(), err)
		return err
	}

	logx.Error("开始发送请求")
	resp, err := ac.sendRequest(ctx)
	if err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_SendRequest", err.Error(), err)
		return err
	}
	defer resp.Body.Close()

	// 创建一个缓冲区来存储读取的数据
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_ReadResponse", err.Error(), err)
		return err
	}
	bodyStr := buf.String()
	logx.Infof("执行完成 %v", bodyStr)

	result := make(map[string]interface{})
	if err := json.Unmarshal([]byte(bodyStr), &result); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_Transform_Response", err.Error(), err)
		return err
	}

	if err := ac.afterAction(); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_After_Hook", err.Error(), err)
		return err
	}

	if err = ac.expectResp(result); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_Expect", err.Error(), err)
		return err
	}

	if err = storeResultToExecutor(fmt.Sprintf("%s.%s", ac.SceneID, ac.ActionID), result); err != nil {
		writeLogFunc("Action", ac.ActionID, "Action_Ouput_Store", err.Error(), err)
		return err
	}

	return nil
}

// func (ac *Action) store(ctx context.Context, executor *ApiExecutor, key string, data map[string]interface{}) error {
// 	executor.mu.Lock()
// 	defer executor.mu.Unlock()
// 	executor.Result[key] = data
// 	logx.Error(executor.Result)
// 	return nil
// }

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
		return errors.New("headers is required")
	}
	return nil
}

// Action执行阶段，只有场景依赖数据是需要动态获取的，所以这个时候只需要处理depend.type=1的数据即可
func (ac *Action) handleActionDepend(fetch FetchDepend, key string) error {
	var err error

	// // 获取场景依赖数据
	// fetchSceneData := func(execID, sceneId, actionId string) (interface{}, error) {
	// 	executor.mu.RLock()
	// 	defer executor.mu.RUnlock()
	// 	startTimeStamp := time.Now()
	// 	for {
	// 		if time.Since(startTimeStamp) >= time.Duration(ac.Conf.Timeout)*time.Second {
	// 			break
	// 		}
	// 		// 尝试去获取结果，直到拿到为止
	// 		if data, ok := executor.Result[fmt.Sprintf("%s.%s", sceneId, actionId)]; ok {
	// 			return data, nil
	// 		}
	// 		<-time.After(time.Second * 3)
	// 	}
	// 	return nil, errors.New("获取依赖超时")
	// }

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
		switch depend.Type {
		case "header":
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
					ac.Request.Headers[depend.Refer.Target] = data
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
					// ac.Request.Payload[depend.Refer.Target] = data
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
	for _, ae := range ac.Expect.ApiExpect {
		if ae.Type == "api_fields" {
			v, ok := respFields[ae.FieldName]
			if !ok {
				return fmt.Errorf("field %s not found", ae.FieldName)
			}

			assertOk, err := assert(v, ae.Desire, ae.DataType, ae.Operation)
			if err != nil {
				return err
			}
			if !assertOk {
				return fmt.Errorf("field %s %s %s 断言失败", ae.FieldName, ae.Operation, ae.Desire)
			}
		}
		if ae.Type == "http" {
			if ae.FieldName == "duration" {
				if ac.Duration == 0 {
					ac.Duration = int(ac.FinishTime.Sub(ac.StartTime).Milliseconds())
				}
				switch ae.Operation {
				case "gt":
					if ac.Duration > ae.Desire.(int) {
						return nil
					}
					return fmt.Errorf("duration %d is not greater than %d", ac.Duration, ae.Desire)
				case "lt":
					if ac.Duration < ae.Desire.(int) {
						return nil
					}
					return fmt.Errorf("duration %d is not less than %d", ac.Duration, ae.Desire)
				case "gte":
					if ac.Duration >= ae.Desire.(int) {
						return nil
					}
					return fmt.Errorf("duration %d is not greater than or equal to %d", ac.Duration, ae.Desire)
				case "lte":
					if ac.Duration <= ae.Desire.(int) {
						return nil
					}
					return fmt.Errorf("duration %d is not less than or equal to %d", ac.Duration, ae.Desire)
				default:
					return fmt.Errorf("unknown operation %s", ae.Operation)
				}
			}
		}
	}

	// for _, e := range ac.Expect.SqlExpect {
	// }
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

// func (ac *Action) processActionDepend(executor *ApiExecutor) error {
// 	var err error
// 	matchReferFunc := func(refer string) string {
// 		pattern := `\$sc\.(?:[a-zA-Z0-9-]{1,40}\.)*[a-zA-Z0-9-]{1,40}$`
// 		re := regexp.MustCompile(pattern)
// 		match := re.FindStringSubmatch(refer)
// 		logx.Errorf("source=%s, match=%s", refer, match)
// 		if len(match) > 0 {
// 			return match[0]
// 		}
// 		return ""
// 	}

// 	getReferData := func(resp map[string]interface{}, dataKey string) interface{} {
// 		if strings.Contains(dataKey, ".") {
// 			logx.Infof("dataKey=%s", dataKey)
// 			var result interface{}
// 			result = resp
// 			parts := strings.Split(dataKey, ".")
// 			for _, part := range parts {
// 				idx, err := strconv.ParseInt(part, 10, 64)
// 				if err == nil {
// 					arr, ok := result.([]interface{})
// 					if !ok {
// 						return nil
// 					}

// 					// 引用的数组索引超出限制
// 					if len(arr) <= int(idx) {
// 						logx.Error("引用 resp.array, index 超出限制")
// 						return nil
// 					}

// 					if len(arr) > int(idx) {
// 						result = arr[idx]
// 						continue
// 					}

// 				} else {
// 					map_, ok := result.(map[string]interface{})
// 					if !ok {
// 						return nil
// 					}
// 					result = map_[part]
// 					continue
// 				}
// 			}
// 			return result
// 		} else {
// 			// 没有层级
// 			return resp[dataKey]
// 		}
// 	}

// 	tryGetResp := func(ctx context.Context, searchKey string) (map[string]interface{}, error) {
// 		executor.mu.RLock()
// 		defer executor.mu.RUnlock()
// 		startTimeStamp := time.Now()
// 		for {
// 			if time.Since(startTimeStamp) >= time.Duration(ac.Conf.Timeout)*time.Second {
// 				break
// 			}
// 			// 尝试去获取结果，直到拿到为止
// 			resp, ok := executor.Result[searchKey]
// 			if ok {
// 				return resp, nil
// 			}
// 			<-time.After(time.Second * 3)
// 		}
// 		return nil, fmt.Errorf("获取resp超时, actionKey=[%s]", searchKey)
// 	}

// 	for hname, hvalue := range ac.Request.Headers {
// 		matchKey := matchReferFunc(hvalue)
// 		if matchKey == "" {
// 			continue
// 		}
// 		refer := strings.Replace(matchKey, "$sc.", "", -1)
// 		referParts := strings.Split(refer, ".")
// 		logx.Error(referParts)
// 		if len(referParts) < 3 {
// 			continue
// 		}
// 		var (
// 			actionKey string = fmt.Sprintf("%s.%s", referParts[0], referParts[1])
// 			dataKey   string = fmt.Sprintf("%s", strings.Join(referParts[2:], "."))
// 		)
// 		logx.Error(dataKey)
// 		resp, err := tryGetResp(context.Background(), actionKey)
// 		if err != nil {
// 			return err
// 		}
// 		data := getReferData(resp, dataKey).(string)
// 		ac.Request.Headers[hname] = strings.Replace(hvalue, matchKey, data, -1)
// 		logx.Infof("%s=[%s]", hname, ac.Request.Headers[hname])
// 	}

// 	// for key, payload := range ac.Request.Payload {
// 	// 	payloadStr, ok := payload.(string)
// 	// 	if !ok {
// 	// 		continue
// 	// 	}
// 	// 	matchKey := matchReferFunc(payloadStr)
// 	// 	if matchKey == "" {
// 	// 		continue
// 	// 	}
// 	// 	ac.Request.Payload[key] = getReferData(nil, payloadStr)
// 	// }

// 	// for qname, qvalue := range ac.Request.Params {
// 	// 	matchKey := matchReferFunc(qvalue)
// 	// 	if matchKey == "" {
// 	// 		continue
// 	// 	}
// 	// 	data := getReferData(nil, qvalue).(string)
// 	// 	ac.Request.Params[qname] = data
// 	// }

// 	// // 处理 api path 依赖
// 	// matchKey := matchReferFunc(ac.Request.Path)
// 	// if matchKey != "" {
// 	// 	data := getReferData(nil, ac.Request.Path).(string)
// 	// 	ac.Request.Path = data
// 	// }
// 	return err
// }

func (ac *Action) sendRequest(ctx context.Context) (*http.Response, error) {
	client := &http.Client{
		// Timeout: time.Duration(ac.Conf.Timeout),
	}
	// success := make(chan bool)
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

	req, err := http.NewRequest(ac.Request.Method, url, strings.NewReader(payloadStr))
	if err != nil {
		logx.Error(err)
		return nil, err
	}

	for key, value := range ac.Request.Headers {
		req.Header.Set(key, value)
	}

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
