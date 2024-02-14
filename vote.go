package main

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"go.vocdoni.io/proto/build/go/models"

	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/util"
	"go.vocdoni.io/dvote/vochain/state"
	"go.vocdoni.io/dvote/vochain/transaction/proofs/farcasterproof"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNotElegible    = fmt.Errorf("not elegible")
	ErrNotInCensus    = fmt.Errorf("not in the census")
	ErrAlreadyVoted   = fmt.Errorf("already voted")
	ErrFrameSignature = fmt.Errorf("frame signature verification failed")
)

// vote creates a vote transaction, including the frame signature packet and sends it to the vochain.
// It returns the nullifier of the vote (which is the unique identifier of the vote), the voterID and an error.
func vote(packet *FrameSignaturePacket, electionID types.HexBytes, root []byte, cli *apiclient.HTTPclient) (types.HexBytes, types.HexBytes, uint64, error) {
	// fetch the public key from the signature and generate the census proof
	actionMessage, message, pubKey, err := VerifyFrameSignature(packet)
	if err != nil {
		return nil, nil, 0, ErrFrameSignature
	}

	// compute the voterID, based on the public key
	voterID := state.NewVoterID(state.VoterIDTypeEd25519, pubKey)
	log.Infow("received vote request", "electionID", electionID, "voterID", fmt.Sprintf("%x", voterID.Address()))

	// compute the nullifier for the vote (a hash of the voterID and the electionID)
	nullifier := farcasterproof.GenerateNullifier(message.Data.Fid, electionID)

	// check if the voter is elegible to vote (in the census)
	proof, err := cli.CensusGenProof(root, voterID.Address())
	if err != nil {
		return nil, voterID.Address(), message.Data.Fid, ErrNotInCensus
	}

	// check if the voter already voted
	_, code, err := cli.Request("GET", nil, "votes", "verify", electionID.String(), fmt.Sprintf("%x", nullifier))
	if err != nil {
		return nullifier, voterID.Address(), 0, fmt.Errorf("failed to verify vote: %w", err)
	}
	if code == http.StatusOK {
		return nullifier, voterID.Address(), message.Data.Fid, ErrAlreadyVoted
	}

	// build the vote package
	votePackage := &state.VotePackage{
		Votes: []int{packet.UntrustedData.ButtonIndex - 1},
	}
	votePackageBytes, err := votePackage.Encode()
	if err != nil {
		return nullifier, voterID.Address(), 0, fmt.Errorf("failed to encode vote package: %w", err)
	}

	// build the vote transaction
	vote := &models.VoteEnvelope{
		Nonce:       util.RandomBytes(16),
		ProcessId:   electionID,
		VotePackage: votePackageBytes,
	}

	log.Debugw("received",
		"msg", packet.TrustedData.MessageBytes,
		"fid", message.Data.Fid,
		"pubkey", fmt.Sprintf("%x", pubKey),
		"button", actionMessage.ButtonIndex,
		"url", actionMessage.Url)

	// build the proof for the vote transaction
	frameSignedMessage, err := hex.DecodeString(packet.TrustedData.MessageBytes)
	if err != nil {
		return nullifier, voterID.Address(), 0, fmt.Errorf("failed to decode frame signed message: %w", err)
	}

	arboProof := &models.ProofArbo{
		Type:            models.ProofArbo_BLAKE2B,
		Siblings:        proof.Proof,
		AvailableWeight: proof.LeafValue,
		KeyType:         proof.KeyType,
		VoteWeight:      proof.LeafValue,
	}

	vote.Proof = &models.Proof{
		Payload: &models.Proof_FarcasterFrame{
			FarcasterFrame: &models.ProofFarcasterFrame{
				SignedFrameMessageBody: frameSignedMessage,
				PublicKey:              pubKey,
				CensusProof:            arboProof,
			},
		},
	}

	// sign and send the vote transaction
	stx := models.SignedTx{}
	stx.Tx, err = proto.Marshal(&models.Tx{Payload: &models.Tx_Vote{Vote: vote}})
	if err != nil {
		return nullifier, voterID.Address(), 0, fmt.Errorf("failed to marshal vote transaction: %w", err)
	}

	txHash, nullifier, err := cli.SignAndSendTx(&stx)
	if err != nil {
		return nullifier, voterID.Address(), 0, fmt.Errorf("failed to sign and send vote transaction: %w", err)
	}

	log.Infow("vote transaction sent", "txHash", fmt.Sprintf("%x", txHash), "nullifier", fmt.Sprintf("%x", nullifier))
	return nullifier, voterID.Address(), message.Data.Fid, nil
}
