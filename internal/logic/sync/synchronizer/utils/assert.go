package utils

import (
	"encoding/json"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

func ReadArrayAny(data map[string]any, key string) (arr []any) {
	mapValue, ok := data[key]
	if !ok {
		return nil
	}
	arr, ok = mapValue.([]any)
	if !ok {
		logx.Error("map-value 不是Array类型")
		mapValueB, _ := json.Marshal(mapValue)
		logx.Error(string(mapValueB))
		return
	}
	return
}

func ReadMapValueInteger(data map[string]any, key string) int {
	mapValue, ok := data[key]
	if !ok {
		return -2
	}
	mapValueByte, _ := json.Marshal(mapValue)
	value, err := strconv.Atoi(string(mapValueByte))
	if err != nil {
		logx.Error(err)
		return -1

	}
	return value
}

func ReadMapValueString(data map[string]any, key string) string {
	mapValue, ok := data[key]
	if !ok {
		return ""
	}
	value, ok := mapValue.(string)
	if !ok {
		return ""
	}
	return value
}

func ReadMapValueObject(data map[string]any, key string) map[string]any {
	mapValue, ok := data[key]
	if !ok {
		return nil
	}
	value, ok := mapValue.(map[string]any)
	if !ok {
		return nil
	}
	return value
}

func AssertString(v any) string {
	var value string
	value, ok := v.(string)
	if !ok {
		return ""
	}
	return value
}

func AssertArrayAny(v any) []any {
	value, ok := v.([]any)
	if !ok {
		return nil
	}
	return value
}
