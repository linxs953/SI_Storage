package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ConsoleReporter 控制台指标上报器
type ConsoleReporter struct {
	BaseReporter
	// 是否使用格式化输出
	Pretty bool
}

// NewConsoleReporter 创建新的控制台上报器
func NewConsoleReporter(pretty bool) *ConsoleReporter {
	return &ConsoleReporter{
		BaseReporter: BaseReporter{name: "console"},
		Pretty:       pretty,
	}
}

// Report 上报指标到控制台
func (r *ConsoleReporter) Report(ctx context.Context, metrics map[string]interface{}) error {
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
	
	log.Printf("[指标] %s", string(data))
	return nil
}

// Close 实现MetricsReporter接口
func (r *ConsoleReporter) Close() error {
	return nil  // 控制台上报器无需关闭资源
}
