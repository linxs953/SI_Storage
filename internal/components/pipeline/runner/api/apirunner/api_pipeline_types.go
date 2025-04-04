package api

import (
	"Storage/internal/logic/workflows/api/apirunner/dependency"
	"Storage/internal/logic/workflows/api/apirunner/expect"
	"Storage/internal/logic/workflows/api/apirunner/extract"
	"Storage/internal/logic/workflows/api/provider"
	"Storage/internal/logic/workflows/core"
	"Storage/storage"
	"context"
	"fmt"
	"time"
)

// ApiPipeline API管道实现
type ApiPipeline struct {
	*core.BasePipeline
	*ApiTaskMeta

	// API执行器
	runner ApiRunner

	// Logic提供者
	providers provider.LogicProvider
}

// API任务的spec信息
type ApiTaskMeta struct {
	// API定义
	apiDefinition *ApiDefinition

	// API指标
	metrics *ApiMetrics

	// 上下文数据
	contextData map[string]interface{}
}

// ApiDefinition API定义
// todo: 需要扩展兼容rpc / graphQL接口
type ApiDefinition struct {
	// API 唯一标识
	ApiID string `json:"api_id"`

	// API 名称
	Name string `json:"name"`

	// HTTP方法
	Method string `json:"method"`

	// 请求路径
	Path string `json:"path"`

	// 请求头
	Headers map[string]string `json:"headers,omitempty"`

	// 查询参数
	QueryParams map[string]string `json:"query_params,omitempty"`

	// 请求体类型
	BodyType string `json:"body_type,omitempty"`

	// 请求体
	Body interface{} `json:"body,omitempty"`

	// 描述
	Description string `json:"description,omitempty"`

	// 标签
	Tags []string `json:"tags,omitempty"`
}

// 依赖类型定义
const (
	DependTypeVariable = "variable" // 变量依赖
	DependTypeHeader   = "header"   // 请求头依赖
	DependTypeQuery    = "query"    // 查询参数依赖
	DependTypeBody     = "body"     // 请求体依赖
	DependTypeResponse = "response" // 响应依赖
)

// ValidationResult 验证结果
type ValidationResult struct {
	// 是否通过
	Passed bool `json:"passed"`

	// 错误数量
	ErrorCount int `json:"error_count"`

	// 断言结果
	AssertionResults []AssertionResult `json:"assertion_results"`
}

// AssertionResult 断言结果
type AssertionResult struct {
	// 断言ID
	AssertID string `json:"assert_id"`

	// 断言名称
	Name string `json:"name"`

	// 是否通过
	Passed bool `json:"passed"`

	// 实际值
	Actual interface{} `json:"actual"`

	// 期望值
	Expected interface{} `json:"expected"`

	// 错误信息
	Error string `json:"error,omitempty"`
}

// ApiMetrics API指标
type ApiMetrics struct {
	// API ID
	ApiID string `json:"api_id"`

	// API 名称
	ApiName string `json:"api_name"`

	// 请求方法
	Method string `json:"method"`

	// 请求路径
	Path string `json:"path"`

	// 开始时间
	StartTime string `json:"start_time"`

	// 结束时间
	EndTime string `json:"end_time"`

	// 执行时长（秒）
	Duration float64 `json:"duration"`

	// 请求大小（字节）
	RequestSize int64 `json:"request_size"`

	// 响应大小（字节）
	ResponseSize int64 `json:"response_size"`

	// HTTP状态码
	StatusCode int `json:"status_code"`

	// 执行状态
	Status string `json:"status"`

	// 通过的断言数
	AssertionsPassed int `json:"assertions_passed"`

	// 失败的断言数
	AssertionsFailed int `json:"assertions_failed"`

	// 错误信息
	Error *core.PipelineError `json:"error,omitempty"`

	// 自定义指标
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
}

// ReportConfig 指标上报配置
type ReportConfig struct {
	// 是否启用
	Enabled bool `json:"enabled"`

	// 上报目标
	Target string `json:"target,omitempty"`

	// 自定义配置
	Config map[string]interface{} `json:"config,omitempty"`
}

