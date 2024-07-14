package utils

import (
	"strconv"
	"strings"
	"time"
)

func GetTimeStamp(cond string) int64 {
	today := time.Now()
	if strings.Contains(cond, "天") {
		increment, _ := strconv.ParseInt(strings.Split(cond, "天")[0], 10, 64)
		newD := today.AddDate(0, 0, int(increment))
		return newD.Unix()
	}
	return 0
}
