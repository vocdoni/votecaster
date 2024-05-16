package signer

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.vocdoni.io/dvote/log"
)

const (
	ApiEndpoint      = "https://api.warpcast.com/v2/signed-key-requests"
	ApiEndpointCheck = "https://api.warpcast.com/v2/signed-key-request"
)

// GenerateSigner generates an Ed25519 key pair.
func GenerateSigner() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}

// SignData signs the data using the private key.
func SignData(privateKey ed25519.PrivateKey, data []byte) (signature []byte, err error) {
	signature = ed25519.Sign(privateKey, data)
	return signature, nil
}

// RegisterSigner sends a request to WarpCast API and returns the deep link URL to send to the user.
// It also returns a channel, which will be used to notify when signer registration is processed. The channel will return the FID of the user.
func RegisterSigner(privKey *ecdsa.PrivateKey, signer ed25519.PublicKey, fid uint64) (string, chan (uint64), error) {
	deadline := uint64(time.Now().Add(time.Minute * 5).Unix())
	signature, err := signKeyRequest(privKey, fid, signer, deadline)
	if err != nil {
		return "", nil, fmt.Errorf("error signing key request: %w", err)
	}

	body := map[string]interface{}{
		"key":        fmt.Sprintf("0x%x", signer),
		"signature":  fmt.Sprintf("0x%x", signature),
		"requestFid": fid,
		"deadline":   deadline,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", nil, err
	}

	resp, err := http.Post(ApiEndpoint, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var result struct {
		Result struct {
			SignedKeyRequest struct {
				Token       string `json:"token"`
				DeeplinkUrl string `json:"deeplinkUrl"`
			} `json:"signedKeyRequest"`
		} `json:"result"`
	}
	if err := json.Unmarshal(responseData, &result); err != nil {
		return "", nil, err
	}

	deeplink := strings.ReplaceAll(
		result.Result.SignedKeyRequest.DeeplinkUrl,
		"farcaster://", "https://client.warpcast.com/deeplinks/",
	)

	isStateDone := make(chan uint64)
	go func() {
		// Wait for the token to be processed
		time.Sleep(20 * time.Second)
		maxAttempts := 60
		for {
			time.Sleep(5 * time.Second)
			if maxAttempts == 0 {
				log.Warnw("max attempts reached for token status check", "token", result.Result.SignedKeyRequest.Token)
				break
			}
			maxAttempts--
			done, fid, err := checkTokenStatus(result.Result.SignedKeyRequest.Token)
			if err != nil {
				log.Warnw("error checking token status", "error", err)
				continue
			}
			if done {
				isStateDone <- fid
				break
			}
		}
	}()

	return deeplink, isStateDone, nil
}

func checkTokenStatus(token string) (bool, uint64, error) {
	client := &http.Client{}
	statusReq, err := http.NewRequest("GET", fmt.Sprintf("%s?token=%s", ApiEndpointCheck, token), nil)
	if err != nil {
		return false, 0, err
	}

	statusReq.Header.Set("Content-Type", "application/json")
	statusResp, err := client.Do(statusReq)
	if err != nil {
		return false, 0, err
	}

	bodyBytes, err := io.ReadAll(statusResp.Body)
	if err != nil {
		return false, 0, err
	}

	result := struct {
		Errors []map[string]string `json:"errors"`
		Result struct {
			SignedKeyRequest struct {
				State string `json:"state"`
				FID   uint64 `json:"userFid"`
			} `json:"signedKeyRequest"`
		} `json:"result"`
	}{}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return false, 0, err
	}
	if len(result.Errors) > 0 {
		return false, 0, fmt.Errorf("error checking token status: %v", result.Errors[0]["message"])
	}
	return result.Result.SignedKeyRequest.State == "completed", result.Result.SignedKeyRequest.FID, nil
}