// NewApiPipeline 创建新的API管道
func NewApiPipeline(name string, description string, runner ApiRunner, provider provider.LogicProvider) *ApiPipeline {
	basePipeline := core.NewBasePipeline(name, description)
	return &ApiPipeline{
		BasePipeline: basePipeline,
		ApiTaskMeta: &ApiTaskMeta{
			apiDefinition: &ApiDefinition{},             // 初始化空的API定义
			metrics:       &ApiMetrics{},                // 初始化指标
			contextData:   make(map[string]interface{}), // 初始化上下文数据
		},
		runner:    runner,   // API执行器
		providers: provider, // Logic提供者
	}
}

// Initialize 初始化管道
func (p *ApiPipeline) Initialize(ctx context.Context) error {
	// 调用基础初始化
	if err := p.BasePipeline.Initialize(ctx); err != nil {
		return err
	}

	// 初始化API执行器
	if p.runner != nil {
		if err := p.runner.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize API runner: %w", err)
		}
	}

	// 初始化API指标
	p.metrics = &ApiMetrics{
		ApiName:   p.Name,
		StartTime: time.Now().Format(time.RFC3339),
		Status:    "pending",
	}

	// 初始化Definition
	if p.apiDefinition == nil {
		// p.apiDefinition = &ApiDefinition{}
		return fmt.Errorf("apiDefinition is nil")
	}

	apiId := p.apiDefinition.ApiID
	if apiId == "" {
		return fmt.Errorf("apiId is empty")
	}

	// 从mgodb获取数据，填充到Definition
	apiInfo, err := p.FetchApiInfoByApiId(ctx, apiId)
	if err != nil {
		return fmt.Errorf("failed to fetch API info: %w", err)
	}

	// 填充到Definition
	p.apiDefinition = &apiInfo

	return nil
}

func (p *ApiPipeline) FetchApiInfoByApiId(ctx context.Context, apiId string) (ApiDefinition, error) {
	var apiDefinition ApiDefinition
	interfaceLogic := p.providers.GetInterfaceLogic().GetDetail()
	resp, err := interfaceLogic.GetInterfaceDetail(&storage.GetInterfaceRequest{
		InterfaceId: apiId,
	})
	if err != nil {
		return apiDefinition, err
	}
	apiDefinition.ApiID = resp.Detail.ApiId
	apiDefinition.Name = resp.Detail.Name
	apiDefinition.Description = resp.Detail.Description
	apiDefinition.Method = resp.Detail.Method
	apiDefinition.Path = resp.Detail.Path

	// 设置Headers
	apiDefinition.Headers = make(map[string]string)
	for _, header := range resp.Detail.Headers {
		apiDefinition.Headers[header.Name] = header.Value
	}

	// 设置QueryParams
	apiDefinition.QueryParams = make(map[string]string)
	for _, param := range resp.Detail.Parameters {
		apiDefinition.QueryParams[param.Name] = param.Value
	}

	// 设置Body
	apiDefinition.Body = resp.Detail.RawData

	// 设置BodyType
	apiDefinition.BodyType = "application/json"
	for _, header := range resp.Detail.Headers {
		if header.Name == "Content-Type" {
			apiDefinition.BodyType = header.Value
		}
	}

	return apiDefinition, nil
}

