package logic

import "errors"

var (
	HTTP_STATUS_NOT_200          = errors.New("状态码不为200")
	APIFOX_DOC_AUTH_FAILED_ERROR = errors.New("apifox cookies获取失败")
)
