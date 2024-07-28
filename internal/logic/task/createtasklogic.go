package task

import (
	"context"

	mong "lexa-engine/internal/model/mongo"
	"lexa-engine/internal/model/mongo/sceneinfo"
	"lexa-engine/internal/model/mongo/taskinfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
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
	murl := mong.GetMongoUrl(l.svcCtx.Config.Database.Mongo)
	smod := sceneinfo.NewSceneInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "SceneInfo")
	var taskRecord taskinfo.TaskInfo
	for _, sid := range req.SceneList {
		var scene *sceneinfo.SceneInfo
		if scene, err = smod.FindOneBySceneID(context.Background(), sid); err != nil {
			return
		}
		taskRecord.Scenes = append(taskRecord.Scenes, scene.Scene)
	}
	taskMod := taskinfo.NewTaskInfoModel(murl, l.svcCtx.Config.Database.Mongo.UseDb, "TaskInfo")
	taskRecord.TaskID = uuid.New().String()
	if err = taskMod.Insert(context.Background(), &taskRecord); err != nil {
		return
	}
	return
}
