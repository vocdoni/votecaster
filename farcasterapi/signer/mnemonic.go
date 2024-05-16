package signer

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func MnemonicToPrivateKey(mnemonic string) (*ecdsa.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "") // Empty passphrase
	// Generate a master key from the seed
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	// Derive the path for Ethereum, m/44'/60'/0'/0/0 is a common path for the first Ethereum address
	childKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 44)
	if err != nil {
		return nil, err
	}
	childKey, err = childKey.NewChildKey(bip32.FirstHardenedChild + 60)
	if err != nil {
		return nil, err
	}
	childKey, err = childKey.NewChildKey(bip32.FirstHardenedChild)
	if err != nil {
		return nil, err
	}
	childKey, err = childKey.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	childKey, err = childKey.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	// Convert to an ECDSA private key
	privKey, err := crypto.ToECDSA(childKey.Key)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}
