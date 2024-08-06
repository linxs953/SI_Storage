package scene

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/scene"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func ModifySceneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SceneUpdate
		var resp types.UpdateSceneVO = types.UpdateSceneVO{
			Code:    0,
			Message: "更新成功",
		}
		w.Header().Set("Content-Type", "application/json")

		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		sceneID := r.URL.Query().Get("scid")
		if sceneID == "" {
			return
		}

		l := scene.NewModifySceneLogic(r.Context(), svcCtx)
		scene, err := l.ModifyScene(&req, sceneID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			actions := []map[string]interface{}{}
			actionsBytes, err := json.Marshal(scene.Actions)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			}
			if err = json.Unmarshal(actionsBytes, &actions); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			}

			resp.Data = types.GetSceneData{
				Actions:     actions,
				Author:      scene.Author,
				CreateAt:    scene.CreateAt.Format("2006-01-02 15:04:05"),
				Description: scene.Description,
				Env:         scene.EnvKey,
				ID:          scene.SceneId,
				SceneId:     scene.SceneId,
				SearchKey:   scene.SearchKey,
				UpdateAt:    scene.UpdateAt.Format("2006-01-02 15:04:05"),
			}
			json.NewEncoder(w).Encode(resp)
			httpx.Ok(w)
		}
	}
}
