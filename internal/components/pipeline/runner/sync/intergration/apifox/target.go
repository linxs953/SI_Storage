package apifox

import (
	"Storage/internal/logic/workflows/sync"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileSystemTargetConfig 文件系统目标配置
type FileSystemTargetConfig struct {
	// 输出目录
	OutputDir string `json:"output_dir"`

	// 文件名前缀
	FilePrefix string `json:"file_prefix"`

	// 是否生成摘要
	GenerateSummary bool `json:"generate_summary"`

	// 是否包含时间戳
	IncludeTimestamp bool `json:"include_timestamp"`

	// 是否格式化JSON
	PrettyJSON bool `json:"pretty_json"`
}

// FileSystemTarget 文件系统目标实现
type FileSystemTarget struct {
	// 配置
	config FileSystemTargetConfig

	// 状态
	status sync.TargetStatus

	// 写入的文件路径
	writtenFiles []string
}

// NewFileSystemTarget 创建文件系统目标
func NewFileSystemTarget() *FileSystemTarget {
	return &FileSystemTarget{
		status: sync.TargetStatus{
			Connected: false,
		},
		writtenFiles: make([]string, 0),
	}
}

// Connect 连接到文件系统
func (t *FileSystemTarget) Connect(ctx context.Context, config map[string]interface{}) error {
	// 解析配置
	if err := t.parseConfig(config); err != nil {
		t.status.Error = err.Error()
		return err
	}

	// 检查目录是否存在
	if _, err := os.Stat(t.config.OutputDir); os.IsNotExist(err) {
		// 创建目录
		if err := os.MkdirAll(t.config.OutputDir, 0755); err != nil {
			t.status.Error = fmt.Sprintf("创建输出目录失败: %v", err)
			return err
		}
	}

	t.status.Connected = true
	t.status.LastWrite = time.Now()
	return nil
}

// parseConfig 解析配置
func (t *FileSystemTarget) parseConfig(config map[string]interface{}) error {
	// 读取输出目录
	if outputDir, ok := config["output_dir"].(string); ok && outputDir != "" {
		t.config.OutputDir = outputDir
	} else {
		t.config.OutputDir = "./api_data"
	}

	// 读取文件前缀
	if filePrefix, ok := config["file_prefix"].(string); ok {
		t.config.FilePrefix = filePrefix
	} else {
		t.config.FilePrefix = "apifox"
	}

	// 读取是否生成摘要
	if generateSummary, ok := config["generate_summary"].(bool); ok {
		t.config.GenerateSummary = generateSummary
	} else {
		t.config.GenerateSummary = true
	}

	// 读取是否包含时间戳
	if includeTimestamp, ok := config["include_timestamp"].(bool); ok {
		t.config.IncludeTimestamp = includeTimestamp
	} else {
		t.config.IncludeTimestamp = true
	}

	// 读取是否格式化JSON
	if prettyJSON, ok := config["pretty_json"].(bool); ok {
		t.config.PrettyJSON = prettyJSON
	} else {
		t.config.PrettyJSON = true
	}

	return nil
}

// Write 写入数据到文件系统
func (t *FileSystemTarget) Write(ctx context.Context, data interface{}, options map[string]interface{}) error {
	// 确保已连接
	if !t.status.Connected {
		return fmt.Errorf("未连接到文件系统")
	}

	// 检查数据类型
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("数据类型不正确")
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102_150405")
	var filename string
	if t.config.IncludeTimestamp {
		filename = fmt.Sprintf("%s_%s.json", t.config.FilePrefix, timestamp)
	} else {
		filename = fmt.Sprintf("%s.json", t.config.FilePrefix)
	}
	filePath := filepath.Join(t.config.OutputDir, filename)

	// 将数据转换为JSON
	var jsonData []byte
	var err error
	if t.config.PrettyJSON {
		jsonData, err = json.MarshalIndent(dataMap, "", "  ")
	} else {
		jsonData, err = json.Marshal(dataMap)
	}
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入JSON文件失败: %w", err)
	}
	t.writtenFiles = append(t.writtenFiles, filePath)

	// 生成摘要
	if t.config.GenerateSummary {
		if err := t.generateSummary(dataMap, timestamp); err != nil {
			return err
		}
	}

	// 更新状态
	t.status.LastWrite = time.Now()
	t.status.WrittenCount = int64(len(t.writtenFiles))

	return nil
}

// generateSummary 生成摘要
func (t *FileSystemTarget) generateSummary(data map[string]interface{}, timestamp string) error {
	// 检查API详情
	apiDetails, ok := data["api_details"]
	if !ok {
		return fmt.Errorf("数据中缺少API详情")
	}

	// 生成摘要文件名
	var summaryFilename string
	if t.config.IncludeTimestamp {
		summaryFilename = fmt.Sprintf("%s_%s_summary.md", t.config.FilePrefix, timestamp)
	} else {
		summaryFilename = fmt.Sprintf("%s_summary.md", t.config.FilePrefix)
	}
	summaryPath := filepath.Join(t.config.OutputDir, summaryFilename)

	// 创建摘要文件
	file, err := os.Create(summaryPath)
	if err != nil {
		return fmt.Errorf("创建摘要文件失败: %w", err)
	}
	defer file.Close()

	// 写入标题
	fmt.Fprintf(file, "# ApiFox API 摘要\n\n")
	fmt.Fprintf(file, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// 写入概述
	if sourceDocID, ok := data["shared_doc_id"].(string); ok {
		fmt.Fprintf(file, "共享文档ID: %s\n", sourceDocID)
	}
	if apiCount, ok := data["api_count"].(int); ok {
		fmt.Fprintf(file, "API总数: %d\n\n", apiCount)
	}

	// 写入文件夹结构
	fmt.Fprintf(file, "## 文件夹结构\n\n")
	if folderStructure, ok := data["folder_structure"].([]*FolderDetail); ok {
		for _, folder := range folderStructure {
			t.writeFolderSummary(file, folder, 0)
		}
	}

	// 写入API列表
	fmt.Fprintf(file, "## API列表\n\n")
	if details, ok := apiDetails.(map[string]*APIDetail); ok {
		for _, api := range details {
			fmt.Fprintf(file, "### %s\n\n", api.Name)
			fmt.Fprintf(file, "- **ID**: %s\n", api.ID)
			fmt.Fprintf(file, "- **方法**: %s\n", api.Method)
			fmt.Fprintf(file, "- **路径**: %s\n", api.Path)
			fmt.Fprintf(file, "- **状态**: %s\n", api.Status)
			if api.Description != "" {
				fmt.Fprintf(file, "- **描述**: %s\n", api.Description)
			}
			fmt.Fprintf(file, "\n")
		}
	}

	t.writtenFiles = append(t.writtenFiles, summaryPath)
	return nil
}

// writeFolderSummary 写入文件夹摘要
func (t *FileSystemTarget) writeFolderSummary(file *os.File, folder *FolderDetail, level int) {
	// 写入标题
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	fmt.Fprintf(file, "%s- **%s**\n", indent, folder.Name)

	// 写入API
	for _, api := range folder.APIs {
		fmt.Fprintf(file, "%s  - %s %s\n", indent, api.Method, api.Path)
	}

	// 递归写入子文件夹
	for _, subFolder := range folder.Folders {
		t.writeFolderSummary(file, subFolder, level+1)
	}
}

// GetStatus 获取目标状态
func (t *FileSystemTarget) GetStatus(ctx context.Context) sync.TargetStatus {
	return t.status
}

// Close 关闭连接
func (t *FileSystemTarget) Close(ctx context.Context) error {
	t.status.Connected = false
	return nil
}
