package logic

// 同步 api doc 相关的接口
const (
	APIFOX_DOC_AUTH_URL = "https://apifox.com/api/v1/shared-doc-auth?locale=zh-CN"
	APIFOX_TREE_URL     = "https://apifox.com/api/v1/shared-docs/{docid}/http-api-tree?locale=zh-CN"
	APIFOX_DETAIL_URL   = "https://apifox.com/api/v1/shared-docs/{docid}/http-apis/{apiid}?locale=zh-CN"
)

const (
	HTTP_OK_STATUS = 200
)

const (
	HEADER_APIFOX_VERSION   = "2.2.30"
	HEADER_USERAGENT_APIFOX = "Apifox/1.0.0 (https://apifox.com)"
)

// 同步器类型
const (
	APIFOXDOC = "APIFOXDOC"
)

// api同步来源
const (
	SYNC_APIFOX = "APIFOX"
)
