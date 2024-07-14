// Code generated by goctl. DO NOT EDIT.
package types

type IndexVO struct {
	Message string `json:"message"`
}

type StartDto struct {
	JobType string `json:"jobType"`
	JobSpec string `json:"jobSpec"`
}

type StartResp struct {
	Code    int         `json:"code"`
	Message string      `json:"messsage"`
	Data    interface{} `json:"data"`
}

type TaskInfo struct {
	TaskId   string      `json:"taskId"`
	TaskName string      `json:"taskName"`
	TaskType string      `json:"taskType"`
	TaskSpec interface{} `json:"taskSpec"`
}

type RunResp struct {
	RunId    string      `json:"runId"`
	TaskId   string      `json:"taskId"`
	TaskSpec interface{} `json:"taskSpec"`
	State    int         `json:"state"`
	Message  string      `json:"message"`
}

type CreateTaskDto struct {
	TaskName string `json:"taskName"`
	TaskType string `json:"taskType"`
	TaskSpec string `json:"taskSpec"`
}

type CreateTaskResp struct {
	Code    int         `json:"code"`
	Message string      `json:"messsage"`
	Data    interface{} `json:"data"`
}

type UpdateTaskDto struct {
	TaskId   string `json:"taskId"`
	TaskName string `json:"taskName"`
	TaskType string `json:"taskType"`
	TaskSpec string `json:"taskSpec"`
}

type UpdateTaskResp struct {
	Code    int         `json:"code"`
	Message string      `json:"messsage"`
	Data    interface{} `json:"data"`
}

type DeleteTaskDto struct {
	TaskId string `json:"taskId"`
}

type DeleteTaskResp struct {
	Code    int         `json:"code"`
	Message string      `json:"messsage"`
	Data    interface{} `json:"data"`
}

type GetTaskDto struct {
	TaskId string `json:"taskId"`
}

type GetTaskResp struct {
	Code    int      `json:"code"`
	Message string   `json:"messsage"`
	Data    TaskInfo `json:"data"`
}

type GetTaskListDto struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
}

type GetTaskListResp struct {
	Code    int        `json:"code"`
	Message string     `json:"messsage"`
	Data    []TaskInfo `json:"data"`
}

type RunTaskDto struct {
	TaskId string `path:"taskId"`
}

type RunTaskResp struct {
	Code    int     `json:"code"`
	Message string  `json:"messsage"`
	Data    RunResp `json:"data"`
}

type ApiDetailDto struct {
	ApiId string `path:"apiId"`
}

type ApiListDto struct {
	PageNum  int `query:"pageNum"`
	PageSize int `query:"pageSize"`
}

type ApiDetailResp struct {
	Code    int                    `json:"code"`
	Message string                 `json:"messsage"`
	Data    map[string]interface{} `json:"data"`
}

type ApiListResp struct {
	Code        int                      `json:"code"`
	Message     string                   `json:"messsage"`
	Total       int                      `json:"total"`
	CurrentPage int                      `json:"currentPage"`
	TotalNum    int                      `json:"totalNum"`
	Data        []map[string]interface{} `json:"data"`
}

type TDInitResp struct {
	Code    int
	Message string
}
