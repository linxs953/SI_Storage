package runner

import (
	"Storage/internal/logic/tools"
	"Storage/internal/logic/workflows/api"
	"Storage/internal/logic/workflows/api/apirunner/dependency"
	expect "Storage/internal/logic/workflows/api/apirunner/expect"
	"Storage/internal/logic/workflows/api/apirunner/store"
	"Storage/internal/logic/workflows/core"
	"Storage/internal/logic/workflows/core/metrics/reporter"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urls "net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
)

// HttpRunner HTTP API执行器
type HttpRunner struct {
	// HTTP客户端
	client *http.Client

	// 上下文数据
	contextData map[string]interface{}

	// 执行状态
	status core.TaskStatus

	// API指标
	metrics *api.ApiMetrics

	// 指标上报器
	metricsReporter reporter.MetricsReporter
}

// HttpMetricsReporter 扩展核心指标上报接口，专用于HTTP API指标
type HttpMetricsReporter interface {
	// 嵌入核心接口
	reporter.MetricsReporter

	// HTTP特有的指标上报方法
	ReportHttpMetrics(ctx context.Context, metrics *api.ApiMetrics) error
}

// NewHttpRunner 创建新的HTTP执行器
func NewHttpRunner(contextData map[string]interface{}) *HttpRunner {
	if contextData == nil {
		contextData = make(map[string]interface{})
	}

	return &HttpRunner{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		contextData: contextData,
		status:      core.TaskStatusPending,
		metrics:     &api.ApiMetrics{},
	}
}

// SetMetricsReporter 设置指标上报器
func (r *HttpRunner) SetMetricsReporter(reporter reporter.MetricsReporter) {
	r.metricsReporter = reporter
}

// Initialize 初始化执行器
func (r *HttpRunner) Initialize(ctx context.Context) error {
	r.status = core.TaskStatusPending
	r.metrics = &api.ApiMetrics{}
	return nil
}

// Validate 验证配置
func (r *HttpRunner) Validate(ctx context.Context) error {
	// HTTP执行器不需要特殊验证
	return nil
}

// Execute 执行API请求
func (r *HttpRunner) Execute(ctx context.Context, spec map[string]interface{}) (map[string]interface{}, error) {
	// 将规格转换为API定义
	apiDef, err := api.ConvertToApiDefinition(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to convert spec to API definition: %w", err)
	}

	// 准备依赖数据
	dependencies := make([]dependency.Dependency, 0)
	if deps, ok := spec["dependencies"].([]dependency.Dependency); ok {
		dependencies = deps
	} else if depsArray, ok := spec["dependencies"].([]interface{}); ok {
		// 转换通用接口数组为依赖数组
		for _, dep := range depsArray {
			if depMap, ok := dep.(map[string]interface{}); ok {
				dependencies = append(dependencies, dependency.Dependency{
					DependID: fmt.Sprintf("%v", depMap["depend_id"]),
					Name:     fmt.Sprintf("%v", depMap["name"]),
					Type:     fmt.Sprintf("%v", depMap["type"]),
					Value:    depMap["value"],
				})
			}
		}
	}

	// 准备依赖数据
	dependencyValues, err := r.PrepareDependencies(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	// 构建请求
	request, err := r.BuildRequest(ctx, apiDef, dependencyValues)
	if err != nil {
		return nil, err
	}

	// 执行请求
	response, err := r.ExecuteRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// 处理响应验证
	assertions := make([]expect.Assertion, 0)
	if assertList, ok := spec["assertions"].([]expect.Assertion); ok {
		assertions = assertList
	}

	assertGroups := expect.AssertionGroup{
		Assertions: assertions,
		Name:       "Default Assertions",
		Options: expect.GroupOptions{
			StopOnFirstFailure: false,
			Timeout:            0,
			Parallel:           false,
			Retry: &expect.RetryConfig{
				MaxRetries: 1,
				Strategy:   expect.RetryLinearBackoff,
			},
		},
		Description: "Default Assertions",
	}
	if groupList, ok := spec["assert_groups"].([]expect.AssertionGroup); ok {
		assertGroups = groupList[0]
	}

	validationResult, err := r.ValidateResponse(ctx, response, &assertGroups)
	if err != nil {
		// 将验证错误包含到响应中，但不中断执行
		response["validation_error"] = err.Error()
	}

	if validationResult != nil {
		response["validation_result"] = validationResult
	}

	// 提取数据
	extractors := make(map[string]string)
	if ext, ok := spec["extractors"].(map[string]string); ok {
		extractors = ext
	}

	extractedData, err := r.ExtractData(ctx, response, extractors)
	if err != nil {
		return nil, err
	}

	if len(extractedData) > 0 {
		response["extracted_data"] = extractedData
	}

	// 上报指标（如果配置了）
	reportConfig := &api.ReportConfig{}
	if cfg, ok := spec["report_config"].(*api.ReportConfig); ok {
		reportConfig = cfg
	}

	if reportConfig.Enabled && r.metricsReporter != nil {
		err = r.ReportMetrics(ctx, r.metrics, reportConfig)
		if err != nil {
			// 记录错误但不中断执行
			response["metrics_report_error"] = err.Error()
		}
	}

	return response, nil
}

// PrepareDependencies 准备依赖数据
func (r *HttpRunner) PrepareDependencies(ctx context.Context, dependencies []dependency.Dependency) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 复制上下文数据
	for k, v := range r.contextData {
		result[k] = v
	}

	// 处理依赖项
	for _, dep := range dependencies {
		switch dep.Type {
		case api.DependTypeVariable:
			// 处理变量依赖
			if value, ok := dep.Value.(string); ok {
				result[dep.Name] = value
			} else {
				result[dep.Name] = dep.Value
			}
		case api.DependTypeHeader, api.DependTypeQuery, api.DependTypeBody:
			// 这些依赖类型将在构建请求时处理
			continue
		case api.DependTypeResponse:
			// 响应依赖项应由上一个请求处理，这里仅确保
		}
	}

	return result, nil
}