// Execute 执行管道
func (p *ApiPipeline) Execute(ctx context.Context, spec map[string]interface{}) (map[string]interface{}, error) {
	// 记录开始时间
	startTime := time.Now()
	p.metrics.StartTime = startTime.Format(time.RFC3339)
	p.metrics.Status = "running"

	// 更新状态
	p.Status = core.TaskStatusRunning
	p.Progress = 0.1

	// 将spec转换为API定义
	apiDef, err := ConvertToApiDefinition(spec)
	if err != nil {
		p.metrics.Status = "failed"
		p.metrics.Error = &core.PipelineError{
			Message: fmt.Sprintf("Failed to convert spec to API definition: %v", err),
			Code:    "INVALID_SPEC",
		}
		return nil, err
	}

	p.apiDefinition = apiDef
	p.metrics.ApiID = apiDef.ApiID
	p.metrics.ApiName = apiDef.Name
	p.metrics.Method = apiDef.Method
	p.metrics.Path = apiDef.Path

	// 准备依赖数据
	p.Progress = 0.2
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

	dependencyValues, err := p.runner.PrepareDependencies(ctx, dependencies)
	if err != nil {
		p.metrics.Status = "failed"
		p.metrics.Error = &core.PipelineError{
			Message: fmt.Sprintf("Failed to prepare dependencies: %v", err),
			Code:    "DEPENDENCY_ERROR",
		}
		return nil, err
	}

	// 构建请求
	p.Progress = 0.4
	request, err := p.runner.BuildRequest(ctx, apiDef, dependencyValues)
	if err != nil {
		p.metrics.Status = "failed"
		p.metrics.Error = &core.PipelineError{
			Message: fmt.Sprintf("Failed to build request: %v", err),
			Code:    "BUILD_REQUEST_ERROR",
		}
		return nil, err
	}

	// 执行请求
	p.Progress = 0.6
	response, err := p.runner.ExecuteRequest(ctx, request)
	if err != nil {
		p.metrics.Status = "failed"
		p.metrics.Error = &core.PipelineError{
			Message: fmt.Sprintf("Failed to execute request: %v", err),
			Code:    "EXECUTE_REQUEST_ERROR",
		}
		return nil, err
	}

	// 处理验证
	p.Progress = 0.8

	validationResult, err := p.runner.ValidateResponse(ctx, response, &expect.AssertionGroup{})
	if err != nil {
		// 将验证错误包含到响应中，但不中断执行
		response["validation_error"] = err.Error()
	}

	var errorCount int
	if validationResult != nil {
		for _, result := range validationResult.Results {
			if !result.Passed {
				errorCount++
			}
		}
		response["validation_result"] = validationResult
		p.metrics.AssertionsPassed = len(validationResult.Results) - errorCount
		p.metrics.AssertionsFailed = errorCount
	}

	// 提取数据
	extractors := make(map[string]string)
	if ext, ok := spec["extractors"].(map[string]string); ok {
		extractors = ext
	}

	extractorList := make([]extract.Extractor, 0)
	for name, path := range extractors {
		extractorList = append(extractorList, extract.Extractor{
			JsonPath: path,
			Target:   extract.TargetValue{Type: "string", Value: name},
		})
	}

	extractedData, err := p.runner.ExtractData(ctx, response, extractorList)
	if err != nil {
		// 将提取错误包含到响应中，但不中断执行
		response["extraction_error"] = err.Error()
	}

	if len(extractedData) > 0 {
		response["extracted_data"] = extractedData
	}

	// 完成执行
	p.Progress = 1.0
	endTime := time.Now()
	p.metrics.EndTime = endTime.Format(time.RFC3339)
	p.metrics.Duration = endTime.Sub(startTime).Seconds()

	// 设置状态码
	if statusCode, ok := response["status_code"].(int); ok {
		p.metrics.StatusCode = statusCode
	}

	// 计算请求和响应大小
	if reqData, ok := request["raw_request"].(string); ok {
		p.metrics.RequestSize = int64(len(reqData))
	}

	if respData, ok := response["raw_response"].(string); ok {
		p.metrics.ResponseSize = int64(len(respData))
	}

	// 设置执行状态
	if p.metrics.Error != nil {
		p.metrics.Status = "failed"
		p.Status = core.TaskStatusFailed
	} else if p.metrics.AssertionsFailed > 0 {
		p.metrics.Status = "partially_succeeded"
		p.Status = core.TaskStatusCompleted
	} else {
		p.metrics.Status = "succeeded"
		p.Status = core.TaskStatusCompleted
	}

	// 上报指标
	reportConfig := &ReportConfig{Enabled: false}
	if cfg, ok := spec["report_config"].(*ReportConfig); ok {
		reportConfig = cfg
	}

	if reportConfig.Enabled {
		if err := p.runner.ReportMetrics(ctx, p.metrics, reportConfig); err != nil {
			// 记录错误但不中断执行
			response["metrics_report_error"] = err.Error()
		}
	}

	// 更新结果
	p.Result = response
	p.EndTime = endTime

	return response, nil
}

