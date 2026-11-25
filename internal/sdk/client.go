package sdk

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type SDKClient struct {
	HttpClient *http.Client
	WSDialer   *websocket.Dialer
}

func CreateSDKClient(HttpClient *http.Client, WSDialer *websocket.Dialer) *SDKClient {
	return &SDKClient{
		HttpClient: HttpClient,
		WSDialer: WSDialer,
	}
}