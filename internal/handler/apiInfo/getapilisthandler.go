package apiInfo

import (
	api "lexa-engine/internal/logic/apiInfo"
	"lexa-engine/internal/svc"
	"lexa-engine/internal/types"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetApiListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApiListDto
		pageSize, err := strconv.ParseInt(r.URL.Query().Get("pageSize"), 10, 64)
		if err != nil {
			logx.Error(err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		pageNum, err := strconv.ParseInt(r.URL.Query().Get("pageNum"), 10, 64)
		if err != nil {
			logx.Error(err)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		req.PageNum = int(pageNum)
		req.PageSize = int(pageSize)
		l := api.NewGetApiListLogic(r.Context(), svcCtx)
		resp, err := l.GetApiList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
