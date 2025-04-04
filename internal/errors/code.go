package errors

// 定义storage rpc的各种错误

// ErrorCode 错误代码类型
type ErrorCode int

const (
	// 成功 (0)
	Success ErrorCode = 100

	//---------------- 基础错误 (1000-1999) ----------------//
	InternalError      ErrorCode = 1001 // 服务器内部错误
	InvalidParameter   ErrorCode = 1002 // 无效参数
	ValidationFailed   ErrorCode = 1003 // 数据验证失败
	Unauthorized       ErrorCode = 1004 // 未授权
	Forbidden          ErrorCode = 1005 // 禁止访问
	NotFound           ErrorCode = 1006 // 资源不存在
	Conflict           ErrorCode = 1007 // 资源冲突
	RequestTimeout     ErrorCode = 1008 // 请求超时
	TooManyRequests    ErrorCode = 1009 // 请求过于频繁
	RateLimitExceeded  ErrorCode = 1010 // 超出速率限制
	ServiceUnavailable ErrorCode = 1011 // 服务不可用

	//---------------- 数据层错误 (2000-2999) ----------------//
	// 数据库错误 (2000-2099)
	DBConnectionFailed           ErrorCode = 2001 // 数据库连接失败
	DBQueryError                 ErrorCode = 2002 // 数据库查询错误
	DBNotFound                   ErrorCode = 2003 // 数据库记录不存在
	DBDuplicateEntry             ErrorCode = 2004 // 数据库记录重复
	DBTxBeginError               ErrorCode = 2005 // 事务开始失败
	DBTxCommitError              ErrorCode = 2006 // 事务提交失败
	DBTxRollbackError            ErrorCode = 2007 // 事务回滚失败
	InvlidMgoRecordError         ErrorCode = 2009 // mongodb record无效
	InvalidMgoRecordVersionError ErrorCode = 2010 // 更新版本不一致
	UpdateMgoRecordError         ErrorCode = 2011 // 更新记录失败
	InvalidMgoObjId              ErrorCode = 2012 // objId 无效
	DeleteMgoRecordError         ErrorCode = 2013 // 删除记录失败
	InvalidSceneIdError          ErrorCode = 2014 // 无效sceneId
	InvalidRelatedApiError       ErrorCode = 2015 // 无效relatedApi

	// 缓存错误 (2100-2199)
	CacheConnectionError ErrorCode = 2101 // 缓存连接失败
	CacheGetError        ErrorCode = 2102 // 缓存读取失败
	CacheSetError        ErrorCode = 2103 // 缓存写入失败
	CacheDeleteError     ErrorCode = 2104 // 缓存删除失败
	CacheMiss            ErrorCode = 2105 // 缓存未命中

	//---------------- 业务逻辑错误 (3000-3999) ----------------//
	InvalidOperation        ErrorCode = 3001 // 无效操作
	BusinessRuleViolated    ErrorCode = 3002 // 违反业务规则
	WorkflowError           ErrorCode = 3003 // 工作流错误
	StateTransitionError    ErrorCode = 3004 // 状态转换错误
	QuotaExceeded           ErrorCode = 3005 // 配额超限
	ExpiredResource         ErrorCode = 3006 // 资源已过期
	InvalidLicense          ErrorCode = 3007 // 无效许可证
	FeatureDisabled         ErrorCode = 3008 // 功能未启用
	GenerateTaskIDError     ErrorCode = 3009 // 生成任务id失败
	GetInterfaceDetailError ErrorCode = 3010 // 获取接口详情失败
	CreateSceneConfigError  ErrorCode = 3011 // 创建场景配置失败
	SceneNotFound           ErrorCode = 3012 // 场景不存在

	//---------------- RPC通信错误 (4000-4999) ----------------//
	RPCClientInitFailed     ErrorCode = 4001 // 客户端初始化失败
	RPCConnectionLost       ErrorCode = 4002 // 连接中断
	RPCRequestTimeout       ErrorCode = 4003 // 请求超时
	RPCProtocolError        ErrorCode = 4004 // 协议错误
	RPCSerializationError   ErrorCode = 4005 // 序列化失败
	RPCDeserializationError ErrorCode = 4006 // 反序列化失败
	RPCServiceNotFound      ErrorCode = 4007 // 服务不存在
	RPCMethodNotFound       ErrorCode = 4008 // 方法不存在

	//---------------- 外部服务错误 (5000-5999) ----------------//
	ThirdPartyAPIError   ErrorCode = 5001 // 第三方API错误
	PaymentGatewayError  ErrorCode = 5002 // 支付网关错误
	SMSGatewayError      ErrorCode = 5003 // 短信网关错误
	GeoServiceError      ErrorCode = 5004 // 地理服务错误
	AntiFraudSystemError ErrorCode = 5005 // 风控系统错误
)

