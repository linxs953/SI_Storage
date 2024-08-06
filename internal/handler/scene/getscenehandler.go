package scene

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/scene"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func GetSceneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSceneDto
		resp := types.GetSceneVO{
			Code:    0,
			Message: "获取场景信息成功",
		}
		w.Header().Set("Content-Type", "application/json")

		req.Scid = r.URL.Query().Get("scid")
		if req.Scid == "" {
			resp.Code = 4
			resp.Message = "scid不能为空"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		l := scene.NewGetSceneLogic(r.Context(), svcCtx)
		scene, err := l.GetScene(&req)
		if err != nil {

			// 获取场景错误
			resp.Code = 2
			resp.Message = err.Error()
			json.NewEncoder(w).Encode(resp)
			return
		} else {

			// 获取场景成功，返回场景数据
			var actions []map[string]interface{}
			var actionsBytes []byte
			if actionsBytes, err = json.Marshal(scene.Actions); err != nil {

				// 序列化action失败
				resp.Code = 3
				resp.Message = err.Error()
				json.NewEncoder(w).Encode(resp)
				return
				// httpx.ErrorCtx(r.Context(), w, err)
			}
			if err = json.Unmarshal(actionsBytes, &actions); err != nil {

				// action反序列 map失败
				resp.Code = 3
				resp.Message = err.Error()
				json.NewEncoder(w).Encode(resp)
				return
			}
			resp := types.GetSceneVO{
				Code:    0,
				Message: "获取场景信息成功",
				Data: types.GetSceneData{
					ID:          scene.ID.Hex(),
					SceneName:   scene.SceneName,
					Author:      scene.Author,
					CreateAt:    scene.CreateAt.Format("2006-01-02 15:04:05"),
					Description: scene.Description,
					Env:         scene.EnvKey,
					SceneId:     scene.SceneId,
					SearchKey:   scene.SearchKey,
					UpdateAt:    scene.UpdateAt.Format("2006-01-02 15:04:05"),
					Actions:     actions,
					Timeout:      scene.Timeout,
					Retry:       scene.Retry,
				},
			}
			json.NewEncoder(w).Encode(resp)
			httpx.Ok(w)
		}
	}
}
