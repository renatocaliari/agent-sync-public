package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type LiveKitService struct {
	host            string
	apiKey          string
	apiSecret       string
	roomClient      *lksdk.RoomServiceClient
	currentRoomName string
}

func NewLiveKitService(host, apiKey, apiSecret string) *LiveKitService {
	return &LiveKitService{
		host:       host,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		roomClient: lksdk.NewRoomServiceClient(host, apiKey, apiSecret),
	}
}

func (s *LiveKitService) CreateRoom(ctx context.Context, name string) (*livekit.Room, error) {
	room, err := s.roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            name,
		EmptyTimeout:    300,
		MaxParticipants: 2,
	})
	if err != nil {
		slog.Error("failed to create room", "error", err)
		return nil, err
	}
	return room, nil
}

func (s *LiveKitService) GenerateToken(roomName, identity string) (string, error) {
	at := auth.NewAccessToken(s.apiKey, s.apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	token, err := at.ToJWT()
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		return "", err
	}
	return token, nil
}

func (s *LiveKitService) DeleteRoom(ctx context.Context, roomName string) error {
	_, err := s.roomClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	return err
}

func (s *LiveKitService) Host() string {
	return s.host
}
