package pipelines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"

	"Storage/internal/components/pipeline/core"
	"Storage/internal/components/tools"
	"Storage/internal/model/taskrecord"
)

// APITreeNode represents a node in the API tree structure
type APITreeNode struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"` // folder or api
	Children []*APITreeNode `json:"children,omitempty"`
	Method   string         `json:"method,omitempty"`
	Path     string         `json:"path,omitempty"`
}

// APIDetail represents detailed information about an API endpoint
type APIDetail struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Description string                 `json:"description"`
	Headers     map[string]string      `json:"headers"`
	Parameters  []Parameter            `json:"parameters"`
	Responses   interface{}            `json:"responses"`
	RawData     map[string]interface{} `json:"-"`
}

// FolderDetail represents metadata about a folder
type FolderDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parentId"`
}

// ApiFoxSyncConfig holds configuration for ApiFox synchronization
type ApiFoxSyncConfig struct {
	ProjectID   string
	SharedDocID string
	Username    string
	Password    string
	Mongo       []tools.MongoConfig
}

// ApiFoxSyncPipeline implements the Pipeline interface for ApiFox synchronization
type ApiError struct {
	DocID string
	ApiID string
	Error error
}

type ApiFoxSyncPipeline struct {
	*core.BasePipeline
	Config  ApiFoxSyncConfig
	Client  *ApiClient
	BaseURL string
	apiIds  []string // Store all extracted API IDs
	// DataChan      chan *APIDetail
	ErrorChan     chan *ApiError
	ApiIdChan     chan string
	ApiDetailChan chan *APIDetail
	mongo         []*tools.MongoClient
	// Hooks           []func(recordId string, taskId string, spec map[string]interface{}, result map[string]interface{}) error
	taskrecordModel taskrecord.TaskRecordModel
}

type ApiClient struct {
	Client  *http.Client
	Cookies []*http.Cookie
}

// NewApiFoxSyncPipeline creates a new instance of ApiFoxSyncPipeline
func NewApiFoxSyncPipeline(config ApiFoxSyncConfig, taskrecordModel taskrecord.TaskRecordModel) *ApiFoxSyncPipeline {
	// hooks := make([]func(recordId string, taskId string, spec map[string]interface{}, result map[string]interface{}) error, 0)
	return &ApiFoxSyncPipeline{
		Config:          config,
		Client:          &ApiClient{Client: &http.Client{}},
		BaseURL:         "https://apifox.com/api/v1",
		apiIds:          make([]string, 0),
		ErrorChan:       make(chan *ApiError, 100),
		ApiIdChan:       make(chan string, 100),
		ApiDetailChan:   make(chan *APIDetail, 100),
		mongo:           make([]*tools.MongoClient, 0),
		taskrecordModel: taskrecordModel,
		// Hooks:           hooks,
	}
}

// Execute implements the Pipeline interface for ApiFox synchronization
// GetExtractedApiIds returns the list of extracted API IDs
func (p *ApiFoxSyncPipeline) GetExtractedApiIds() []string {
	return p.apiIds
}

func (p *ApiFoxSyncPipeline) Execute(ctx context.Context) error {
	// 初始化 MongoDB 客户端
	p.mongo = make([]*tools.MongoClient, 0, len(p.Config.Mongo))

	// 尝试连接到每个 MongoDB 实例
	for _, mongoConfig := range p.Config.Mongo {
		logx.Infof("正在连接到 MongoDB: %s:%d", mongoConfig.MongoHost, mongoConfig.MongoPort)

		// 创建新的客户端
		client := tools.NewMongoClient(mongoConfig)

		// 尝试连接
		if err := client.Connect(); err != nil {
			logx.Errorf("连接到 MongoDB %s:%d 失败: %v",
				mongoConfig.MongoHost, mongoConfig.MongoPort, err)
			continue
		}

		// 添加到客户端列表
		p.mongo = append(p.mongo, client)
		logx.Infof("成功连接到 MongoDB: %s:%d",
			mongoConfig.MongoHost, mongoConfig.MongoPort)
	}

	// 检查是否至少有一个可用的 MongoDB 连接
	if len(p.mongo) == 0 {
		return fmt.Errorf("所有 MongoDB 连接都失败")
	}

	// Authenticate with shared document
	if err := p.authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Start the pipeline
	go p.runPipeline()
	return nil
}

