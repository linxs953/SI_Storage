package expect

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// AssertAll executes all assertions in the group and returns the results
func (g *AssertionGroup) AssertAll() *AssertionGroupResult {
	result := &AssertionGroupResult{
		Name:        g.Name,
		Description: g.Description,
		Results:     make([]*AssertionResult, 0, len(g.Assertions)),
		Parallel:    g.Options.Parallel,
	}

	// 如果启用了并行执行
	if g.Options.Parallel {
		resultChan := make(chan *AssertionResult, len(g.Assertions))
		for _, assertion := range g.Assertions {
			assertion := assertion // 创建一个新的变量绑定
			go func() {
				resultChan <- assertion.Assert()
			}()
		}

		// 收集所有结果
		for i := 0; i < len(g.Assertions); i++ {
			assertResult := <-resultChan
			result.Results = append(result.Results, assertResult)
			if !assertResult.Passed {
				result.Passed = false
			}
		}
	} else {
		// 串行执行
		result.Passed = true
		for _, assertion := range g.Assertions {
			assertResult := assertion.Assert()
			result.Results = append(result.Results, assertResult)

			// 如果断言失败且配置了失败即停止
			if !assertResult.Passed {
				result.Passed = false
				if g.Options.StopOnFirstFailure {
					break
				}
			}
		}
	}

	return result
}

// Assert executes the assertion
func (a *Assertion) Assert() *AssertionResult {
	result := &AssertionResult{
		Name:        a.Name,
		ActualValue: a.ActualValue,
	}

	// 获取预期值（支持依赖注入）
	expectedValue := a.ExpectedValue
	if a.Dependency != nil {
		expectedValue = a.Dependency.Value
	}
	result.ExpectedValue = expectedValue

	// 根据断言类型执行相应的检查
	switch a.Type {
	case AssertEqual:
		if a.Options.DeepComparison {
			result.Passed = reflect.DeepEqual(a.ActualValue, expectedValue)
		} else {
			result.Passed = a.ActualValue == expectedValue
		}

	case AssertNotEqual:
		if a.Options.DeepComparison {
			result.Passed = !reflect.DeepEqual(a.ActualValue, expectedValue)
		} else {
			result.Passed = a.ActualValue != expectedValue
		}

	case AssertContains:
		result.Passed = containsValue(a.ActualValue, expectedValue, a.Options.IgnoreCase)

	case AssertNotContains:
		result.Passed = !containsValue(a.ActualValue, expectedValue, a.Options.IgnoreCase)

	case AssertGreaterThan:
		result.Passed = compareValues(a.ActualValue, expectedValue, a.Options.Tolerance) > 0

	case AssertLessThan:
		result.Passed = compareValues(a.ActualValue, expectedValue, a.Options.Tolerance) < 0

	case AssertGreaterOrEqual:
		result.Passed = compareValues(a.ActualValue, expectedValue, a.Options.Tolerance) >= 0

	case AssertLessOrEqual:
		result.Passed = compareValues(a.ActualValue, expectedValue, a.Options.Tolerance) <= 0

	case AssertRegexMatch:
		if pattern, ok := expectedValue.(string); ok {
			if str, ok := a.ActualValue.(string); ok {
				if a.Options.IgnoreCase {
					pattern = "(?i)" + pattern
					str = strings.ToLower(str)
				}
				matched, err := regexp.MatchString(pattern, str)
				if err != nil {
					result.Error = fmt.Sprintf("invalid regex pattern: %v", err)
					result.Passed = false
				} else {
					result.Passed = matched
				}
			} else {
				result.Error = "actual value is not a string"
				result.Passed = false
			}
		} else {
			result.Error = "expected value is not a string pattern"
			result.Passed = false
		}

	case AssertLengthEqual:
		if length, ok := expectedValue.(int); ok {
			actualLength := getLength(a.ActualValue)
			if actualLength >= 0 {
				result.Passed = actualLength == length
			} else {
				result.Error = "actual value has no length property"
				result.Passed = false
			}
		} else {
			result.Error = "expected value is not an integer"
			result.Passed = false
		}

	case AssertTypeMatch:
		expectedType := reflect.TypeOf(expectedValue)
		actualType := reflect.TypeOf(a.ActualValue)
		result.Passed = actualType == expectedType
		if !result.Passed {
			result.Error = fmt.Sprintf("type mismatch: expected %v, got %v", expectedType, actualType)
		}

	case AssertHasField:
		result.Passed = hasField(a.ActualValue, a.JsonPath)
		if !result.Passed {
			result.Error = fmt.Sprintf("field '%s' not found", a.JsonPath)
		}

	default:
		result.Error = fmt.Sprintf("unsupported assertion type: %s", a.Type)
		result.Passed = false
	}

	// 如果断言失败且有错误模板，使用模板格式化错误信息
	if !result.Passed && a.ErrorTemplate != "" {
		result.Error = formatErrorMessage(a.ErrorTemplate, map[string]interface{}{
			"name":           a.Name,
			"type":           a.Type,
			"actual_value":   a.ActualValue,
			"expected_value": expectedValue,
			"error":          result.Error,
		})
	}

	return result
}

