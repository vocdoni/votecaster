package main

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"go.vocdoni.io/proto/build/go/models"

	"github.com/ethereum/go-ethereum/common"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/util"
	"go.vocdoni.io/dvote/vochain/state"
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
func vote(packet *FrameSignaturePacket, electionID types.HexBytes, root []byte, cli *apiclient.HTTPclient) (types.HexBytes, types.HexBytes, error) {
	// fetch the public key from the signature and generate the census proof
	_, pubKey, err := VerifyFrameSignature(packet)
	if err != nil {
		return nil, nil, ErrFrameSignature
	}

	// compute the voterID, based on the public key
	voterID := state.NewVoterID(state.VoterIDTypeEd25519, pubKey)
	log.Infow("received vote request", "electionID", electionID, "voterID", fmt.Sprintf("%x", voterID.Address()))

	// check if the voter is elegible to vote (in the census)
	proof, err := cli.CensusGenProof(root, voterID.Address())
	if err != nil {
		return nil, voterID.Address(), ErrNotInCensus
	}

	// compute the nullifier for the vote (a hash of the voterID and the electionID)
	nullifier := state.GenerateNullifier(common.Address(voterID.Address()), electionID)

	// check if the voter already voted
	_, code, err := cli.Request("GET", nil, "votes", "verify", electionID.String(), fmt.Sprintf("%x", nullifier))
	if err != nil {
		return nullifier, voterID.Address(), fmt.Errorf("failed to verify vote: %w", err)
	}
	if code == http.StatusOK {
		return nullifier, voterID.Address(), ErrAlreadyVoted
	}

	// build the vote package
	votePackage := &state.VotePackage{
		Votes: []int{packet.UntrustedData.ButtonIndex},
	}
	votePackageBytes, err := votePackage.Encode()
	if err != nil {
		return nullifier, voterID.Address(), fmt.Errorf("failed to encode vote package: %w", err)
	}

	// build the vote transaction
	vote := &models.VoteEnvelope{
		Nonce:       util.RandomBytes(16),
		ProcessId:   electionID,
		VotePackage: votePackageBytes,
	}

	// build the proof for the vote transaction
	frameSignedMessage, err := hex.DecodeString(packet.TrustedData.MessageBytes)
	if err != nil {
		return nullifier, voterID.Address(), fmt.Errorf("failed to decode frame signed message: %w", err)
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
		return nullifier, voterID.Address(), fmt.Errorf("failed to marshal vote transaction: %w", err)
	}

	txHash, nullifier, err := cli.SignAndSendTx(&stx)
	if err != nil {
		return nullifier, voterID.Address(), fmt.Errorf("failed to sign and send vote transaction: %w", err)
	}

	log.Infow("vote transaction sent", "txHash", fmt.Sprintf("%x", txHash), "nullifier", fmt.Sprintf("%x", nullifier))
	return nullifier, voterID.Address(), nil
}