func (p *ApiFoxSyncPipeline) authenticate() error {
	authURL := "https://apifox.com/api/v1/shared-doc-auth"

	// Create form data
	formData := url.Values{}
	formData.Set("id", p.Config.SharedDocID)
	formData.Set("password", "psu123456")

	// Create request
	req, err := http.NewRequest(http.MethodPost, authURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	// Send request
	logx.Infof("Authenticating with shared doc ID: %s", p.Config.SharedDocID)
	resp, err := p.Client.Client.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}
	logx.Infof("Auth response: %s", string(body))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	logx.Info("Authentication successful")
	p.Client.Cookies = resp.Cookies()
	return nil
}

func (p *ApiFoxSyncPipeline) runPipeline() {
	p.OnStart("recordId", "taskID", nil)
	p.ErrorChan = make(chan *ApiError)
	p.ApiIdChan = make(chan string)
	p.ApiDetailChan = make(chan *APIDetail)

	defer func() {
		close(p.ApiIdChan)
		close(p.ApiDetailChan)
	}()

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create errgroup with timeout context
	g, gctx := errgroup.WithContext(timeoutCtx)

	g.Go(func() error {
		p.getApifoxTree(gctx)
		return nil
	})

	g.Go(func() error {
		p.transformApiDetail(gctx)
		return nil
	})

	g.Go(func() error {
		p.storeAPI(gctx)
		p.ErrorChan <- nil
		return nil
	})

	g.Go(func() error {
		for apiId := range p.ErrorChan {
			if apiId == nil {
				return nil
			}
			p.OnError("recordId", "taskID", nil, nil)
		}
		return nil
	})

	select {
	// case <-p.ErrorChan:
	// 	logx.Errorf("Pipeline encountered an error")
	// 	p.OnError("recordId", "taskID", nil, nil)
	// 	return
	case <-gctx.Done():
		if gctx.Err() == context.DeadlineExceeded {
			logx.Error("Pipeline timed out after 5 minutes")
			p.OnError("recordId", "taskID", nil, nil)
		} else {
			logx.Error(gctx.Err())
			logx.Error("Pipeline was cancelled")
			p.OnError("recordId", "taskID", nil, nil)
		}
		return
	default:
		if err := g.Wait(); err != nil {
			logx.Errorf("Pipeline failed with error: %v", err)
			close(p.ErrorChan)
		} else {
			p.OnFinish("recordId", "taskID", nil, nil)
			logx.Info("Pipeline completed successfully")
			close(p.ErrorChan)
		}
		return
	}
}

