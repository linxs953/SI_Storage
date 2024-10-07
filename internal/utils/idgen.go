package utils

import (
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	// 时间戳偏移量，可以根据需要调整
	timestampLeftShift = 12 + 10
	// 序列号位数
	sequenceBits = 12
	// 机器ID位数
	machineIdBits = 10
	// 最大序列号
	maxSequence = -1 ^ (-1 << sequenceBits)
	// 序列掩码
	sequenceMask = maxSequence
	// 机器ID掩码
	machineIdMask = -1 ^ (-1 << machineIdBits)
	// 起始时间戳，可以根据需要调整
	startTime int64 = 1577836800000 // 2020-01-01 00:00:00 UTC
)

var (
	lastTimestamp int64 = -1
	sequence      int64 = 0
)

// 固定的机器ID，因为是单机部署
const machineId = 1

// 定义字符集
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateId() int64 {
	var timestamp int64

	// 获取当前时间戳，单位为毫秒
	timestamp = time.Now().UnixNano() / 1e6

	// 如果当前时间戳小于上一次时间戳，则发生时钟回拨
	if timestamp < lastTimestamp {
		panic(fmt.Sprintf("Clock moved backwards. Refusing to generate id for %d milliseconds", lastTimestamp-timestamp))
	}

	// 如果当前时间戳与上一次相同，则使用序列号
	if timestamp == lastTimestamp {
		sequence = (sequence + 1) & sequenceMask
		if sequence == 0 {
			timestamp = getNextMillisecond(lastTimestamp)
		}
	} else {
		sequence = 0
	}

	lastTimestamp = timestamp

	// 返回拼接后的ID
	return ((timestamp - startTime) << timestampLeftShift) | (machineId << sequenceBits) | sequence
}

func getNextMillisecond(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixNano() / 1e6
	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixNano() / 1e6
	}
	return timestamp
}

// 将雪花算法生成的ID转换为包含小写字母和数字的字符串
func EncodeToBase36(id int64) string {
	b := big.NewInt(id)
	var result strings.Builder
	base := big.NewInt(62)
	for b.Sign() > 0 {
		mod := new(big.Int).Mod(b, base)
		result.WriteByte(charset[mod.Int64()])
		b.Div(b, base)
	}
	// 反转结果字符串
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
