package usecases_sdk

import (
	"context"
	"log/slog"

	"github.com/JaKoSTaWiND/polymarket_bot/internal/sdk"
	sdk_websocket "github.com/JaKoSTaWiND/polymarket_bot/internal/sdk/websocket"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
)

type SubscribeUserChannelUC struct {
	Logger    *slog.Logger
	ApisState *states.ApisState
	Client    *sdk.SDKClient
}

func (UC *SubscribeUserChannelUC) Execute(ctx context.Context, onEvent func(context.Context, []byte) error) error {
	Apis := UC.ApisState

	SubscriptionRequest := sdk_websocket.SubscriptionRequest{
		Auth: sdk_websocket.Auth{
			ApiKey:     Apis.PolyApiKey,
			Secret:     Apis.PolySecret,
			Passphrase: Apis.PolyPassphrase,
		},
		Type: "user",
	}

	conn, err := sdk_websocket.UserChannelSubscribe(ctx, UC.Client.WSDialer, SubscriptionRequest)
	if err != nil {
		return err
	}
	return sdk_websocket.ReadUserChannel(ctx, conn, onEvent)
}