// BuildRequest 构建HTTP请求
func (r *HttpRunner) BuildRequest(ctx context.Context, api *api.ApiDefinition, dependencies map[string]interface{}) (map[string]interface{}, error) {
	request := make(map[string]interface{})

	// 设置请求方法
	request["method"] = api.Method

	// 处理URL和路径参数
	// TODO: 实现路径参数替换
	request["url"] = api.Path

	// 处理请求头
	headers := make(map[string]string)
	if api.Headers != nil {
		for k, v := range api.Headers {
			// TODO: 处理从依赖中替换的值
			headers[k] = v
		}
	}
	request["headers"] = headers

	// 处理查询参数
	queryParams := make(map[string]string)
	if api.QueryParams != nil {
		for k, v := range api.QueryParams {
			// TODO: 处理从依赖中替换的值
			queryParams[k] = v
		}
	}
	request["query_params"] = queryParams

	// 处理请求体
	switch api.BodyType {
	case "json":
		request["body"] = api.Body
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "application/json"
		}
	case "form":
		formData := make(map[string]interface{})
		if body, ok := api.Body.(map[string]interface{}); ok {
			formData = body
		}
		request["body"] = formData
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	case "multipart":
		// TODO: 处理多部分表单数据
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "multipart/form-data"
		}
	case "raw":
		if raw, ok := api.Body.(string); ok {
			request["body"] = raw
		}
	case "binary":
		// TODO: 处理二进制数据
	default:
		// 默认为JSON
		request["body"] = api.Body
		if _, ok := headers["Content-Type"]; !ok && api.Body != nil {
			headers["Content-Type"] = "application/json"
		}
	}

	return request, nil
}

