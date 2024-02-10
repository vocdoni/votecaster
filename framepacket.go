package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"go.vocdoni.io/dvote/log"
	farcasterpb "go.vocdoni.io/dvote/vochain/transaction/proofs/farcasterproof/proto"
	"google.golang.org/protobuf/proto"
	"lukechampine.com/blake3"
)

const frameHashSize = 20

// FrameSignaturePacket mirrors the JSON structure received by the Frame server.
type FrameSignaturePacket struct {
	UntrustedData struct {
		FID         int64  `json:"fid"`
		URL         string `json:"url"`
		MessageHash string `json:"messageHash"`
		Timestamp   int64  `json:"timestamp"`
		Network     int    `json:"network"`
		ButtonIndex int    `json:"buttonIndex"`
		InputText   string `json:"inputText"`
		CastID      struct {
			FID  int64  `json:"fid"`
			Hash string `json:"hash"`
		} `json:"castId"`
	} `json:"untrustedData"`
	TrustedData struct {
		MessageBytes string `json:"messageBytes"`
	} `json:"trustedData"`
}

// VerifyFrameSignature validates the frame message and returns de deserialized frame action and public key.
func VerifyFrameSignature(packet *FrameSignaturePacket) (*farcasterpb.FrameActionBody, ed25519.PublicKey, error) {
	// Decode the message bytes from hex
	messageBytes, err := hex.DecodeString(packet.TrustedData.MessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode message bytes: %w", err)
	}

	msg := farcasterpb.Message{}
	if err := proto.Unmarshal(messageBytes, &msg); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal Message: %w", err)
	}
	log.Debugf("farcaster signed message: %s", log.FormatProto(&msg))

	if msg.Data == nil {
		return nil, nil, fmt.Errorf("invalid message data")
	}
	if msg.SignatureScheme != farcasterpb.SignatureScheme_SIGNATURE_SCHEME_ED25519 {
		return nil, nil, fmt.Errorf("invalid signature scheme")
	}
	if msg.Data.Type != farcasterpb.MessageType_MESSAGE_TYPE_FRAME_ACTION {
		return nil, nil, fmt.Errorf("invalid message type, got %s", msg.Data.Type.String())
	}
	var pubkey ed25519.PublicKey = msg.GetSigner()

	if pubkey == nil {
		return nil, nil, fmt.Errorf("signer is nil")
	}

	// Verify the hash and signature
	msgDataBytes, err := proto.Marshal(msg.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal message data: %w", err)
	}

	log.Debugw("verifying message signature", "size", len(msgDataBytes))
	h := blake3.New(160, nil)
	if _, err := h.Write(msgDataBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to hash message: %w", err)
	}
	hashed := h.Sum(nil)[:frameHashSize]

	if !bytes.Equal(msg.Hash, hashed) {
		return nil, nil, fmt.Errorf("hash mismatch (got %x, expected %x)", hashed, msg.Hash)
	}

	if !ed25519.Verify(pubkey, hashed, msg.GetSignature()) {
		return nil, nil, fmt.Errorf("signature verification failed")
	}
	actionBody := msg.Data.GetFrameActionBody()
	if actionBody == nil {
		return nil, nil, fmt.Errorf("invalid action body")
	}

	return actionBody, pubkey, nil
}
