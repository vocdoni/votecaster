package main

import "go.vocdoni.io/dvote/types"

// FrameSignaturePacket mirrors the JSON structure received by the Frame server.
type FrameSignaturePacket struct {
	UntrustedData struct {
		FID           int64          `json:"fid"`
		URL           string         `json:"url"`
		MessageHash   string         `json:"messageHash"`
		Timestamp     int64          `json:"timestamp"`
		Network       int            `json:"network"`
		ButtonIndex   int            `json:"buttonIndex"`
		State         []byte         `json:"state"`
		InputText     string         `json:"inputText"`
		TransactionID types.HexBytes `json:"transactionId"`
		Address       types.HexBytes `json:"address"`
		CastID        struct {
			FID  int64  `json:"fid"`
			Hash string `json:"hash"`
		} `json:"castId"`
	} `json:"untrustedData"`
	TrustedData struct {
		MessageBytes string `json:"messageBytes"`
	} `json:"trustedData"`
}
