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
)

const (
	ApiEndpoint    = "https://api.warpcast.com/v2/signed-key-requests"
	DeadlineOffset = 60 * time.Minute
)

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
func CreateSignedKeyRequest(privKey *ecdsa.PrivateKey, signer ed25519.PublicKey, fid uint64) (deeplinkUrl string, err error) {
	deadline := uint64(time.Now().Add(time.Hour).Unix())
	signature, err := signKeyRequest(privKey, fid, signer, deadline)
	if err != nil {
		return "", fmt.Errorf("error signing key request: %w", err)
	}

	body := map[string]interface{}{
		"key":        fmt.Sprintf("0x%x", signer),
		"signature":  fmt.Sprintf("0x%x", signature),
		"requestFid": fid,
		"deadline":   deadline,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	fmt.Printf("Sending request to API: %s\n%s\n", ApiEndpoint, jsonData)
	resp, err := http.Post(ApiEndpoint, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
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
		return "", err
	}

	return result.Result.SignedKeyRequest.DeeplinkUrl, nil
}

// GenerateQRCode generates a QR code from the given URL.
func GenerateQRCode(url string) ([]byte, error) {
	fmt.Printf("Generating QR code for URL: %s\n", url)
	return qrcode.Encode(url, qrcode.Medium, 256)
}
