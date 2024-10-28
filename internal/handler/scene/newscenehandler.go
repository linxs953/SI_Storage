package scene

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"lexa-engine/internal/logic/scene"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
)

func NewSceneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateSceneDto
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := scene.NewNewSceneLogic(r.Context(), svcCtx)
		resp, err := l.NewScene(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
