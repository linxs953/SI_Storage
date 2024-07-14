package apirunner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (ac *Action) TriggerAc(ctx context.Context) error {
	ac.validate()
	ac.processActionDepend()
	ac.sendRequest(ctx)
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
		return errors.New("headers is required")
	}
	return nil
}

func (ac *Action) processActionDepend() error {
	var err error
	matchReferFunc := func(refer string) string {
		pattern := `\$sc\.(.*?)\.(.*?)`
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(refer)
		if len(match) > 0 {
			return match[0]
		}
		return ""
	}

	getReferData := func(resp map[string]interface{}, dataKey string) interface{} {
		if strings.Contains(dataKey, ".") {
			var result interface{}
			result = resp
			parts := strings.Split(dataKey, ".")
			for _, part := range parts {
				idx, err := strconv.ParseInt(part, 10, 64)
				if err == nil {
					arr, ok := result.([]interface{})
					if !ok {
						return nil
					}

					// 引用的数组索引超出限制
					if len(arr) <= int(idx) {
						logx.Error("引用 resp.array, index 超出限制")
						return nil
					}

					if len(arr) > int(idx) {
						result = arr[idx]
						continue
					}

				} else {
					map_, ok := result.(map[string]interface{})
					if !ok {
						return nil
					}
					result = map_[part]
					continue
				}
			}
			return result
		} else {
			// 没有层级
			return resp[dataKey]
		}
	}

	for hname, hvalue := range ac.Request.Headers {
		matchKey := matchReferFunc(hvalue)
		if matchKey == "" {
			continue
		}
		// 阻塞: 读取 resp
		data := getReferData(nil, "").(string)
		ac.Request.Headers[hname] = data
	}

	for key, payload := range ac.Request.Payload {
		payloadStr, ok := payload.(string)
		if !ok {
			continue
		}
		matchKey := matchReferFunc(payloadStr)
		if matchKey == "" {
			continue
		}
		ac.Request.Payload[key] = getReferData(nil, payloadStr)
	}

	for qname, qvalue := range ac.Request.Params {
		matchKey := matchReferFunc(qvalue)
		if matchKey == "" {
			continue
		}
		data := getReferData(nil, qvalue).(string)
		ac.Request.Params[qname] = data
	}

	// 处理 api path 依赖
	matchKey := matchReferFunc(ac.Request.Path)
	if matchKey != "" {
		data := getReferData(nil, ac.Request.Path).(string)
		ac.Request.Path = data
	}

	return err
}

// func (ac *Action) ParameterizeAction(fetch FetchDepend) error {
// 	var err error
// 	for _, dp := range ac.Request.Dependency {
// 		actionResp := fetch(dp.ActionKey)
// 		value, ok := actionResp[dp.DataKey]
// 		if !ok {
// 			err = fmt.Errorf("获取 Action 结果失败, datakey=%s", dp.DataKey)
// 			return err
// 		}
// 		key := strings.Split(string(dp.Refer.Target), ".")[1]
// 		if dp.Refer.Type == "header" {
// 			valueStr := value.(string)
// 			ac.Request.Headers[key] = valueStr
// 		}
// 		if dp.Refer.Type == "path" {
// 			valueStr := value.(string)
// 			ac.Request.Path = strings.ReplaceAll(ac.Request.Path, fmt.Sprintf("$%s", key), valueStr)
// 		}
// 		if dp.Refer.Type == "payload" {
// 			switch dp.Refer.DataType {
// 			case "string":
// 				valueStr := value.(string)
// 				ac.Request.Payload[key] = valueStr
// 			case "int":
// 				valueInt := value.(int)
// 				ac.Request.Payload[key] = valueInt
// 			case "map":
// 				valueMap := value.(map[string]interface{})
// 				ac.Request.Payload[key] = valueMap
// 			case "array":
// 				valueArr := value.([]interface{})
// 				ac.Request.Payload[key] = valueArr
// 			}
// 		}
// 	}
// 	return err
// }

func (ac *Action) sendRequest(ctx context.Context) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Duration(ac.Conf.Timeout),
	}
	success := make(chan bool)
	// 输入验证
	if ac.SceneID == "" || ac.ApiID == "" || ac.ActionID == "" {
		return nil, errors.New("missing required field(s)")
	}

	// ActionRequest 有效性检查（假设）
	if err := ac.validate(); err != nil {
		return nil, err
	}

	// 构建HTTP请求（伪代码）
	url := fmt.Sprintf("%s/%s?", ac.Request.Domain, ac.Request.Path)
	for key, value := range ac.Request.Params {
		url += fmt.Sprintf("%s=%s&", key, value)
	}
	url = strings.TrimRight(url, "&")

	payloadBytes, err := json.Marshal(ac.Request.Payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(ac.Request.Method, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	for key, value := range ac.Request.Headers {
		req.Header.Set(key, value)
	}

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
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	success <- true
	select {
	case <-ctx.Done():
		{
			return nil, fmt.Errorf("上下文取消, 取消执行")
		}
	case <-time.After(time.Duration(ac.Conf.Timeout) * time.Second):
		{
			return nil, fmt.Errorf("发送请求超时")
		}
	case <-success:
		{
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
				return fmt.Errorf("field %s %s %s", ae.FieldName, ae.Operation, ae.Desire)
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
