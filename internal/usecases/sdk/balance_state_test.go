package usecases_sdk_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JaKoSTaWiND/polymarket_bot/internal"
	di_usecases "github.com/JaKoSTaWiND/polymarket_bot/internal/di/usecases"
	"github.com/JaKoSTaWiND/polymarket_bot/internal/states"
)

// rpcReply подменяет ответ узла Polygon при проверке обновления состояния.
type rpcReply string

// RoundTrip возвращает тело ответа для тестового запроса.
func (reply rpcReply) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(reply)))}, nil
}

func TestGetBalanceUpdatesGeneralState(t *testing.T) {
	const previous int64 = 7_000_000
	settings := &internal.BootstrapSettings{PolyWalletAddress: "0x" + strings.Repeat("1", 40)}
	for _, tc := range []struct {
		name    string
		reply   string
		want    int64
		wantErr bool
	}{
		{"positive balance", fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"result":"0x%064x"}`, 15_000_000), 15_000_000, false},
		{"wallet emptied", fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"result":"0x%064x"}`, 0), 0, false},
		{"rpc error", `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"unavailable"}}`, previous, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &states.GeneralState{Balance: previous}
			client := &http.Client{Transport: rpcReply(tc.reply)}
			uc := di_usecases.CreateGetBalanceUC(nil, client, settings, state)
			balance, err := uc.Execute(t.Context())
			if (err != nil) != tc.wantErr || (!tc.wantErr && balance != tc.want) {
				t.Fatalf("Execute() = %d, %v", balance, err)
			}
			if state.Balance != tc.want {
				t.Errorf("GeneralState.Balance = %d, want %d", state.Balance, tc.want)
			}
		})
	}
}
