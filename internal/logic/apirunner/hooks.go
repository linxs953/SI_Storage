package apirunner

import (
	"fmt"
	"lexa-engine/internal/logic/apirunner/actionHooks"
	"time"
)

const (
	Time_Hook = "time_hook"
)

type Hook struct {
	Type      string                `json:"type"`
	SQLEngine actionHooks.SqlEngine `json:"sql_engine"`
	MQEngine  actionHooks.MqEngine  `json:"mq_engine"`
	ReqEngine actionHooks.ReqEngine `json:"send_req_engine"`
}

func (h *Hook) Run() error {
	var err error
	switch h.Type {
	default:
		err = fmt.Errorf("unknown hook type: %s", h.Type)
	}
	return err
}

func NewHook(hookType string) *Hook {
	return &Hook{
		Type: hookType,
	}
}

func triggerSqlHook() {
	// TODO: trigger sql hook
}

func triggerHttpHook() {
	// TODO: trigger http hook
}

func triggerMqHook() {
	// TODO: trigger mq hook
}

func triggerRedisHook() {
	// TODO: trigger redis hook
}

func triggerTimeStampHook(extra int) []int64 {
	now := time.Now()
	future := now.AddDate(0, 0, extra)
	var timestamps = []int64{now.Unix(), future.Unix()}
	return timestamps
}
