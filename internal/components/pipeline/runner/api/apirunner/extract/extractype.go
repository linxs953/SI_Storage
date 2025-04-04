package extract

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Extractor struct {
	Name     string                 `json:"name"`
	Data     map[string]interface{} `json:"data"`
	JsonPath string                 `json:"json_path"`
	Target   TargetValue            `json:"target"`
}

type TargetValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (e *Extractor) Extract() (TargetValue, error) {
	// 解析e.jsonpath, 格式类似 $.data.name, 需要兼容$.data.0.name的场景
	jsonPath := strings.Split(e.JsonPath, ".")
	var result interface{}
	result = e.Data
	for _, p := range jsonPath[1:] {
		switch v := result.(type) {
		case map[string]interface{}:
			// 如果当前是 map，不应该出现数字索引
			if _, err := strconv.Atoi(p); err == nil {
				return TargetValue{}, fmt.Errorf("在 map 中不能使用数字索引: %s", e.JsonPath)
			}
			result = v[p]
		case []interface{}:
			i, err := strconv.Atoi(p)
			if err != nil {
				return TargetValue{}, fmt.Errorf("数组索引必须是数字: %s", e.JsonPath)
			}
			if i < 0 || i >= len(v) {
				return TargetValue{}, fmt.Errorf("数组索引越界: %s", e.JsonPath)
			}
			result = v[i]
		default:
			return TargetValue{}, fmt.Errorf("invalid json path: %s", e.JsonPath)
		}
	}
	if result == nil {
		return TargetValue{}, fmt.Errorf("invalid json path: %s", e.JsonPath)
	}
	return TargetValue{
		Type:  fmt.Sprintf("%T", result),
		Value: result,
	}, nil

}

func getTargetValue[T any](value interface{}) (T, error) {
	switch v := value.(type) {
	case string:
		var s T
		return s, json.Unmarshal([]byte(v), &s)
	case int, int8, int16, int32, int64:
		var i T
		return i, json.Unmarshal([]byte(strconv.FormatInt(v.(int64), 10)), &i)
	case float32, float64:
		var f T
		return f, json.Unmarshal([]byte(strconv.FormatFloat(v.(float64), 'f', -1, 64)), &f)
	case bool:
		var b T
		return b, json.Unmarshal([]byte(strconv.FormatBool(v)), &b)
	default:
		return *new(T), fmt.Errorf("unsupported type %T", value)
	}

}
