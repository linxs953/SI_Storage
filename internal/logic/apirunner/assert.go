package apirunner

import "fmt"

func assert(value interface{}, desire interface{}, datatype string, operation string) (bool, error) {
	switch datatype {
	case "int":
		{
			intV, ok := value.(int)
			if !ok {
				return false, fmt.Errorf("resp field %s is not int", value)
			}
			intDesire, ok := desire.(int)
			if !ok {
				return false, fmt.Errorf("desire %s is not int", value)
			}
			return assertInt(intV, intDesire, operation), nil
		}
	case "string":
		{
			strV, ok := value.(string)
			if !ok {
				return false, fmt.Errorf("resp field %s is not string", value)
			}
			strDesire, ok := desire.(string)
			if !ok {
				return false, fmt.Errorf("desire %s is not string", value)
			}
			return assertString(strV, strDesire, operation), nil
		}
	case "bool":
		{
			boolV, ok := value.(bool)
			if !ok {
				return false, fmt.Errorf("resp field %s is not bool", value)
			}
			boolDesire, ok := desire.(bool)
			if !ok {
				return false, fmt.Errorf("desire %s is not bool", value)
			}
			return assertBool(boolV, boolDesire, operation), nil
		}
	case "array_len":
		{
			arrayV, ok := value.([]interface{})
			if !ok {
				return false, fmt.Errorf("resp field %s is not int", value)
			}
			desireInt, ok := desire.(int)
			if !ok {
				return false, fmt.Errorf("desire %s is not int", value)
			}
			return assertArrayLen(arrayV, desireInt, operation), nil
		}
	default:
		{
			return false, fmt.Errorf("datatype %s is not supported", datatype)
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
	case "eq":
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