func (p *ApiFoxSyncPipeline) getApifoxTree(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("getApifoxTree panic: %v", r)
			// 安全地写入 channel
			select {
			case p.ErrorChan <- &ApiError{
				DocID: p.Config.SharedDocID,
				Error: fmt.Errorf("getApifoxTree panic: %v", r),
			}:
			default:
				// channel 已关闭或已满，记录日志但不panic
				logx.Errorf("Failed to send error to ErrorChan: channel closed or full")
			}
		}
	}()

	apiURL := fmt.Sprintf("https://apifox.com/api/v1/shared-docs/%s/http-api-tree", p.Config.SharedDocID)
	logx.Infof("API tree URL: %s", apiURL)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("failed to create API tree request: %w", err)}
		return
	}

	// Set required headers
	req.Header.Set("x-client-version", "2.2.30")

	// Add cookies to request
	for _, cookie := range p.Client.Cookies {
		req.AddCookie(cookie)
	}

	// Send request
	logx.Infof("Fetching API tree for shared doc ID: %s", p.Config.SharedDocID)
	resp, err := p.Client.Client.Do(req)
	if err != nil {
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("API tree request failed: %w", err)}
		return
	}
	defer resp.Body.Close()

	// 在处理响应前再次检查 context
	select {
	case <-ctx.Done():
		logx.Errorf("getApifoxTree cancelled after request")
		return
	default:
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("failed to read API tree response: %w", err)}
		return
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("failed to get API tree with status %d: %s", resp.StatusCode, string(body))}
		return
	}

	// Parse response
	var apiTree map[string]interface{}
	if err := json.Unmarshal(body, &apiTree); err != nil {
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("failed to parse API tree: %w", err)}
		return
	}
	logx.Info("Successfully retrieved API tree")

	// Convert node to JSON string
	jsonStr, err := json.Marshal(apiTree)
	if err != nil {
		logx.Errorf("failed to marshal node: %v", err)
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("failed to marshal API tree: %w", err)}
		return
	}

	// Use regex to extract apiDetail.xxx
	re := regexp.MustCompile(`"apiDetail\.([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(jsonStr), -1)

	// Check if any API IDs were found
	if len(matches) == 0 {
		logx.Error("No API IDs found in API tree")
		p.ErrorChan <- &ApiError{DocID: p.Config.SharedDocID, Error: fmt.Errorf("no API IDs found in API tree")}
		return
	}

	// Extract xxx from matches and add to apiIds
	for _, match := range matches {
		// 在处理响应前再次检查 context
		select {
		case <-ctx.Done():
			logx.Errorf("getApifoxTree cancelled after request")
			return
		default:
			if len(match) > 1 {
				logx.Error(match[1])
				p.ApiIdChan <- match[1]
			}
		}
	}
}

func (p *ApiFoxSyncPipeline) transformApiDetail(ctx context.Context) {
	_, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("transformApiDetail panic: %v", r)
			p.ErrorChan <- &ApiError{Error: fmt.Errorf("transformApiDetail panic: %v", r)}
		}
		_ = cancel
		// cancel()
	}()
	// for apiId := range p.ApiIdChan {
	// 	apiDetail, err := p.fetchAPIDetail(ctx, p.Client, apiId, p.Config.SharedDocID)
	// 	if err != nil {
	// 		p.ErrorChan <- &ApiError{ApiID: apiId, Error: err}
	// 		return
	// 	}

	// 	// Send API detail to data channel
	// 	p.ApiDetailChan <- apiDetail
	// 	logx.Infof("Successfully fetched API detail: %s", apiId)
	// }

	for {
		select {
		case <-ctx.Done():
			logx.Errorf("transformApiDetail encountered an error")
			return
		case apiId, ok := <-p.ApiIdChan:
			if !ok {
				return
			}
			if apiId == "" {
				break
			}
			apiDetail, err := p.fetchAPIDetail(ctx, p.Client, apiId, p.Config.SharedDocID)
			if err != nil {
				p.ErrorChan <- &ApiError{ApiID: apiId, Error: err}
				return
			}

			// Send API detail to data channel
			p.ApiDetailChan <- apiDetail
			logx.Infof("Successfully fetched API detail: %s", apiId)
		}
	}
}

func (p *ApiFoxSyncPipeline) storeAPI(ctx context.Context) {
	if len(p.mongo) == 0 {
		logx.Error("没有配置 MongoDB 客户端")
		return
	}

	// 获取所有 MongoDB 实例的集合
	var collections []*mongo.Collection
	for i := range p.mongo {
		collection := p.mongo[i].GetCollection("apifox_apis")
		if collection == nil {
			logx.Error("获取 MongoDB 集合失败")
			continue
		}
		collections = append(collections, collection)
	}

	if len(collections) == 0 {
		logx.Error("没有可用的 MongoDB 集合")
		return
	}

	for {
		select {
		case <-ctx.Done():
			logx.Error("storeAPI context cancelled")
			return
		case apiDetail, ok := <-p.ApiDetailChan:
			if !ok {
				return
			}
			if apiDetail == nil {
				continue
			}

			// Convert APIDetail to BSON document
			doc, err := bson.Marshal(apiDetail)
			if err != nil {
				p.ErrorChan <- &ApiError{
					ApiID: apiDetail.ID,
					Error: fmt.Errorf("failed to marshal API detail to BSON: %w", err),
				}
				continue
			}

			// 使用 goroutine 异步存储数据到所有 MongoDB 实例
			go func(apiID string, document []byte) {
				// 为操作创建带超时的上下文
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// 创建用于 upsert 操作的过滤器
				filter := bson.D{{Key: "id", Value: apiID}}

				// 反序列化回 BSON 文档用于更新
				var updateDoc bson.D
				if err := bson.Unmarshal(document, &updateDoc); err != nil {
					p.ErrorChan <- &ApiError{
						ApiID: apiID,
						Error: fmt.Errorf("反序列化 BSON 文档失败: %w", err),
					}
					return
				}

				// 遍历所有集合执行 upsert 操作
				for i, collection := range collections {
					_, err = collection.UpdateOne(
						ctx,
						filter,
						bson.D{{Key: "$set", Value: updateDoc}},
						options.Update().SetUpsert(true),
					)

					if err != nil {
						p.ErrorChan <- &ApiError{
							ApiID: apiID,
							Error: fmt.Errorf("存储 API 详情到 MongoDB-%d 失败: %w", i+1, err),
						}
						continue
					}

					logx.Infof("成功存储 API 详情到 MongoDB-%d: %s", i+1, apiID)
				}
			}(apiDetail.ID, doc)
		}
	}
}

func (p *ApiFoxSyncPipeline) OnStart(recordId, taskID string, taskSpec map[string]interface{}) error {
	record := &taskrecord.TaskRecord{
		TaskID:    taskID,
		TaskType:  "sync",
		SubType:   "apifox",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "0",
		TaskSpec:  taskSpec,
		Result:    make([]map[string]interface{}, 0),
	}

	err := p.taskrecordModel.Create(context.Background(), record)
	if err != nil {
		return fmt.Errorf("failed to create task record: %w", err)
	}

	return nil
}

func (p *ApiFoxSyncPipeline) OnFinish(recordId, taskID string, taskSpec, result map[string]interface{}) error {
	update := bson.M{
		"$set": bson.M{
			"status":     "1",
			"updated_at": time.Now(),
		},
		"$push": bson.M{
			"result": result,
		},
	}

	err := p.taskrecordModel.UpdateByRecordId(context.Background(), recordId, update)
	if err != nil {
		return fmt.Errorf("failed to update task record: %w", err)
	}

	return nil
}

func (p *ApiFoxSyncPipeline) OnError(recordId, taskID string, taskSpec, errorInfo map[string]interface{}) error {
	update := bson.M{
		"$set": bson.M{
			"status":     "2",
			"updated_at": time.Now(),
		},
		"$push": bson.M{
			"result": errorInfo,
		},
	}

	err := p.taskrecordModel.UpdateByRecordId(context.Background(), recordId, update)
	if err != nil {
		return fmt.Errorf("failed to update task record: %w", err)
	}

	return nil
}

// func (p *ApiFoxSyncPipeline) getAPITree(ctx context.Context) (map[string]interface{}, error) {
// 	// Construct URL for shared document API tree
// 	apiURL := fmt.Sprintf("https://apifox.com/api/v1/shared-docs/%s/http-api-tree", p.Config.SharedDocID)
// 	logx.Infof("API tree URL: %s", apiURL)

// 	// Create request
// 	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create API tree request: %w", err)
// 	}

// 	// Set required headers
// 	req.Header.Set("x-client-version", "2.2.30")

// 	// Add cookies to request
// 	for _, cookie := range p.Client.Cookies {
// 		req.AddCookie(cookie)
// 	}

// 	// Send request
// 	logx.Infof("Fetching API tree for shared doc ID: %s", p.Config.SharedDocID)
// 	resp, err := p.Client.Client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("API tree request failed: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read response body
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read API tree response: %w", err)
// 	}

// 	// Check response status
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("failed to get API tree with status %d: %s", resp.StatusCode, string(body))
// 	}

// 	// Parse response
// 	var apiTree map[string]interface{}
// 	if err := json.Unmarshal(body, &apiTree); err != nil {
// 		return nil, fmt.Errorf("failed to parse API tree: %w", err)
// 	}

// 	logx.Info("Successfully retrieved API tree")
// 	return apiTree, nil
// }

// func (p *ApiFoxSyncPipeline) extractAndStoreAPI(ctx context.Context, node map[string]interface{}) {
// 	// var apis []*APIDetail

// 	// Skip excluded folders
// 	if node["type"] == "folder" {
// 		if node["name"] == "快速开始" || node["name"] == "开发指南" {
// 			logx.Infof("Skipping excluded folder: %s", node["name"])
// 			return
// 		}
// 	}
// 	// Extract API IDs from the node data
// 	apiIds := []string{}
// 	// Convert node to JSON string
// 	jsonStr, err := json.Marshal(node)
// 	if err != nil {
// 		logx.Errorf("failed to marshal node: %v", err)
// 		return
// 	}

// 	// Use regex to extract apiDetail.xxx
// 	re := regexp.MustCompile(`"apiDetail\.([^"]+)"`)
// 	matches := re.FindAllStringSubmatch(string(jsonStr), -1)