// ExecuteRequest 执行HTTP请求
func (r *HttpRunner) ExecuteRequest(ctx context.Context, request map[string]interface{}) (map[string]interface{}, error) {
	response := make(map[string]interface{})

	// 获取请求参数
	method, _ := request["method"].(string)
	url, _ := request["url"].(string)
	headers, _ := request["headers"].(map[string]string)
	body, _ := request["body"].(map[string]interface{})
	queryParams, _ := request["query_params"].(map[string]string)

	// 检查请求方法和URL
	if method == "" {
		return nil, fmt.Errorf("请求方法不能为空")
	}
	if url == "" {
		return nil, fmt.Errorf("请求URL不能为空")
	}

	// 处理查询参数
	if len(queryParams) > 0 {
		urlObj, err := urls.Parse(url)
		if err != nil {
			return nil, fmt.Errorf("解析URL失败: %w", err)
		}

		// 获取现有查询参数
		query := urlObj.Query()
		for k, v := range queryParams {
			query.Set(k, v)
		}

		// 更新URL
		urlObj.RawQuery = query.Encode()
		url = urlObj.String()
	}

	// 准备请求体
	var reqBody []byte
	var err error
	var reqBodyReader *bytes.Reader
	contentType := ""

	if headers != nil {
		for k, v := range headers {
			if strings.ToLower(k) == "content-type" {
				contentType = v
				break
			}
		}
	}

	if body != nil {
		switch {
		case strings.Contains(contentType, "application/json"):
			reqBody, err = json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("序列化JSON请求体失败: %w", err)
			}
		case strings.Contains(contentType, "application/x-www-form-urlencoded"):
			formValues := urls.Values{}
			for k, v := range body {
				formValues.Set(k, fmt.Sprintf("%v", v))
			}
			reqBody = []byte(formValues.Encode())
		default:
			// 默认尝试JSON序列化
			reqBody, err = json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("序列化请求体失败: %w", err)
			}
		}
	}

	if reqBody != nil {
		reqBodyReader = bytes.NewReader(reqBody)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, method, url, reqBodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 记录原始请求
	rawRequest := fmt.Sprintf("%s %s\n", req.Method, req.URL.String())
	for k, v := range req.Header {
		rawRequest += fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", "))
	}
	if reqBody != nil {
		rawRequest += "\n" + string(reqBody)
	}
	response["raw_request"] = rawRequest

	// 更新执行状态
	r.status = core.TaskStatusRunning

	// 发送请求
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 记录响应结束时间
	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 记录原始响应
	rawResponse := fmt.Sprintf("HTTP/%d.%d %d %s\n", resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status)
	for k, v := range resp.Header {
		rawResponse += fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", "))
	}
	rawResponse += "\n" + string(respBody)
	response["raw_response"] = rawResponse

	// 设置响应数据
	response["status_code"] = resp.StatusCode
	response["status"] = resp.Status
	response["headers"] = make(map[string]string)
	for k, v := range resp.Header {
		response["headers"].(map[string]string)[k] = strings.Join(v, ", ")
	}
	response["body"] = string(respBody)
	response["duration"] = duration

	// 尝试解析JSON响应
	contentType = resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal(respBody, &jsonData); err == nil {
			response["json"] = jsonData
		}
	}

	// 更新指标
	// 从request中提取API信息
	if apiID, ok := request["api_id"].(string); ok {
		r.metrics.ApiID = apiID
	}
	if apiName, ok := request["api_name"].(string); ok {
		r.metrics.ApiName = apiName
	}
	r.metrics.Method = method
	r.metrics.Path = url
	r.metrics.StartTime = startTime.Format(time.RFC3339)
	r.metrics.EndTime = endTime.Format(time.RFC3339)
	r.metrics.Duration = duration
	r.metrics.StatusCode = resp.StatusCode
	r.metrics.RequestSize = int64(len(reqBody))
	r.metrics.ResponseSize = int64(len(respBody))

	// 设置状态
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		r.metrics.Status = "succeeded"
		r.status = core.TaskStatusCompleted
	} else {
		r.metrics.Status = "failed"
		r.metrics.Error = &core.PipelineError{
			Message: fmt.Sprintf("HTTP request failed with status code: %d", resp.StatusCode),
			Code:    "HTTP_ERROR",
		}
		r.status = core.TaskStatusFailed
	}

	return response, nil
}

// ValidateResponse 验证响应
func (r *HttpRunner) ValidateResponse(ctx context.Context, response map[string]interface{}, assertions *expect.AssertionGroup) (*expect.AssertionGroupResult, error) {
	result := assertions.AssertAll()

	var errorCount int
	for _, assertion := range result.Results {
		if !assertion.Passed {
			errorCount++
		}
	}

	// 更新指标
	r.metrics.AssertionsPassed = len(result.Results) - errorCount
	r.metrics.AssertionsFailed = errorCount

	return result, nil
}

// getValueByPath 通过路径获取值
func getValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	// TODO: 实现更复杂的JSON路径解析
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part], nil
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("路径 %s 不存在", path)
		}
	}

	return current, nil
}

// ExtractData 从响应中提取数据
func (r *HttpRunner) ExtractData(ctx context.Context, response map[string]interface{}, extractors map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for name, path := range extractors {
		value, err := getValueByPath(response, path)
		if err != nil {
			return nil, fmt.Errorf("提取数据失败: %w", err)
		}

		result[name] = value
	}

	return result, nil
}

