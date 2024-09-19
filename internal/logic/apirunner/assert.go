package apirunner

import (
	"fmt"
	// "strconv"
	// "reflect"
		"github.com/zeromicro/go-zero/core/logx"
)

func assert(value interface{}, desire interface{}, datatype string, operation string) (bool, error) {
	switch datatype {
	case "integer":
		{
			// vv := reflect.ValueOf(value)
			// var intVal int
			// if vv.Kind() == reflect.Float64 {
			// 	intVal = int(vv.Float())
			// } else if vv.Kind() == reflect.Int {
			// 	intVal = int(vv.Int())
			// } else {
			// 	return false, fmt.Errorf("resp field %v is not int, getType: %s", value,reflect.TypeOf(value))
			// }
			intVal := fmt.Sprintf("%v",value)
			intDesire := fmt.Sprintf("%v",desire)
			logx.Error(intVal,intDesire)
			// if !ok {
			// 	return false, fmt.Errorf("desire %v is not int", value)
			// }
			return assertString(intVal, intDesire, operation), nil
		}
	case "string":
		{
			strV, ok := value.(string)
			if !ok {
				return false, fmt.Errorf("resp field %v is not string", value)
			}
			strDesire, ok := desire.(string)
			if !ok {
				return false, fmt.Errorf("desire %v is not string", value)
			}
			return assertString(strV, strDesire, operation), nil
		}
	case "bool":
		{
			boolV, ok := value.(bool)
			if !ok {
				return false, fmt.Errorf("resp field %v is not bool", value)
			}
			boolDesire, ok := desire.(bool)
			if !ok {
				return false, fmt.Errorf("desire %v is not bool", value)
			}
			return assertBool(boolV, boolDesire, operation), nil
		}
	case "array_len":
		{
			arrayV, ok := value.([]interface{})
			if !ok {
				return false, fmt.Errorf("resp field %v is not int", value)
			}
			desireInt, ok := desire.(int)
			if !ok {
				return false, fmt.Errorf("desire %v is not int", value)
			}
			return assertArrayLen(arrayV, desireInt, operation), nil
		}
	default:
		{
			return false, fmt.Errorf("datatype %v is not supported", datatype)
		}
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
