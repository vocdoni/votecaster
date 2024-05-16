package signer

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	qt "github.com/frankban/quicktest"
)

func TestJavascriptCompatibility(t *testing.T) {
	q := qt.New(t)

	// hardcoded values from the test case
	signerPubKey, err := hex.DecodeString("7abe15cd0740f0cf02670bb43ac6b508fe506c8b2ec92a41479d0f22ecf2c02a")
	q.Assert(err, qt.IsNil)
	deadline := uint64(1715693405)
	FID := uint64(529726)

	// ethereum account privateKey
	sk, err := crypto.HexToECDSA("46f02985c70cd39ec3e5856ecd41470957cda875165e4148c30cbfd80e95fdd0")
	q.Assert(err, qt.IsNil)

	// perform signature
	signature, err := signKeyRequest(sk, FID, signerPubKey, deadline)
	q.Assert(err, qt.IsNil)

	// expected signature
	expectedSignature := "907aefb0b851d5fbd6d36eb755dec6f13eb24a6b99025e481a88a4e4ade8724c33ad7add801d8d2307a98e0a2049c42a26e4487abea72bb57f7ee03074b264821c"
	q.Assert(hex.EncodeToString(signature), qt.Equals, expectedSignature)
}
