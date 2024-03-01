package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go.vocdoni.io/proto/build/go/models"

	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
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

func (v *vocdoniHandler) vote(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the electionID from the URL and the frame signature packet from the body of the request
	electionID := ctx.URLParam("electionID")
	electionIDbytes, err := hex.DecodeString(electionID)
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// check if the election is finished and if so, send the final results
	if v.checkIfElectionFinishedAndHandle(electionIDbytes, ctx) {
		return nil
	}

	election, err := v.election(electionIDbytes)
	if err != nil {
		log.Warnw("failed to fetch election", "error", err)
		png, err := errorImage(err)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", v.addImageToCache(png, electionIDbytes))
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
		response = strings.ReplaceAll(response, "{processID}", electionID)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if election.FinalResults {
		png, err := errorImage(errors.New("The poll has finalized"))
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", v.addImageToCache(png, electionIDbytes))
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
		response = strings.ReplaceAll(response, "{processID}", electionID)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}

	// cast the vote
	nullifier, voterID, fid, err := vote(packet, electionIDbytes, election.Census.CensusRoot, v.cli)

	// handle the vote result
	if errors.Is(err, ErrNotInCensus) {
		log.Infow("participant not in the census", "voterID", fmt.Sprintf("%x", voterID))
		png, err := textToImage(textToImageContents{title: ""}, frames[BackgroundNotElegible])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameNotElegible), "{image}", v.addImageToCache(png, electionIDbytes))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if errors.Is(err, ErrAlreadyVoted) {
		log.Infow("participant already voted", "voterID", fmt.Sprintf("%x", voterID))
		png, err := textToImage(textToImageContents{title: ""}, frames[BackgroundAlreadyVoted])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameAlreadyVoted), "{image}", v.addImageToCache(png, electionIDbytes))
		response = strings.ReplaceAll(response, "{nullifier}", fmt.Sprintf("%x", nullifier))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if err != nil {
		log.Warnw("failed to vote", "error", err)
		png, err := textToImage(textToImageContents{title: fmt.Sprintf("Error: %s", err.Error())}, frames[BackgroundGeneric])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", v.addImageToCache(png, electionIDbytes))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	go func() {
		if !v.db.UserExists(fid) {
			if err := v.db.AddUser(fid, "", []string{}, []string{}, "", 0); err != nil {
				log.Errorw(err, "failed to add user to database")
			}
		}
		if err := v.db.IncreaseVoteCount(fid, electionIDbytes); err != nil {
			log.Errorw(err, "failed to increase vote count")
		}
	}()

	response := strings.ReplaceAll(frame(frameAfterVote), "{nullifier}", fmt.Sprintf("%x", nullifier))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", electionID)
	png, err := textToImage(textToImageContents{title: ""}, frames[BackgroundAfterVote])
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	response = strings.ReplaceAll(response, "{image}", v.addImageToCache(png, electionIDbytes))
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

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
