package expect

import "Storage/internal/logic/workflows/api/apirunner/dependency"

// AssertionType defines the type of assertion to perform
type AssertionType string

const (
	AssertEqual            AssertionType = "equal"             // Exact equality
	AssertNotEqual         AssertionType = "not_equal"         // Not equal
	AssertContains         AssertionType = "contains"          // String/array contains
	AssertNotContains      AssertionType = "not_contains"      // String/array does not contain
	AssertGreaterThan      AssertionType = "greater_than"      // Numeric comparison >
	AssertLessThan         AssertionType = "less_than"         // Numeric comparison <
	AssertGreaterOrEqual   AssertionType = "greater_or_equal"  // Numeric comparison >=
	AssertLessOrEqual      AssertionType = "less_or_equal"     // Numeric comparison <=
	AssertRegexMatch       AssertionType = "regex_match"       // Regular expression match
	AssertLengthEqual      AssertionType = "length_equal"      // Array/string length equal
	AssertLengthGreater    AssertionType = "length_greater"    // Array/string length >
	AssertLengthLess       AssertionType = "length_less"       // Array/string length <
	AssertTypeMatch        AssertionType = "type_match"        // Type assertion
	AssertHasField         AssertionType = "has_field"         // Field existence
	AssertCustomValidation AssertionType = "custom_validation" // Custom validation function
)

// RetryStrategy 定义重试策略类型
type RetryStrategy string

const (
	// RetryConstant 固定间隔重试
	RetryConstant RetryStrategy = "constant"

	// RetryLinearBackoff 线性递增间隔，每次重试间隔增加一个固定值
	// 例如：1s, 2s, 3s, 4s...
	RetryLinearBackoff RetryStrategy = "linear_backoff"

	// RetryExponentialBackoff 指数递增间隔，每次重试间隔翻倍
	// 例如：1s, 2s, 4s, 8s...
	RetryExponentialBackoff RetryStrategy = "exponential_backoff"

	// RetryFibonacciBackoff 斐波那契数列间隔
	// 例如：1s, 1s, 2s, 3s, 5s, 8s...
	RetryFibonacciBackoff RetryStrategy = "fibonacci_backoff"

	// RetryRandomBackoff 随机间隔，在指定范围内随机选择间隔
	// 适用于需要避免集中重试的场景
	RetryRandomBackoff RetryStrategy = "random_backoff"
)

// Assertion defines a single assertion to be performed on API response
type Assertion struct {
	// 断言名称
	Name string `json:"name"`

	// 断言类型
	Type AssertionType `json:"type"`

	// JsonPath表达式，用于从响应中提取要断言的字段
	JsonPath string `json:"json_path"`

	// 实际值
	ActualValue interface{} `json:"actual_value"`

	// 预期值，支持以下几种形式：
	// 1. 直接值：直接指定预期值
	// 2. 依赖注入：使用Dependency对象指定数据来源
	ExpectedValue interface{} `json:"expected_value"`

	// 依赖配置，用于注入预期值
	Dependency *dependency.Dependency `json:"dependency,omitempty"`

	// 断言失败时的错误消息模板
	ErrorTemplate string `json:"error_template,omitempty"`

	// 断言选项
	Options AssertionOptions `json:"options,omitempty"`
}

// AssertionOptions provides configuration options for assertions
type AssertionOptions struct {
	// 是否忽略大小写（用于字符串比较）
	IgnoreCase bool `json:"ignore_case,omitempty"`

	// 数值比较时的容差范围
	Tolerance float64 `json:"tolerance,omitempty"`

	// 是否进行深度比较（用于对象比较）
	DeepComparison bool `json:"deep_comparison,omitempty"`

	// 类型转换配置
	TypeConversion TypeConversion `json:"type_conversion,omitempty"`
}

// TypeConversion provides configuration for type conversion before assertion
type TypeConversion struct {
	// 是否启用类型转换
	Enable bool `json:"enable"`

	// 目标类型
	TargetType string `json:"target_type,omitempty"`

	// 日期时间格式化模板
	TimeFormat string `json:"time_format,omitempty"`
}

// AssertionResult represents the result of an assertion
type AssertionResult struct {
	// 断言名称
	Name string `json:"name"`

	// 是否通过
	Passed bool `json:"passed"`

	// 实际值
	ActualValue interface{} `json:"actual_value,omitempty"`

	// 预期值
	ExpectedValue interface{} `json:"expected_value,omitempty"`

	// 错误信息
	Error string `json:"error,omitempty"`

	// 断言执行的详细信息
	Details map[string]interface{} `json:"details,omitempty"`
}

// AssertionGroupResult represents the result of executing an assertion group
type AssertionGroupResult struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Passed      bool               `json:"passed"`
	Results     []*AssertionResult `json:"results"`
	Parallel    bool               `json:"parallel,omitempty"`
}

// AssertionGroup represents a group of related assertions
type AssertionGroup struct {
	// 分组名称
	Name string `json:"name"`

	// 分组描述
	Description string `json:"description,omitempty"`

	// 断言列表
	Assertions []Assertion `json:"assertions"`

	// 分组选项
	Options GroupOptions `json:"options,omitempty"`
}

// GroupOptions provides configuration options for assertion groups
type GroupOptions struct {
	// 是否在第一个断言失败时停止
	StopOnFirstFailure bool `json:"stop_on_first_failure,omitempty"`

	// 分组执行超时时间（秒）
	Timeout int `json:"timeout,omitempty"`

	// 是否并行执行断言
	Parallel bool `json:"parallel,omitempty"`

	// 重试配置
	Retry *RetryConfig `json:"retry,omitempty"`
}