// 	// Extract xxx from matches and add to apiIds
// 	for _, match := range matches {
// 		if len(match) > 1 {
// 			apiIds = append(apiIds, match[1])
// 		}
// 	}
// 	// logx.Infof("Extracted API IDs: %v", apiIds)

// 	// 启动处理 goroutines
// 	errChan := make(chan error, 2)
// 	go func() {
// 		errChan <- p.storeAPIData(ctx)
// 	}()
// 	go func() {
// 		errChan <- p.extractDetail(ctx, p.Client, apiIds, p.Config.SharedDocID)
// 	}()

// 	// 等待所有 goroutine 完成或出错
// 	go func() {
// 		for err := range errChan {
// 			if err != nil {
// 				logx.Errorf("Error in goroutine: %v", err)
// 				return
// 			}
// 		}
// 	}()
// }

// func (p *ApiFoxSyncPipeline) extractDetail(ctx context.Context, client *ApiClient, apis []string, docId string) error {

// 	// Start workers for each API
// 	for _, apiID := range apis {
// 		// Fetch API detail
// 		apiDetail, err := p.fetchAPIDetail(ctx, client, apiID, docId)
// 		if err != nil {
// 			p.ErrorChan <- &ApiError{ApiID: apiID, Error: err}
// 			return err
// 		}

