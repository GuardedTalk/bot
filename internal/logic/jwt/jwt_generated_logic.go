package jwt

import (
	"context"
	"time"

	"github.com/GuardedTalk/bot/internal/svc"
	"github.com/GuardedTalk/bot/internal/types"
	"github.com/livekit/protocol/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

type JwtGeneratedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJwtGeneratedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JwtGeneratedLogic {
	return &JwtGeneratedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JwtGeneratedLogic) JwtGenerated(req *types.JwtGenerateReq) (resp *types.JwtInfo, err error) {
	key, err := getJoinToken("APIE2J6mepSno7o", "rbWwpd5OpND3qz5UtufNZviMdPs6CbhBxGZXehBLc7b", req.RoomName, req.Identity, req.Name)
	if err != nil {
		return nil, err
	}
	return &types.JwtInfo{
		AccessToken: key,
	}, nil
}

func getJoinToken(apiKey, apiSecret, room, identity, name string) (string, error) {
	canPublish := true
	canSubscribe := true

	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin:     true,
		Room:         room,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetName(name).
		SetValidFor(time.Hour)

	return at.ToJWT()
}
