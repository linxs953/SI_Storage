package apirunner

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

)

func (runner *ApiExecutor) Initialize(rdsClient *redis.Redis) {
	runner.parametrization(rdsClient)
	// requestID := uuid.New().String()
}

func (runner *ApiExecutor) parametrization(rdsClient *redis.Redis) error {
	var err error
	for sdx, scene := range runner.Cases {
		for adx, action := range scene.Actions {
			currentRefer := action.CurrentRefer
			// 处理payload
			for field, value := range action.Request.Payload {
				valueStr := value.(string)
				referExpressList, data := runner.proccessRefer(rdsClient, `\$[^ ]*`, currentRefer, valueStr)
				if len(referExpressList) > 1 {
					for _, referExpress := range referExpressList {
						runner.Cases[sdx].Actions[adx].Request.Payload[field] = data[referExpress]
					}
				}
			}

			//  处理path
			referExpressList, data := runner.proccessRefer(rdsClient, `\$(\w+(\.\w+|\:\w+)+)`, action.Request.Path, currentRefer)
			if len(referExpressList) > 1 {
				for _, referExpress := range referExpressList {
					apiPath := runner.Cases[sdx].Actions[adx].Request.Path
					runner.Cases[sdx].Actions[adx].Request.Path = strings.Replace(apiPath, referExpress, data[referExpress], -1)
				}
			}

			// 处理query
			for field, value := range action.Request.Params {
				referExpressList, data := runner.proccessRefer(rdsClient, `\$(\w+(\.\w+|\:\w+)+)`, value, currentRefer)
				if len(referExpressList) > 1 {
					for _, referExpress := range referExpressList {
						fieldValue := runner.Cases[sdx].Actions[adx].Request.Params[field]
						runner.Cases[sdx].Actions[adx].Request.Params[field] = strings.Replace(fieldValue, referExpress, data[referExpress], -1)
					}
				}
			}

			// 处理headers
			for key, value := range action.Request.Headers {
				referExpressList, data := runner.proccessRefer(rdsClient, `\$(\w+(\.\w+|\:\w+)+)`, currentRefer, value)
				if len(referExpressList) > 1 {
					for _, referExpress := range referExpressList {
						headervalue := runner.Cases[sdx].Actions[adx].Request.Headers[key]
						runner.Cases[sdx].Actions[adx].Request.Headers[key] = strings.Replace(headervalue, referExpress, data[referExpress], -1)
					}
				}
			}
		}
	}
	return err
}

// 处理数据源依赖
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
				result = fmt.Sprintf("$sc.%s.%s", runner.SceneMap[currentActParts[0]], pa)
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
	result = fmt.Sprintf("$sc.%s.%s", targetSceneId, targetActionId)
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

func (runner *ApiExecutor) Run(ctx context.Context, rdsClient *redis.Redis) {
	runner.Initialize(rdsClient)
	ctx = context.WithValue(ctx, "apirunner", ApiExecutorContext{
		ExecID: uuid.New().String(),
		Store:  runner.StoreActionResult,
		Fetch:  runner.FetchDependency,
	})
	for _, scene := range runner.Cases {
		go scene.Execute(ctx)
	}
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
	defer runner.mu.Unlock()
	runner.Result[key] = respFields
	return nil
}