// GetMetrics 获取执行指标
func (p *ApiPipeline) GetMetrics(ctx context.Context) map[string]interface{} {
	baseMetrics := p.BasePipeline.GetMetrics(ctx)

	// 添加API特定指标
	if p.metrics != nil {
		baseMetrics["api_id"] = p.metrics.ApiID
		baseMetrics["api_name"] = p.metrics.ApiName
		baseMetrics["method"] = p.metrics.Method
		baseMetrics["path"] = p.metrics.Path
		baseMetrics["status_code"] = p.metrics.StatusCode
		baseMetrics["assertions_passed"] = p.metrics.AssertionsPassed
		baseMetrics["assertions_failed"] = p.metrics.AssertionsFailed
		baseMetrics["request_size"] = p.metrics.RequestSize
		baseMetrics["response_size"] = p.metrics.ResponseSize
	}

	return baseMetrics
}

// ConvertToApiDefinition 将规格转换为API定义
func ConvertToApiDefinition(spec map[string]interface{}) (*ApiDefinition, error) {
	apiDef := &ApiDefinition{}

	// 设置API ID
	if apiID, ok := spec["api_id"].(string); ok && apiID != "" {
		apiDef.ApiID = apiID
	} else {
		return nil, fmt.Errorf("missing or invalid api_id")
	}

	// 设置API名称
	if name, ok := spec["name"].(string); ok && name != "" {
		apiDef.Name = name
	} else {
		return nil, fmt.Errorf("missing or invalid name")
	}

	// 设置HTTP方法
	if method, ok := spec["method"].(string); ok && method != "" {
		apiDef.Method = method
	} else {
		return nil, fmt.Errorf("missing or invalid method")
	}

	// 设置路径
	if path, ok := spec["path"].(string); ok && path != "" {
		apiDef.Path = path
	} else {
		return nil, fmt.Errorf("missing or invalid path")
	}

	// 设置请求头
	if headers, ok := spec["headers"].(map[string]string); ok {
		apiDef.Headers = headers
	} else if headersMap, ok := spec["headers"].(map[string]interface{}); ok {
		// 转换通用接口到字符串
		headers := make(map[string]string)
		for k, v := range headersMap {
			headers[k] = fmt.Sprintf("%v", v)
		}
		apiDef.Headers = headers
	}

	// 设置查询参数
	if queryParams, ok := spec["query_params"].(map[string]string); ok {
		apiDef.QueryParams = queryParams
	} else if queryParamsMap, ok := spec["query_params"].(map[string]interface{}); ok {
		// 转换通用接口到字符串
		queryParams := make(map[string]string)
		for k, v := range queryParamsMap {
			queryParams[k] = fmt.Sprintf("%v", v)
		}
		apiDef.QueryParams = queryParams
	}

	// 设置请求体类型
	if bodyType, ok := spec["body_type"].(string); ok {
		apiDef.BodyType = bodyType
	}

	// 设置请求体
	if body, ok := spec["body"]; ok {
		apiDef.Body = body
	}

	// 设置描述
	if description, ok := spec["description"].(string); ok {
		apiDef.Description = description
	}

	// 设置标签
	if tags, ok := spec["tags"].([]string); ok {
		apiDef.Tags = tags
	} else if tagsArray, ok := spec["tags"].([]interface{}); ok {
		// 转换通用接口到字符串
		tags := make([]string, 0, len(tagsArray))
		for _, tag := range tagsArray {
			tags = append(tags, fmt.Sprintf("%v", tag))
		}
		apiDef.Tags = tags
	}

	return apiDef, nil
}
