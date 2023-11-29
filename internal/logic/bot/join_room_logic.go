package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/GuardedTalk/bot/internal/engine"
	"github.com/GuardedTalk/bot/internal/svc"
	"github.com/GuardedTalk/bot/internal/types"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"github.com/zeromicro/go-zero/core/logx"
)

type JoinRoomLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJoinRoomLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinRoomLogic {
	return &JoinRoomLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JoinRoomLogic) JoinRoom(req *types.RoomReq) (resp *types.Response, err error) {
	host := "ws://localhost:7880"
	apiKey := "APIE2J6mepSno7o"
	apiSecret := "rbWwpd5OpND3qz5UtufNZviMdPs6CbhBxGZXehBLc7b"
	roomName := req.RoomID
	identity := "botuser"
	roomCB := &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: trackSubscribed,
		},
	}
	room, err := lksdk.ConnectToRoom(host, lksdk.ConnectInfo{
		APIKey:              apiKey,
		APISecret:           apiSecret,
		RoomName:            roomName,
		ParticipantIdentity: identity,
	}, roomCB)
	if err != nil {
		panic(err)
	}
	l.svcCtx.Bot[roomName] = room
	return
}

func trackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	if strings.EqualFold(track.Codec().MimeType, webrtc.MimeTypeOpus) {
		oggFile, err := oggwriter.New("output.ogg", 48000, 2)
		if err != nil {
			panic(err)
		}
		saveToDisk(oggFile, track)

	}
}
func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	// e, _ := engine.New()
	ae, err := engine.NewAudioEngine()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			fmt.Println(err)
			return
		}
		if _, err := ae.DecodePacket(rtpPacket); err != nil {
			fmt.Println(err)
		}
		if err := i.WriteRTP(rtpPacket); err != nil {
			fmt.Println(err)
			return
		}
	}
}
