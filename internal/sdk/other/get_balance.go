package sdk_other

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	transport_http "github.com/JaKoSTaWiND/polymarket_bot/internal/transport/http"
)

const (
	polygonRPCURL = "https://polygon-bor-rpc.publicnode.com"
	// Контракт pUSD в Polygon mainnet из https://docs.polymarket.com/resources/contracts.
	pUSDContract = "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
)

var polygonWalletAddress = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// GetBalance возвращает баланс кошелька walletAddress в микро-pUSD на последнем блоке Polygon.
// Контекст ctx позволяет отменить запрос. Если client равен nil, используется HTTP-клиент проекта.
// При некорректном адресе, сбое RPC или ошибке разбора функция возвращает 0 и ошибку.
// Сумма, превышающая uint64, также приводит к ошибке.
func GetBalance(ctx context.Context, client *http.Client, walletAddress string) (uint64, error) {
	if !polygonWalletAddress.MatchString(walletAddress) {
		return 0, fmt.Errorf("некорректный адрес кошелька Polygon: %q", walletAddress)
	}
	if client == nil {
		client = transport_http.HttpClient()
	}

	// В ABI аргумент address занимает 32 байта с 12 нулевыми байтами слева.
	calldata := "0x70a08231" + "000000000000000000000000" + walletAddress[2:]
	return readPUSDBalance(ctx, client, calldata)
}

// readPUSDBalance отправляет calldata для balanceOf через client с контекстом ctx.
// Возвращает исходное число микро-pUSD либо ошибку транспорта, RPC или разбора суммы.
func readPUSDBalance(ctx context.Context, client *http.Client, calldata string) (uint64, error) {
	const payload = `{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":%q,"data":%q},"latest"]}`
	body := fmt.Sprintf(payload, pUSDContract, calldata)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, polygonRPCURL, strings.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("запрос balanceOf: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Polygon RPC: HTTP %d", resp.StatusCode)
	}

	var reply struct {
		Version string `json:"jsonrpc"` // Версия протокола JSON-RPC.
		ID      int    `json:"id"`      // Идентификатор запроса eth_call.
		Result  string `json:"result"`  // Одно 32-байтовое ABI-слово с префиксом 0x.
		Error   *struct {
			Code    int    `json:"code"`    // Код ошибки узла.
			Message string `json:"message"` // Описание ошибки узла.
		} `json:"error"` // Заполняется при ошибке RPC.
	}
	if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
		return 0, fmt.Errorf("ответ Polygon RPC: %w", err)
	}
	if reply.Error != nil {
		return 0, fmt.Errorf("Polygon RPC (%d): %s", reply.Error.Code, reply.Error.Message)
	}
	if reply.Version != "2.0" || reply.ID != 1 {
		return 0, fmt.Errorf("некорректный ответ Polygon RPC на eth_call")
	}

	// Контракт возвращает микро-pUSD как uint256. Результат должен помещаться в uint64.
	word, ok := strings.CutPrefix(reply.Result, "0x")
	if !ok || len(word) != 64 {
		return 0, fmt.Errorf("balanceOf: ожидалось 32-байтовое значение uint256")
	}
	amount, err := strconv.ParseUint(word, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("некорректный баланс pUSD: %w", err)
	}
	return amount, nil
}
