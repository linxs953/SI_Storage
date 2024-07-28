package scene

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/scene"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func DeleteSceneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteSceneDto
		req.Scid = r.URL.Query().Get("sceneId")
		l := scene.NewDeleteSceneLogic(r.Context(), svcCtx)
		resp, err := l.DeleteScene(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