// containsValue checks if expectedValue is contained in actualValue
func containsValue(actual, expected interface{}, ignoreCase bool) bool {
	actualValue := reflect.ValueOf(actual)

	switch actualValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < actualValue.Len(); i++ {
			if reflect.DeepEqual(actualValue.Index(i).Interface(), expected) {
				return true
			}
		}

	case reflect.String:
		if expectedStr, ok := expected.(string); ok {
			actualStr := actual.(string)
			if ignoreCase {
				return strings.Contains(strings.ToLower(actualStr), strings.ToLower(expectedStr))
			}
			return strings.Contains(actualStr, expectedStr)
		}

	case reflect.Map:
		for _, key := range actualValue.MapKeys() {
			if reflect.DeepEqual(actualValue.MapIndex(key).Interface(), expected) {
				return true
			}
		}
	}

	return false
}

// compareValues compares two values and returns:
// -1 if actual < expected
// 0 if actual == expected
// 1 if actual > expected
func compareValues(actual, expected interface{}, tolerance float64) int {
	switch actual.(type) {
	case int:
		if expectedInt, ok := expected.(int); ok {
			actualInt := actual.(int)
			if float64(actualInt-expectedInt) <= -tolerance {
				return -1
			} else if float64(actualInt-expectedInt) >= tolerance {
				return 1
			}
			return 0
		}

	case float64:
		if expectedFloat, ok := expected.(float64); ok {
			actualFloat := actual.(float64)
			if actualFloat-expectedFloat <= -tolerance {
				return -1
			} else if actualFloat-expectedFloat >= tolerance {
				return 1
			}
			return 0
		}

	case string:
		if expectedStr, ok := expected.(string); ok {
			return strings.Compare(actual.(string), expectedStr)
		}
	}

	return 0
}

// hasField checks if the specified field exists in the actual value
func hasField(actual interface{}, fieldPath string) bool {
	actualValue := reflect.ValueOf(actual)

	// 处理指针
	if actualValue.Kind() == reflect.Ptr {
		actualValue = actualValue.Elem()
	}

	// 处理结构体
	if actualValue.Kind() == reflect.Struct {
		_, found := actualValue.Type().FieldByName(fieldPath)
		return found
	}

	// 处理 map
	if actualValue.Kind() == reflect.Map {
		return actualValue.MapIndex(reflect.ValueOf(fieldPath)).IsValid()
	}

	return false
}

// getLength returns the length of a value if it supports length operations
func getLength(value interface{}) int {
	valueOf := reflect.ValueOf(value)

	switch valueOf.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return valueOf.Len()
	default:
		return -1
	}
}

// formatErrorMessage formats the error message using the provided template and data
func formatErrorMessage(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
	}
	return result
}
