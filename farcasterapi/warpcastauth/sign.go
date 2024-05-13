package main

// reference https://docs.farcaster.xyz/reference/contracts/reference/signed-key-request-validator#signedkeyrequest-signature

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// reference https://github.com/farcasterxyz/hub-monorepo/blob/main/packages/core/src/eth/contracts/signedKeyRequestValidator.ts#L19
// https://github.com/farcasterxyz/contracts/blob/1aceebe916de446f69b98ba1745a42f071785730/src/validators/SignedKeyRequestValidator.sol
// signature over keccak256("SignedKeyRequest(uint256 requestFid,bytes key,uint256 deadline)");

var SIGNED_KEY_REQUEST_VALIDATOR_EIP_712_TYPES = map[string][]apitypes.Type{
	"EIP712Domain": {
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	},
	"SignedKeyRequest": {
		{Name: "requestFid", Type: "uint256"},
		{Name: "key", Type: "bytes"},
		{Name: "deadline", Type: "uint256"},
	},
}

var SIGNED_KEY_REQUEST_VALIDATOR_EIP_712_DOMAIN = apitypes.TypedDataDomain{
	Name:              "Farcaster SignedKeyRequestValidator",
	Version:           "1",
	ChainId:           math.NewHexOrDecimal256(10),
	VerifyingContract: "0x00000000FC700472606ED4fA22623Acf62c60553",
}

// signKeyRequest signs the request using EIP-712 structured data signing
func signKeyRequest(privateKey *ecdsa.PrivateKey, requestFid uint64, publicKey []byte, deadline uint64) ([]byte, error) {
	fid := new(big.Int).SetUint64(requestFid)

	data := apitypes.TypedData{
		Types:       SIGNED_KEY_REQUEST_VALIDATOR_EIP_712_TYPES,
		PrimaryType: "SignedKeyRequest",
		Domain:      SIGNED_KEY_REQUEST_VALIDATOR_EIP_712_DOMAIN,
		Message: apitypes.TypedDataMessage{
			"requestFid": fid,
			"key":        "0x" + hex.EncodeToString(publicKey),
			"deadline":   new(big.Int).SetUint64(deadline),
		},
	}

	dataHash, _, err := apitypes.TypedDataAndHash(data)
	if err != nil {
		return nil, fmt.Errorf("error hashing typed data: %w", err)
	}

	signature, err := crypto.Sign(dataHash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error signing typed data: %w", err)
	}

	// update the recovery Id
	// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	signature[64] += 27

	return signature, nil
}

func signEIP191(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	hash := crypto.Keccak256([]byte(msg))
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}
