package airstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ac "github.com/vocdoni/vote-frame/airstack/client"
	"go.vocdoni.io/dvote/log"
)

const validateFrameEndpoint = "https://hubs.airstack.xyz/v1/validateMessage"

// Airstack wraps all the required artifacts for interacting with the Airstack API
type Airstack struct {
	*ac.Client
	maxHolders         uint32          // maxHolders is the maximum number of holders to be retrieved from the Airstack API
	supportAPIEndpoint string          // supportAPI is the URL of the support API
	tokenWhitelist     map[string]bool // whitelist of tokens to be considered for avoiding some checks
}

// NewAirstack creates a new Airstack artifact with a reference to a MongoDB and an Airstack client that
// enables to make predefined queries to the Airstack GraphQL API.
func NewAirstack(
	ctx context.Context,
	endpoint,
	apiKey,
	supportAPI string,
	supportedBlockchains,
	tokenWhitelist []string,
	maxHolders uint32,
) (*Airstack, error) {
	client, err := ac.NewClient(ctx, endpoint, apiKey, supportedBlockchains)
	if err != nil {
		return nil, fmt.Errorf("error creating Airstack: %w", err)
	}
	whitelist := make(map[string]bool)
	for _, token := range tokenWhitelist {
		if len(token) < common.AddressLength {
			continue
		}
		if !common.IsHexAddress(token) {
			log.Warnf("invalid token address, skipping: %s", token)
			continue
		}
		whitelist[token] = true
	}
	return &Airstack{
		Client:             client,
		maxHolders:         maxHolders,
		supportAPIEndpoint: supportAPI,
		tokenWhitelist:     whitelist,
	}, nil
}

func (a *Airstack) ApiKey() string {
	return a.Client.ApiKey()
}

// MaxHolders returns the maximum number of holders allowed to be retrieved from the Airstack API
func (a *Airstack) MaxHolders() uint32 {
	return a.maxHolders
}

func (a *Airstack) TokenDecimalsByToken(tokenAddress, blockchain string) (int, error) {
	td, err := a.TokenDetails(common.HexToAddress(tokenAddress), blockchain)
	if err != nil {
		return 0, fmt.Errorf("error getting token details: %w", err)
	}
	return td.Decimals, nil
}

// TokenWhitelist returns the list of tokens to be considered for avoiding some checks
func (a *Airstack) TokenWhitelist() map[string]bool {
	return a.tokenWhitelist
}

func (a *Airstack) NumHoldersByTokenAnkrAPI(tokenAddress, blockchain string) (uint32, error) {
	if a.supportAPIEndpoint == "" {
		log.Warnf("No support API endpoint provided, skipping token holder count retrieval")
		return 0, nil
	}

	if blockchain == "ethereum" {
		blockchain = "eth" // Ankr API uses "eth" instead of "ethereum"
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "ankr_getTokenHoldersCount",
		"params": map[string]interface{}{
			"blockchain":      blockchain,
			"contractAddress": tokenAddress,
			"pageSize":        1, // we just want the latest holder counts
			"pageToken":       "",
		},
		"id": 1,
	}

	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("error marshalling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", a.supportAPIEndpoint, bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	// Attempt to send the request up to 3 times with a delay between attempts
	maxAttempts := 3
	attempt := 0
	var respBytes []byte
	for attempt < maxAttempts {
		resp, err := client.Do(req)
		if err != nil {
			log.Warnf("error sending request %v:", err)
			attempt++
			if attempt < maxAttempts {
				log.Debugf("Retrying... Attempt %d of %d\n", attempt+1, maxAttempts)
				time.Sleep(2 * time.Second)
				continue
			} else {
				return 0, fmt.Errorf("failed to send request after retries: %w", err)
			}
		}
		defer resp.Body.Close()
		respBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("error reading response body: %w", err)
		}
		break
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(respBytes, &responseMap); err != nil {
		return 0, fmt.Errorf("error unmarshalling response JSON: %w", err)
	}

	// Check for errors in the response
	errorMsg, found := responseMap["error"].(map[string]interface{})
	if found {
		// if error in data contains the string "no currency found" return 0 and nil
		log.Warnf("error in airstack response: %+v", errorMsg)
		return 0, nil
	}

	result, found := responseMap["result"].(map[string]interface{})
	if !found {
		return 0, fmt.Errorf("result field missing in response")
	}
	holderCountHistory, found := result["holderCountHistory"].([]interface{})
	if !found || len(holderCountHistory) == 0 {
		return 0, fmt.Errorf("holderCountHistory field missing or empty in response")
	}
	holderCount, found := holderCountHistory[0].(map[string]interface{})["holderCount"]
	if !found {
		return 0, fmt.Errorf("holderCount field missing in first item of holderCountHistory")
	}

	return uint32(holderCount.(float64)), nil
}

func ValidateFrameMessage(msg []byte, apikey string) {
	go func() {
		req, err := http.NewRequest(http.MethodPost, validateFrameEndpoint, bytes.NewBuffer(msg))
		if err != nil {
			log.Warn("error creating request:", err)
			return
		}
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("x-airstack-hubs", apikey)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Warn("error sending request:", err)
			return
		}
		airstackResponse := make(map[string]interface{})
		if err := json.NewDecoder(res.Body).Decode(&airstackResponse); err != nil {
			log.Warn("error decoding response:", err)
			return
		}
		if res.StatusCode != http.StatusOK {
			log.Warnw("airstack API returned unexpected status",
				"code", res.StatusCode,
				"url", req.URL.String(),
				"authorization", req.Header.Get("x-airstack-hubs"),
				"apikey", apikey,
				"response", fmt.Sprintf("%+v", airstackResponse),
				"request", printHTTPRequest(req),
			)
			return
		}
		isValid, ok := airstackResponse["isValid"]
		if !ok {
			log.Warn("isValid field missing in response")
			return
		}
		if valid, ok := isValid.(bool); !ok || !valid {
			log.Warn("invalid frame message")
			return
		}
	}()
}

func printHTTPRequest(req *http.Request) string {
	// Dump the request as a string
	requestDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err.Error()
	}
	return (string(requestDump))
}
