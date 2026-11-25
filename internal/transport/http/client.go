package transport_http

import (
	"net/http"
	"time"
)

func HttpClient() *http.Client {
	HttpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	return &HttpClient
}