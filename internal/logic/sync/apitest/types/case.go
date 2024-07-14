package types

/*
	api用例yaml格式
*/
type ApiSuit struct {
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	SceneId     string `yaml:"sceneId"`
	Steps       []Step `yaml:"steps"`
}

type Step struct {
	StepId     string           `yaml:"stepId"`
	ApiId      int              `yaml:"apiId"`
	ApiName    string           `yaml:"apiName"`
	Dependency []stepDependency `yaml:"dependency"`
	Global     []stepGlobal     `yaml:"global"`
	Output     []stepOutput     `yaml:"output"`
	Expect     stepExpect       `yaml:"expect"`
}

type stepDependency struct {
	// Type       string                `yaml:"type"`
	DataId     string                `yaml:"dataId"`
	UpstreamId string                `yaml:"upstreamId"`
	Refer      []stepDependencyRefer `yaml:"refer"`
}
type stepDependencyRefer struct {
	Type  string `yaml:"type"`
	Field string `yaml:"field"`
	UseAt string `yaml:"useAt"`
}
type stepGlobal struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type stepOutput struct {
	// job执行导出的路径, sceneId.stepId
	Key            string `yaml:"key"`
	OutputName     string `yaml:"outputName"`
	OutputValue    any    `yaml:"outputValue"`
	OutputDataType string `yaml:"outputDataType"`
}

type stepExpect struct {
	HttpCode int `yaml:"httpCode"`
	// BizCode  int               `yaml:"code"`
	Duration int               `yaml:"duration"`
	Fields   []stepExpectFiled `yaml:"fields"`
}
type stepExpectFiled struct {
	FieldName string `yaml:"fielName"`
	Operation string `yaml:"operation"`
	DataType  string `yaml:"type"`
	// 只能做基础数据类型断言
	Desire any `yaml:"desire"`
}
