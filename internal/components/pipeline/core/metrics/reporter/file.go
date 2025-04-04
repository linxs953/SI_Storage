package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FileReporter 文件指标上报器
type FileReporter struct {
	BaseReporter
	// 输出目录
	OutputDir string
	// 文件名前缀
	FilePrefix string
	// 是否使用格式化输出
	Pretty bool
}

// NewFileReporter 创建新的文件上报器
func NewFileReporter(outputDir, filePrefix string, pretty bool) *FileReporter {
	return &FileReporter{
		BaseReporter: BaseReporter{name: "file"},
		OutputDir:   outputDir,
		FilePrefix:  filePrefix,
		Pretty:      pretty,
	}
}

// Report 上报指标到文件
func (r *FileReporter) Report(ctx context.Context, metrics map[string]interface{}) error {
	// 确保输出目录存在
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	
	// 生成带时间戳的文件名
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s.json", r.FilePrefix, timestamp)
	filePath := filepath.Join(r.OutputDir, fileName)
	
	// 序列化指标数据
	var data []byte
	var err error
	
	if r.Pretty {
		data, err = json.MarshalIndent(metrics, "", "  ")
	} else {
		data, err = json.Marshal(metrics)
	}
	
	if err != nil {
		return fmt.Errorf("序列化指标数据失败: %w", err)
	}
	
	// 写入文件
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入指标文件失败: %w", err)
	}
	
	log.Printf("[指标] 已写入文件: %s", filePath)
	return nil
}

// Close 实现MetricsReporter接口
func (r *FileReporter) Close() error {
	return nil  // 文件上报器无需关闭资源
}