var codeMessages = map[ErrorCode]string{
	Success:            "操作成功",
	InternalError:      "服务器内部错误",
	InvalidParameter:   "无效参数: %v",
	ValidationFailed:   "数据验证失败: %s",
	Unauthorized:       "未授权访问",
	Forbidden:          "禁止访问该资源",
	NotFound:           "请求资源不存在",
	Conflict:           "资源状态冲突",
	RequestTimeout:     "请求处理超时",
	TooManyRequests:    "请求过于频繁，请稍后重试",
	RateLimitExceeded:  "超出系统速率限制",
	ServiceUnavailable: "服务暂时不可用",

	// 数据库错误
	DBConnectionFailed:           "数据库连接失败",
	DBQueryError:                 "数据库查询错误",
	DBNotFound:                   "数据库记录不存在",
	DBDuplicateEntry:             "数据库记录重复",
	DBTxBeginError:               "事务启动失败",
	DBTxCommitError:              "事务提交失败",
	DBTxRollbackError:            "事务回滚失败",
	InvlidMgoRecordError:         "mongodb 记录无效",
	InvalidMgoRecordVersionError: "数据记录前后版本不一致",
	UpdateMgoRecordError:         "记录更新失败",
	InvalidMgoObjId:              "无效的ObjID",
	DeleteMgoRecordError:         "删除记录失败",

	// 缓存错误
	CacheConnectionError: "缓存服务器连接失败",
	CacheGetError:        "缓存读取失败",
	CacheSetError:        "缓存写入失败",
	CacheDeleteError:     "缓存删除失败",
	CacheMiss:            "缓存未命中",

	// 业务逻辑错误
	InvalidOperation:     "当前操作不允许执行",
	BusinessRuleViolated: "违反业务规则: %s",
	WorkflowError:        "工作流执行异常",
	StateTransitionError: "状态转换不符合规则",
	QuotaExceeded:        "操作超出配额限制",
	ExpiredResource:      "资源已过期",
	InvalidLicense:       "无效的授权许可",
	FeatureDisabled:      "该功能当前未启用",
	GenerateTaskIDError:  "生成任务ID失败",

	// RPC错误
	RPCClientInitFailed:     "RPC客户端初始化失败",
	RPCConnectionLost:       "RPC连接意外中断",
	RPCRequestTimeout:       "RPC请求超时",
	RPCProtocolError:        "RPC协议解析错误",
	RPCSerializationError:   "RPC请求序列化失败",
	RPCDeserializationError: "RPC响应反序列化失败",
	RPCServiceNotFound:      "目标服务不存在",
	RPCMethodNotFound:       "请求方法不存在",

	// 外部服务错误
	ThirdPartyAPIError:   "第三方服务调用失败",
	PaymentGatewayError:  "支付网关通信异常",
	SMSGatewayError:      "短信网关发送失败",
	GeoServiceError:      "地理位置服务异常",
	AntiFraudSystemError: "风控系统检测异常",
}
