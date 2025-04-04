package apifox

import (
	"Storage/internal/components/pipeline/runner/sync"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Constants for ApiFox integration
const (
	ApiFoxAuthEndpoint      = "https://apifox.com/api/v1/shared-doc-auth"
	ApiFoxAPITreeEndpoint   = "https://apifox.com/api/v1/shared-docs/%s/http-api-tree"
	ApiFoxAPIDetailEndpoint = "https://apifox.com/api/v1/shared-docs/%s/apis/%s"
)

// ApiFoxSourceConfig ApiFox数据源配置
type ApiFoxSourceConfig struct {
	// 共享文档ID
	SharedDocID string `json:"shared_doc_id"`

	// 文档密码
	Password string `json:"password,omitempty"`

	// 过滤的文件夹名称列表
	ExcludeFolders []string `json:"exclude_folders,omitempty"`

	// 输出目录
	OutputDir string `json:"output_dir"`

	// 是否生成摘要
	GenerateSummary bool `json:"generate_summary"`
}

// ApiFoxSource ApiFox数据源实现
type ApiFoxSource struct {
	*sync.ApiDocSource

	// 配置
	config ApiFoxSourceConfig

	// 访问令牌
	token string

	// API树
	apiTree []*APITreeNode

	// 文件夹结构
	folderStructure []*FolderDetail
}

// NewApiFoxSource 创建ApiFox数据源
func NewApiFoxSource() *ApiFoxSource {
	return &ApiFoxSource{
		ApiDocSource: &sync.ApiDocSource{
			client: &http.Client{
				Timeout: 30 * time.Second,
			},
			status: sync.SourceStatus{
				Connected: false,
			},
			apiDetails: make(map[string]*APIDetail),
		},
	}
}

// Connect 连接到ApiFox
func (s *ApiFoxSource) Connect(ctx context.Context, config map[string]interface{}) error {
	// 解析配置
	if err := s.parseConfig(config); err != nil {
		s.status.Error = err.Error()
		return err
	}

	// 认证
	if err := s.authenticate(ctx); err != nil {
		s.status.Error = err.Error()
		return err
	}

	s.status.Connected = true
	s.status.LastSync = time.Now()
	return nil
}

// parseConfig 解析配置
func (s *ApiFoxSource) parseConfig(config map[string]interface{}) error {
	// 读取共享文档ID
	if docID, ok := config["shared_doc_id"].(string); ok && docID != "" {
		s.config.SharedDocID = docID
	} else {
		return fmt.Errorf("必须提供共享文档ID")
	}

	// 读取密码
	if password, ok := config["password"].(string); ok {
		s.config.Password = password
	}

	// 读取排除文件夹
	if excludeFolders, ok := config["exclude_folders"].([]string); ok {
		s.config.ExcludeFolders = excludeFolders
	} else if excludeFoldersArray, ok := config["exclude_folders"].([]interface{}); ok {
		folders := make([]string, 0, len(excludeFoldersArray))
		for _, folder := range excludeFoldersArray {
			if folderName, ok := folder.(string); ok {
				folders = append(folders, folderName)
			}
		}
		s.config.ExcludeFolders = folders
	}

	// 读取输出目录
	if outputDir, ok := config["output_dir"].(string); ok && outputDir != "" {
		s.config.OutputDir = outputDir
	} else {
		s.config.OutputDir = "./api_data"
	}

	// 读取是否生成摘要
	if generateSummary, ok := config["generate_summary"].(bool); ok {
		s.config.GenerateSummary = generateSummary
	} else {
		s.config.GenerateSummary = true
	}

	return nil
}

// authenticate 认证共享文档
func (s *ApiFoxSource) authenticate(ctx context.Context) error {
	// 准备认证表单
	form := url.Values{}
	form.Add("id", s.config.SharedDocID)
	if s.config.Password != "" {
		form.Add("password", s.config.Password)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", ApiFoxAuthEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("创建认证请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送认证请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取认证响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("认证失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
		Msg string `json:"msg"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析认证响应失败: %w", err)
	}

	// 检查响应码
	if response.Code != 0 {
		return fmt.Errorf("认证失败, 错误码: %d, 消息: %s", response.Code, response.Msg)
	}

	// 保存令牌
	s.token = response.Data.Token
	return nil
}

// Fetch 获取API数据
func (s *ApiFoxSource) Fetch(ctx context.Context, options map[string]interface{}) (interface{}, error) {
	// 确保已连接
	if !s.status.Connected {
		return nil, fmt.Errorf("未连接到ApiFox")
	}

	// 获取API树
	if err := s.fetchAPITree(ctx); err != nil {
		return nil, err
	}

	// 获取API详情
	if err := s.fetchAPIDetails(ctx); err != nil {
		return nil, err
	}

	// 构建文件夹结构
	s.buildFolderStructure()

	// 返回API数据
	return map[string]interface{}{
		"api_tree":         s.apiTree,
		"api_details":      s.apiDetails,
		"folder_structure": s.folderStructure,
	}, nil
}

// fetchAPITree 获取API树
func (s *ApiFoxSource) fetchAPITree(ctx context.Context) error {
	// 构建URL
	url := fmt.Sprintf(ApiFoxAPITreeEndpoint, s.config.SharedDocID)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建API树请求失败: %w", err)
	}

	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+s.token)

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送API树请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取API树响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取API树失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code int            `json:"code"`
		Data []*APITreeNode `json:"data"`
		Msg  string         `json:"msg"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析API树响应失败: %w", err)
	}

	// 检查响应码
	if response.Code != 0 {
		return fmt.Errorf("获取API树失败, 错误码: %d, 消息: %s", response.Code, response.Msg)
	}

	// 保存API树
	s.apiTree = response.Data
	s.status.ItemCount = s.countAPIs(s.apiTree)

	return nil
}

// countAPIs 计算API节点数量
func (s *ApiFoxSource) countAPIs(nodes []*APITreeNode) int64 {
	var count int64
	for _, node := range nodes {
		if node.Type == "api" {
			count++
		}
		if len(node.Children) > 0 {
			count += s.countAPIs(node.Children)
		}
	}
	return count
}

// fetchAPIDetails 获取API详情
func (s *ApiFoxSource) fetchAPIDetails(ctx context.Context) error {
	// 收集所有API ID
	apiIDs := s.collectAPIIDs(s.apiTree)

	// 获取每个API的详细信息
	for _, apiID := range apiIDs {
		detail, err := s.fetchSingleAPIDetail(ctx, apiID)
		if err != nil {
			return err
		}
		s.apiDetails[apiID] = detail
	}

	return nil
}

// collectAPIIDs 收集所有API ID
func (s *ApiFoxSource) collectAPIIDs(nodes []*APITreeNode) []string {
	var ids []string
	for _, node := range nodes {
		if node.Type == "api" {
			ids = append(ids, node.ID)
		}
		if len(node.Children) > 0 {
			ids = append(ids, s.collectAPIIDs(node.Children)...)
		}
	}
	return ids
}

// fetchSingleAPIDetail 获取单个API详情
func (s *ApiFoxSource) fetchSingleAPIDetail(ctx context.Context, apiID string) (*APIDetail, error) {
	// 构建URL
	url := fmt.Sprintf(ApiFoxAPIDetailEndpoint, s.config.SharedDocID, apiID)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建API详情请求失败: %w", err)
	}

	// 设置认证头
	req.Header.Set("Authorization", "Bearer "+s.token)

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送API详情请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取API详情响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取API详情失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response struct {
		Code int       `json:"code"`
		Data APIDetail `json:"data"`
		Msg  string    `json:"msg"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析API详情响应失败: %w", err)
	}

	// 检查响应码
	if response.Code != 0 {
		return nil, fmt.Errorf("获取API详情失败, 错误码: %d, 消息: %s", response.Code, response.Msg)
	}

	return &response.Data, nil
}

// buildFolderStructure 构建文件夹结构
func (s *ApiFoxSource) buildFolderStructure() {
	s.folderStructure = s.convertNodesToFolders(s.apiTree)
}

// convertNodesToFolders 将树节点转换为文件夹结构
func (s *ApiFoxSource) convertNodesToFolders(nodes []*APITreeNode) []*FolderDetail {
	result := make([]*FolderDetail, 0, len(nodes))

	for _, node := range nodes {
		// 检查是否排除该文件夹
		if node.Type == "folder" && s.shouldExcludeFolder(node.Name) {
			continue
		}

		if node.Type == "folder" {
			folder := &FolderDetail{
				ID:      node.ID,
				Name:    node.Name,
				APIs:    make([]*APIDetail, 0),
				Folders: make([]*FolderDetail, 0),
			}

			// 处理子节点
			if len(node.Children) > 0 {
				// 收集子API和子文件夹
				for _, child := range node.Children {
					if child.Type == "api" {
						if detail, ok := s.apiDetails[child.ID]; ok {
							folder.APIs = append(folder.APIs, detail)
						}
					}
				}

				// 递归处理子文件夹
				folder.Folders = s.convertNodesToFolders(node.Children)
			}

			result = append(result, folder)
		}
	}

	return result
}

// shouldExcludeFolder 判断是否应该排除文件夹
func (s *ApiFoxSource) shouldExcludeFolder(folderName string) bool {
	for _, excludeFolder := range s.config.ExcludeFolders {
		if folderName == excludeFolder {
			return true
		}
	}
	return false
}

// Transform 转换API数据
func (s *ApiFoxSource) Transform(ctx context.Context, data interface{}, options map[string]interface{}) (interface{}, error) {
	// 检查数据类型
	dataMap, ok := data.(map[string]interface{})
	_ = dataMap
	if !ok {
		return nil, fmt.Errorf("数据类型不正确")
	}

	// 补充默认值
	for _, detail := range s.apiDetails {
		// 添加默认Content-Type
		hasContentType := false
		if detail.Headers != nil {
			for _, header := range detail.Headers {
				if name, ok := header["name"].(string); ok && strings.ToLower(name) == "content-type" {
					hasContentType = true
					break
				}
			}
		} else {
			detail.Headers = make([]map[string]interface{}, 0)
		}

		if !hasContentType {
			detail.Headers = append(detail.Headers, map[string]interface{}{
				"name":  "Content-Type",
				"value": "application/json",
			})
		}
	}

	// 返回处理后的数据
	transformedData := map[string]interface{}{
		"source":           "ApiFox",
		"shared_doc_id":    s.config.SharedDocID,
		"timestamp":        time.Now().Format(time.RFC3339),
		"api_count":        len(s.apiDetails),
		"folder_structure": s.folderStructure,
		"api_details":      s.apiDetails,
	}

	return transformedData, nil
}

// GetStatus 获取数据源状态
func (s *ApiFoxSource) GetStatus(ctx context.Context) sync.SourceStatus {
	return s.status
}

// Close 关闭连接
func (s *ApiFoxSource) Close(ctx context.Context) error {
	s.status.Connected = false
	return nil
}

// // APITreeNode API树节点
// type APITreeNode struct {
// 	// 节点ID
// 	ID string `json:"id"`

// 	// 节点名称
// 	Name string `json:"name"`

// 	// 节点类型 (folder/api)
// 	Type string `json:"type"`

// 	// HTTP方法
// 	Method string `json:"method,omitempty"`

// 	// API路径
// 	Path string `json:"path,omitempty"`

// 	// 子节点
// 	Children []*APITreeNode `json:"children,omitempty"`
// }

// // APIDetail API详情
// type APIDetail struct {
// 	// API ID
// 	ID string `json:"id"`

// 	// API名称
// 	Name string `json:"name"`

// 	// HTTP方法
// 	Method string `json:"method"`

// 	// 路径
// 	Path string `json:"path"`

// 	// 状态
// 	Status string `json:"status"`

// 	// 描述
// 	Description string `json:"description,omitempty"`

// 	// 请求头
// 	Headers []map[string]interface{} `json:"headers,omitempty"`

// 	// 请求参数
// 	Parameters []map[string]interface{} `json:"parameters,omitempty"`

// 	// 请求体
// 	RequestBody map[string]interface{} `json:"requestBody,omitempty"`

// 	// 响应
// 	Responses []map[string]interface{} `json:"responses,omitempty"`
// }

// // FolderDetail 文件夹详情
// type FolderDetail struct {
// 	// 文件夹ID
// 	ID string `json:"id"`

// 	// 文件夹名称
// 	Name string `json:"name"`

// 	// 子API列表
// 	APIs []*APIDetail `json:"apis,omitempty"`

// 	// 子文件夹
// 	Folders []*FolderDetail `json:"folders,omitempty"`
// }
