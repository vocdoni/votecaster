package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"go.vocdoni.io/dvote/log"
)

const (
	ApiEndpoint    = "https://api.warpcast.com/v2/signed-key-requests"
	DeadlineOffset = 60 * time.Minute
)

type SignedKeyRequest struct {
	Token       string `json:"token"`
	DeeplinkUrl string `json:"deeplinkUrl"`
	Key         string `json:"key"`
	RequestFid  int    `json:"requestFid"`
	State       string `json:"state"`
}

// GenerateKeyPair generates an Ed25519 key pair.
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}

// SignData signs the data using the private key.
func SignData(privateKey ed25519.PrivateKey, data []byte) (signature []byte, err error) {
	signature = ed25519.Sign(privateKey, data)
	return signature, nil
}

// CreateSignedKeyRequest sends a request to WarpCast API and returns the token and deep link URL.
func CreateSignedKeyRequest(privKey *ecdsa.PrivateKey, signer ed25519.PublicKey, fid uint64) (string, chan (bool), error) {
	deadline := uint64(time.Now().Add(time.Hour).Unix())
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
	fmt.Printf("Response: %s\n", responseData)

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

	isStateDone := make(chan bool)

	go func() {
		maxAttempts := 60
		for {
			time.Sleep(5 * time.Second)
			if maxAttempts == 0 {
				log.Warnw("max attempts reached for token status check", "token", result.Result.SignedKeyRequest.Token)
				break
			}

			done, err := checkTokenStatus(result.Result.SignedKeyRequest.Token)
			if err != nil {
				log.Warnw("error checking token status", "error", err)
				continue
			}
			if done {
				isStateDone <- true
				break
			}
			maxAttempts--
		}
	}()

	return deeplink, isStateDone, nil
}

// GenerateQRCode generates a QR code from the given URL.
func GenerateQRCode(url string) ([]byte, error) {
	fmt.Printf("Generating QR code for URL: %s\n", url)
	return qrcode.Encode(url, qrcode.Medium, 256)
}

func checkTokenStatus(token string) (bool, error) {
	client := &http.Client{}
	statusReq, err := http.NewRequest("GET", fmt.Sprintf("%s?token=%s", ApiEndpoint, token), nil)
	if err != nil {
		return false, err
	}

	statusReq.Header.Set("Content-Type", "application/json")
	statusResp, err := client.Do(statusReq)
	if err != nil {
		return false, err
	}

	bodyBytes, err := io.ReadAll(statusResp.Body)
	if err != nil {
		return false, err
	}
	fmt.Printf("Token status response: %s\n", bodyBytes)

	result := struct {
		Result struct {
			SignedKeyRequest struct {
				State string `json:"state"`
			} `json:"signedKeyRequest"`
		} `json:"result"`
	}{}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return false, err
	}

	fmt.Printf("Token status: %s\n", result.Result.SignedKeyRequest.State)
	return result.Result.SignedKeyRequest.State == "accept", nil
}
