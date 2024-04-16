package communities

import "testing"

func Test_decodeNetworkAddress(t *testing.T) {
	vocdoniParty := "base:0x225D58E18218E8d87f365301aB6eEe4CbfAF820b"
	address, blockchain, err := decodeNetworkAddress(vocdoniParty)
	if err != nil {
		t.Fatal(err)
	}
	if address.Hex() != "0x225D58E18218E8d87f365301aB6eEe4CbfAF820b" {
		t.Fatal("address is not correct")
	}
	if blockchain != "base" {
		t.Fatal("blockchain is not correct")
	}

	invalidParty := "base:0x225D58E18218E8d87f365301aB6eEe4CbfAF820b:"
	_, _, err = decodeNetworkAddress(invalidParty)
	if err == nil {
		t.Fatal("error should not be nil")
	}

	invalidAddr, noBlockchain, err := decodeNetworkAddress("base:")
	if err == nil {
		t.Fatal("error should not be nil")
	}
	if invalidAddr.Hex() != zeroAddress {
		t.Fatal("address is not correct")
	}
	if noBlockchain != "" {
		t.Fatal("blockchain is not correct")
	}

	zeroAddr, noBlockchain, err := decodeNetworkAddress("base:0x0")
	if err == nil {
		t.Fatal("error should not be nil")
	}
	if zeroAddr.Hex() != zeroAddress {
		t.Fatal("address is not correct")
	}
	if noBlockchain != "base" {
		t.Fatal("blockchain is not correct")
	}
}