// 		// Send API detail to data channel
// 		p.DataChan <- apiDetail
// 		logx.Infof("Successfully fetched API detail: %s", apiID)
// 	}
// 	logx.Info("Successfully fetched API details")
// 	p.DataChan <- nil
// 	return nil
// }

func (p *ApiFoxSyncPipeline) fetchAPIDetail(ctx context.Context, client *ApiClient, apiID, docId string) (*APIDetail, error) {
	// Construct URL for API detail
	apiURL := fmt.Sprintf("https://apifox.com/api/v1/shared-docs/%s/http-apis/%s", docId, apiID)

	// Create request
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create API detail request: %w", err)
	}

	// Set required headers
	req.Header.Set("x-client-version", "2.2.30")

	// Add cookies to request
	for _, cookie := range client.Cookies {
		req.AddCookie(cookie)
	}

	// Send request
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API detail request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API detail response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get API detail with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiDetail APIDetail
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse API detail response: %w", err)
	}
	data := response["data"].(map[string]interface{})

	// Initialize headers map
	headers := make(map[string]string)
	headers["Authorization"] = ""
	headers["Content-Type"] = "application/json"

	// Extract headers from commonParameters
	if commonParams, ok := data["commonParameters"].(map[string]interface{}); ok {
		if headerList, ok := commonParams["header"].([]interface{}); ok {
			for _, h := range headerList {
				if header, ok := h.(map[string]interface{}); ok {
					if name, ok := header["name"].(string); ok {
						headers[name] = ""
					}
				}
			}
		}
	}

	// Set Content-Type from requestBody.type
	if requestBody, ok := data["requestBody"].(map[string]interface{}); ok {
		if contentType, ok := requestBody["type"].(string); ok {
			if contentType == "none" {
				headers["Content-Type"] = "application/json"
			} else {
				headers["Content-Type"] = contentType
			}
		}
	}

	// Extract parameters from requestBody.parameters
	var params []Parameter
	if requestBody, ok := data["requestBody"].(map[string]interface{}); ok {
		// Set Content-Type from requestBody.type
		if contentType, ok := requestBody["type"].(string); ok && contentType != "" {
			if contentType == "none" {
				headers["Content-Type"] = "application/json"
			} else {
				headers["Content-Type"] = contentType
			}
		}

		// Extract parameters
		if paramList, ok := requestBody["parameters"].([]interface{}); ok {
			for _, p := range paramList {
				if param, ok := p.(map[string]interface{}); ok {
					required, _ := param["required"].(bool)
					name, _ := param["name"].(string)
					pType, _ := param["type"].(string)

					params = append(params, Parameter{
						Name:     name,
						Type:     pType,
						Required: required,
					})
				}
			}
		}
	}

	// Prepare responses map
	responses := map[string]interface{}{
		"responses":        data["responses"],
		"responseExamples": data["responseExamples"],
	}

	apiId := strconv.FormatInt(int64(data["id"].(float64)), 10)
	apiDetail = APIDetail{
		ID:          apiId,
		Name:        fmt.Sprintf("%v", data["name"]),
		Method:      fmt.Sprintf("%v", data["method"]),
		Path:        fmt.Sprintf("%v", data["path"]),
		Description: fmt.Sprintf("%v", data["description"]),
		Headers:     headers,
		Parameters:  params,
		Responses:   responses,
		RawData:     data,
	}

	// Extract headers from commonParameters if available
	if commonParams, ok := data["commonParameters"].(map[string]interface{}); ok {
		// headers := make(map[string]string)
		if headerList, ok := commonParams["header"].([]interface{}); ok {
			for _, h := range headerList {
				if header, ok := h.(map[string]interface{}); ok {
					if name, ok := header["name"].(string); ok {
						headers[name] = ""
					}
				}
			}
		}
		apiDetail.Headers = headers
	}

	return &apiDetail, nil
}

// func (p *ApiFoxSyncPipeline) storeAPIData(ctx context.Context) error {
// 	for apiDetail := range p.DataChan {
// 		if apiDetail == nil {
// 			break
// 		}

// 		fileName := fmt.Sprintf("./api_data/%s.json", apiDetail.ID)
// 		jsonData, err := json.MarshalIndent(apiDetail, "", "  ")
// 		if err != nil {
// 			return fmt.Errorf("failed to marshal API detail: %w", err)
// 		}

// 		err = ioutil.WriteFile(fileName, jsonData, 0644)
// 		if err != nil {
// 			return fmt.Errorf("failed to write API detail to file: %w", err)
// 		}

// 		logx.Infof("Stored API data for ID %s in %s", apiDetail.ID, fileName)
// 	}
// 	return nil

// }
