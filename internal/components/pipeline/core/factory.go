package core

import (
	"errors"
)

// PipelineType 管道类型
type PipelineType string

const (
	// TypeAPI API管道类型
	TypeAPI PipelineType = "api"

	// TypeApiFox ApiFox同步管道类型
	TypeApiFox PipelineType = "apifox"
)

// PipelineConfig 管道配置
type PipelineConfig struct {
	// 管道类型
	Type PipelineType `json:"type"`

	// 管道名称
	Name string `json:"name"`

	// 管道描述
	Description string `json:"description"`

	// 管道特定配置
	Config map[string]interface{} `json:"config"`
}

// PipelineFactory 管道工厂
type PipelineFactory interface {
	// Create 创建管道
	Create(config *PipelineConfig) (PipelineRunner, error)
}

// DefaultPipelineFactory 默认管道工厂实现
type DefaultPipelineFactory struct {
	// 管道创建函数映射
	creators map[PipelineType]func(*PipelineConfig) (PipelineRunner, error)
}

// NewPipelineFactory 创建新的管道工厂
func NewPipelineFactory() *DefaultPipelineFactory {
	return &DefaultPipelineFactory{
		creators: make(map[PipelineType]func(*PipelineConfig) (PipelineRunner, error)),
	}
}

// Register 注册管道创建函数
func (f *DefaultPipelineFactory) Register(pipelineType PipelineType, creator func(*PipelineConfig) (PipelineRunner, error)) {
	f.creators[pipelineType] = creator
}

// Create 创建管道
func (f *DefaultPipelineFactory) Create(config *PipelineConfig) (PipelineRunner, error) {
	// 检查是否注册了创建函数
	creator, ok := f.creators[config.Type]
	if !ok {
		return nil, errors.New("unsupported pipeline type: " + string(config.Type))
	}

	// 调用创建函数
	return creator(config)
}

// CreateBasePipeline 创建基础管道
func CreateBasePipeline(config *PipelineConfig) (PipelineRunner, error) {
	return NewBasePipeline(config.Name, config.Description), nil
}
