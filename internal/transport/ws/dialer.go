package transport_ws

import "github.com/gorilla/websocket"

func WSDialer() *websocket.Dialer {
	Dialer := *websocket.DefaultDialer
	return &Dialer
}