package task

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

type CreateTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTaskLogic {
	return &CreateTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTaskLogic) CreateTask(req *types.CreateTaskDto) (resp *types.CreateTaskResp, err error) {
	resp = &types.CreateTaskResp{
		Code:    0,
		Message: "创建任务成功",
	}
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	smod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	var taskRecord taskinfo.TaskInfo
	for _, sid := range req.SceneList {
		var scene *sceneinfo.SceneInfo
		if scene, err = smod.FindOneBySceneID(context.Background(), sid.SceneId); err != nil {
			resp.Code = 1
			resp.Message = err.Error()
			return
		}
		for i := 0; i < sid.Count; i++ {
			scene.SceneName = fmt.Sprintf("%s-%d", scene.SceneName, i+1)
			taskRecord.Scenes = append(taskRecord.Scenes, scene.Scene)
		}
	}
	taskRecord.TaskName = req.TaskName
	taskRecord.Author = req.Author
	taskMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskRecord.TaskID = uuid.New().String()
	if err = taskMod.Insert(context.Background(), &taskRecord); err != nil {
		resp.Code = 2
		resp.Message = "创建任务记录失败"
		return
	}
	return
}