// ReportMetrics 上报指标
func (r *HttpRunner) ReportMetrics(ctx context.Context, metrics *api.ApiMetrics, config *api.ReportConfig) error {
	if r.metricsReporter == nil {
		return nil
	}

	// 检查是否支持HTTP特有的上报接口
	if httpReporter, ok := r.metricsReporter.(HttpMetricsReporter); ok {
		return httpReporter.ReportHttpMetrics(ctx, metrics)
	}

	// 将API特定指标转换为通用指标格式
	genericMetrics := r.convertApiMetricsToGeneric(metrics, config)

	// 使用核心上报器上报
	return r.metricsReporter.Report(ctx, genericMetrics)
}

// convertApiMetricsToGeneric 将API指标转换为通用指标格式
func (r *HttpRunner) convertApiMetricsToGeneric(metrics *api.ApiMetrics, config *api.ReportConfig) map[string]interface{} {
	result := make(map[string]interface{})

	// 基本信息
	result["type"] = "http_api"
	result["timestamp"] = time.Now().UnixNano() / int64(time.Millisecond)

	// API指标
	result["duration_ms"] = metrics.Duration * 1000 // 转换秒为毫秒
	result["status_code"] = metrics.StatusCode
	result["status"] = metrics.Status
	result["success"] = metrics.Error == nil // 根据是否有错误判断成功与否
	result["request_size"] = metrics.RequestSize
	result["response_size"] = metrics.ResponseSize
	result["assertions_passed"] = metrics.AssertionsPassed
	result["assertions_failed"] = metrics.AssertionsFailed

	// 添加API统计信息
	result["api_id"] = metrics.ApiID
	result["api_name"] = metrics.ApiName
	result["method"] = metrics.Method
	result["path"] = metrics.Path
	result["start_time"] = metrics.StartTime
	result["end_time"] = metrics.EndTime

	// 添加自定义指标
	if metrics.CustomMetrics != nil {
		for k, v := range metrics.CustomMetrics {
			result[k] = v
		}
	}

	// 添加配置中的标签
	if config != nil && config.Config != nil {
		// 如果配置中有labels键
		if labels, ok := config.Config["labels"].(map[string]interface{}); ok {
			result["labels"] = labels
		} else {
			// 将整个Config作为标签
			result["config"] = config.Config
		}
	}

	return result
}

// Cancel 取消执行
func (r *HttpRunner) Cancel(ctx context.Context) error {
	// TODO: 实现取消HTTP请求
	r.status = core.TaskStatusCanceled
	return nil
}

// GetStatus 获取状态
func (r *HttpRunner) GetStatus(ctx context.Context) core.TaskStatus {
	return r.status
}

// GetProgress 获取进度
func (r *HttpRunner) GetProgress(ctx context.Context) (float64, error) {
	// HTTP请求没有中间进度
	switch r.status {
	case core.TaskStatusPending:
		return 0.0, nil
	case core.TaskStatusRunning:
		return 0.5, nil
	case core.TaskStatusCompleted, core.TaskStatusFailed, core.TaskStatusCanceled:
		return 1.0, nil
	default:
		return 0.0, nil
	}
}

// GetMetrics 获取指标
func (r *HttpRunner) GetMetrics(ctx context.Context) map[string]interface{} {
	if r.metrics == nil {
		return make(map[string]interface{})
	}

	metrics := make(map[string]interface{})
	metrics["api_id"] = r.metrics.ApiID
	metrics["api_name"] = r.metrics.ApiName
	metrics["method"] = r.metrics.Method
	metrics["path"] = r.metrics.Path
	metrics["start_time"] = r.metrics.StartTime
	metrics["end_time"] = r.metrics.EndTime
	metrics["duration"] = r.metrics.Duration
	metrics["status_code"] = r.metrics.StatusCode
	metrics["status"] = r.metrics.Status
	metrics["request_size"] = r.metrics.RequestSize
	metrics["response_size"] = r.metrics.ResponseSize
	metrics["assertions_passed"] = r.metrics.AssertionsPassed
	metrics["assertions_failed"] = r.metrics.AssertionsFailed

	if r.metrics.Error != nil {
		metrics["error"] = r.metrics.Error.Message
	}

	return metrics
}

// Cleanup 清理资源
func (r *HttpRunner) Cleanup(ctx context.Context) error {
	// HTTP执行器不需要特殊清理
	return nil
}

