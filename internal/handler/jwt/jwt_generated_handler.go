package jwt

import (
	"net/http"

	"github.com/GuardedTalk/bot/internal/logic/jwt"
	"github.com/GuardedTalk/bot/internal/svc"
	"github.com/GuardedTalk/bot/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func JwtGeneratedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.JwtGenerateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := jwt.NewJwtGeneratedLogic(r.Context(), svcCtx)
		resp, err := l.JwtGenerated(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