// RetryConfig provides configuration for assertion retries
type RetryConfig struct {
	// 最大重试次数
	MaxRetries int `json:"max_retries"`

	// 基础间隔（毫秒）
	// 对于 constant 策略，这是固定的间隔
	// 对于其他策略，这是初始间隔
	Interval int `json:"interval"`

	// 重试策略
	Strategy RetryStrategy `json:"strategy,omitempty"`

	// 最大间隔（毫秒）
	// 限制指数和斐波那契策略的间隔增长
	MaxInterval int `json:"max_interval,omitempty"`

	// 随机间隔范围（毫秒）
	// 只在 random_backoff 策略下使用
	RandomRange *struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"random_range,omitempty"`
}

// NewAssertion creates a new Assertion with the given parameters
func NewAssertion(name string, assertType AssertionType, jsonPath string, actualValue interface{}, expectedValue interface{}) *Assertion {
	return &Assertion{
		Name:          name,
		Type:          assertType,
		JsonPath:      jsonPath,
		ActualValue:   actualValue,
		ExpectedValue: expectedValue,
		Options: AssertionOptions{
			DeepComparison: true,
		},
	}
}

// NewAssertionGroup creates a new AssertionGroup with the given parameters
func NewAssertionGroup(name string, assertions ...Assertion) *AssertionGroup {
	return &AssertionGroup{
		Name:       name,
		Assertions: assertions,
		Options: GroupOptions{
			StopOnFirstFailure: true, // 默认失败即停止
		},
	}
}

// WithDependency adds a dependency to the assertion
func (a *Assertion) WithDependency(dep *dependency.Dependency) *Assertion {
	a.Dependency = dep
	a.ExpectedValue = nil // 清空直接值，避免混淆
	return a
}

// WithOptions sets the options for the assertion
func (a *Assertion) WithOptions(options AssertionOptions) *Assertion {
	a.Options = options
	return a
}

// WithErrorTemplate sets the error template for the assertion
func (a *Assertion) WithErrorTemplate(template string) *Assertion {
	a.ErrorTemplate = template
	return a
}

// WithGroupOptions sets the options for the assertion group
func (g *AssertionGroup) WithGroupOptions(options GroupOptions) *AssertionGroup {
	g.Options = options
	return g
}

// WithRetry adds retry configuration to the assertion group
func (g *AssertionGroup) WithRetry(maxRetries int, interval int, strategy RetryStrategy) *AssertionGroup {
	g.Options.Retry = &RetryConfig{
		MaxRetries: maxRetries,
		Interval:   interval,
		Strategy:   strategy,
	}
	return g
}

// WithDescription adds a description to the assertion group
func (g *AssertionGroup) WithDescription(description string) *AssertionGroup {
	g.Description = description
	return g
}

// AddAssertion adds a new assertion to the group
func (g *AssertionGroup) AddAssertion(assertion Assertion) *AssertionGroup {
	g.Assertions = append(g.Assertions, assertion)
	return g
}

// WithParallelExecution enables parallel execution for the assertion group
func (g *AssertionGroup) WithParallelExecution() *AssertionGroup {
	g.Options.Parallel = true
	g.Options.StopOnFirstFailure = false // 并行执行时不应该立即停止
	return g
}

// WithTimeout sets the timeout for the assertion group
func (g *AssertionGroup) WithTimeout(timeoutSeconds int) *AssertionGroup {
	g.Options.Timeout = timeoutSeconds
	return g
}

// Common assertion creation helpers

// NewEqualAssertion creates an equality assertion
func NewEqualAssertion(name, jsonPath string, actualValue interface{}, expectedValue interface{}) *Assertion {
	return NewAssertion(name, AssertEqual, jsonPath, actualValue, expectedValue)
}

// NewNotEqualAssertion creates a non-equality assertion
func NewNotEqualAssertion(name, jsonPath string, actualValue interface{}, expectedValue interface{}) *Assertion {
	return NewAssertion(name, AssertNotEqual, jsonPath, actualValue, expectedValue)
}

// NewContainsAssertion creates a contains assertion
func NewContainsAssertion(name, jsonPath string, actualValue interface{}, expectedValue interface{}) *Assertion {
	return NewAssertion(name, AssertContains, jsonPath, actualValue, expectedValue)
}

// NewGreaterThanAssertion creates a greater than assertion
func NewGreaterThanAssertion(name, jsonPath string, actualValue interface{}, expectedValue interface{}) *Assertion {
	return NewAssertion(name, AssertGreaterThan, jsonPath, actualValue, expectedValue)
}

// NewHasFieldAssertion creates a field existence assertion
func NewHasFieldAssertion(name, jsonPath string, actualValue interface{}) *Assertion {
	return NewAssertion(name, AssertHasField, jsonPath, actualValue, nil)
}

// NewRegexMatchAssertion creates a regex match assertion
func NewRegexMatchAssertion(name, jsonPath string, actualValue interface{}, pattern string) *Assertion {
	return NewAssertion(name, AssertRegexMatch, jsonPath, actualValue, pattern)
}

// NewLengthEqualAssertion creates a length equality assertion
func NewLengthEqualAssertion(name, jsonPath string, actualValue interface{}, length int) *Assertion {
	return NewAssertion(name, AssertLengthEqual, jsonPath, actualValue, length)
}
