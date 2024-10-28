package apirunner

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func assert(actual interface{}, expected interface{}, dataType string, operation string) (bool, error) {
	switch dataType {
	case "integer":
		actualInt, err := toInt(actual)
		if err != nil {
			return false, fmt.Errorf("实际值转换为整数失败: %v", err)
		}
		expectedInt, err := toInt(expected)
		if err != nil {
			return false, fmt.Errorf("预期值转换为整数失败: %v", err)
		}
		return compareInts(actualInt, expectedInt, operation)
	case "string":
		actualStr, ok := actual.(string)
		if !ok {
			return false, fmt.Errorf("实际值不是字符串类型")
		}
		expectedStr, ok := expected.(string)
		if !ok {
			return false, fmt.Errorf("预期值不是字符串类型")
		}
		return compareStrings(actualStr, expectedStr, operation)
	case "bool":
		actualBool, ok := actual.(bool)
		if !ok {
			return false, fmt.Errorf("实际值不是布尔类型")
		}
		expectedBool, ok := expected.(bool)
		if !ok {
			return false, fmt.Errorf("预期值不是布尔类型")
		}
		return compareBools(actualBool, expectedBool, operation)
	case "array_len":
		actualLen, err := getArrayLength(actual)
		if err != nil {
			return false, fmt.Errorf("获取实际数组长度失败: %v", err)
		}
		expectedLen, err := toInt(expected)
		if err != nil {
			return false, fmt.Errorf("预期值转换为整数失败: %v", err)
		}
		return compareInts(actualLen, expectedLen, operation)
	default:
		return false, fmt.Errorf("不支持的数据类型: %s", dataType)
	}
}

func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("无法转换为整数: %v", value)
	}
}

func getArrayLength(value interface{}) (int, error) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return 0, fmt.Errorf("值不是数组或切片类型")
	}
	return rv.Len(), nil
}

func compareInts(actual, expected int, operation string) (bool, error) {
	switch operation {
	case "equal":
		return actual == expected, nil
	case "noeql":
		return actual != expected, nil
	case "gt":
		return actual > expected, nil
	case "lt":
		return actual < expected, nil
	case "gte":
		return actual >= expected, nil
	case "lte":
		return actual <= expected, nil
	default:
		return false, fmt.Errorf("不支持的操作类型: %s", operation)
	}
}

func compareStrings(actual, expected string, operation string) (bool, error) {
	switch operation {
	case "equal":
		return actual == expected, nil
	case "noeql":
		return actual != expected, nil
	case "contains":
		return strings.Contains(actual, expected), nil
	case "startswith":
		return strings.HasPrefix(actual, expected), nil
	case "endswith":
		return strings.HasSuffix(actual, expected), nil
	default:
		return false, fmt.Errorf("不支持的操作类型: %s", operation)
	}
}

func compareBools(actual, expected bool, operation string) (bool, error) {
	switch operation {
	case "equal":
		return actual == expected, nil
	case "noeql":
		return actual != expected, nil
	default:
		return false, fmt.Errorf("布尔类型不支持的操作类型: %s", operation)
	}
}

func assertArrayLen(value []interface{}, desire int, operation string) bool {
	switch operation {
	case "eq":
		return len(value) == desire
	case "lt":
		return len(value) < desire
	case "gt":
		return len(value) > desire
	case "lte":
		return len(value) <= desire
	case "gte":
		return len(value) >= desire
	}
	return false
}

func assertBool(value bool, desire bool, operation string) bool {
	switch operation {
	case "eq":
		return value == desire
	}
	return false
}

func assertInt(value int, desire int, operation string) bool {
	switch operation {
	case "equal":
		return value == desire
	case "gt":
		return value > desire
	case "lt":
		return value < desire
	case "gte":
		return value >= desire
	case "lte":
		return value <= desire
	}
	return false
}

func assertString(value string, desire string, operation string) bool {
	switch operation {
	case "equal":
		return value == desire
	default:
		{
			return value == desire
		}
	}
}