func (r *HttpRunner) StoreData(ctx context.Context, data map[string]interface{}, storeConfig []*store.ReportRunData) error {
	// 1. 数据有效性检查
	if len(data) == 0 {
		logx.Info("StoreData: 没有可存储的数据")
		return nil
	}

	if len(storeConfig) == 0 {
		logx.Error(fmt.Sprintf("StoreData: 未找到存储配置，上下文信息：%v", r.contextData))
		return fmt.Errorf("未找到有效的存储配置")
	}

	// 2. 并发存储处理
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(len(storeConfig)) // 限制并发数

	// 3. 存储处理函数
	for _, config := range storeConfig {
		cfg := config // 避免闭包陷阱
		g.Go(func() error {
			start := time.Now()
			var err error

			defer func() {
				duration := time.Since(start)
				if err != nil {
					logx.Errorf("存储操作失败: 类型=%s, 耗时=%v, 错误=%v",
						cfg.Type, duration, err)
				} else {
					logx.Infof("存储操作成功: 类型=%s, 耗时=%v",
						cfg.Type, duration)
				}
			}()

			switch cfg.Type {
			case "scene":
				err = processSceneStore(data, cfg.Config.Scene)
			case "db":
				err = processDBStore(data, cfg.Config.DB)
			default:
				err = fmt.Errorf("不支持的存储类型: %s", cfg.Type)
			}

			return err
		})
	}

	// 4. 等待并返回错误
	if err := g.Wait(); err != nil {
		return fmt.Errorf("部分存储操作失败: %w", err)
	}

	return nil
}

// processSceneStore 场景存储处理
func processSceneStore(data map[string]interface{}, config *store.SceneStoreConfig) error {
	if config == nil {
		return fmt.Errorf("场景存储配置为空")
	}

	value := make(map[string]interface{})
	value[config.StepID] = data

	msgChan := config.MsgChan
	msgChan <- value

	// 实场所景存储逻辑
	return nil
}

// processDBStore 数据库存储处理
func processDBStore(data map[string]interface{}, config *store.DBStoreConfig) error {
	if config == nil {
		return fmt.Errorf("数据库存储配置为空")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	redisClient := tools.NewRedisClient(tools.RedisConfig{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	if err := redisClient.Connect(); err != nil {
		return fmt.Errorf("无法连接到 Redis: %v", err)
	}
	defer redisClient.Close()

	var errs []error

	for k, v := range data {
		switch config.DataStoreType {
		case store.RedisDataStoreTypeString:
			strVal, ok := v.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("键 %s 的值不是字符串类型", k))
				continue
			}
			if err := redisClient.Set(ctx, k, strVal, 0); err != nil {
				errs = append(errs, fmt.Errorf("存储字符串失败 %s: %v", k, err))
			}

		case store.RedisDataStoreTypeHash:
			hashVal, ok := v.(map[string]interface{})
			if !ok {
				errs = append(errs, fmt.Errorf("键 %s 的值不是哈希类型", k))
				continue
			}
			if err := redisClient.HMSet(ctx, k, hashVal); err != nil {
				errs = append(errs, fmt.Errorf("存储哈希失败 %s: %v", k, err))
			}

		case store.RedisDataStoreTypeList:
			listVal, ok := v.([]interface{})
			if !ok {
				errs = append(errs, fmt.Errorf("键 %s 的值不是列表类型", k))
				continue
			}
			if err := redisClient.LPush(ctx, k, listVal...); err != nil {
				errs = append(errs, fmt.Errorf("存储列表失败 %s: %v", k, err))
			}

		case store.RedisDataStoreTypeSet:
			setVal, ok := v.([]interface{})
			if !ok {
				errs = append(errs, fmt.Errorf("键 %s 的值不是集合类型", k))
				continue
			}
			if err := redisClient.SAdd(ctx, k, setVal...); err != nil {
				errs = append(errs, fmt.Errorf("存储集合失败 %s: %v", k, err))
			}

		case store.RedisDataStoreTypeSortedSet:
			sortedSetVal, ok := v.(map[string]float64)
			if !ok {
				errs = append(errs, fmt.Errorf("键 %s 的值不是有序集合类型", k))
				continue
			}
			var members []redis.Z
			for member, score := range sortedSetVal {
				members = append(members, redis.Z{Score: score, Member: member})
			}
			if err := redisClient.ZAdd(ctx, k, members...); err != nil {
				errs = append(errs, fmt.Errorf("存储有序集合失败 %s: %v", k, err))
			}

		default:
			errs = append(errs, fmt.Errorf("不支持的存储类型: %s", config.DataStoreType))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("存储过程中发生 %d 个错误: %v", len(errs), errs)
	}

	return nil
}
