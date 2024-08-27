package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vocdoni/vote-frame/helpers"
	"github.com/vocdoni/vote-frame/imageframe"
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
	ErrVoteDelegated  = fmt.Errorf("vote delegated")
	ErrFrameSignature = fmt.Errorf("frame signature verification failed")
)

// voteData contains the data needed to cast a vote.
type voteData struct {
	Nullifier types.HexBytes
	VoterID   state.VoterID
	FID       uint64
	Proof     *apiclient.CensusProof
	PubKey    ed25519.PublicKey
}

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
	metadata := helpers.UnpackMetadata(election.Metadata)
	if err != nil {
		log.Warnw("failed to fetch election", "error", err)
		png, err2 := imageframe.ErrorImage(err.Error())
		if err2 != nil {
			return fmt.Errorf("failed to create image: %w", err2)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", imageLink(png))
		response = strings.ReplaceAll(response, "{title}", metadata.Title["default"])
		response = strings.ReplaceAll(response, "{processID}", electionID)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if election.FinalResults {
		response := strings.ReplaceAll(frame(frameError), "{image}", imageLink(imageframe.NotFoundImage()))
		response = strings.ReplaceAll(response, "{title}", metadata.Title["default"])
		response = strings.ReplaceAll(response, "{processID}", electionID)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}
	// check if the user has delegated their vote
	dbElection, err := v.db.Election(electionIDbytes)
	if err != nil {
		response, err := handleVoteError(err, nil, electionIDbytes)
		if err != nil {
			return fmt.Errorf("failed to handle election error: %w", err)
		}
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}
	if dbElection.Community != nil {
		delegations, err := v.db.DelegationsByCommunityFrom(dbElection.Community.ID, uint64(packet.UntrustedData.FID))
		if err != nil {
			log.Warnw("failed to fetch delegations", "error", err)
		}
		if len(delegations) > 0 {
			if response, err := handleVoteError(ErrVoteDelegated, &voteData{
				FID: uint64(packet.UntrustedData.FID),
			}, electionIDbytes); err != nil {
				ctx.SetResponseContentType("text/html; charset=utf-8")
				return ctx.Send([]byte(response), http.StatusOK)
			}
		}
	}

	// get the vote count for future check
	voteCount, err := v.cli.ElectionVoteCount(electionIDbytes)
	if err != nil {
		log.Warnw("failed to fetch vote count", "error", err)
	}

	// cast the vote
	voteData, err := vote(packet, electionIDbytes, election.Census.CensusRoot, v.cli)
	// handle the error (if any)
	if response, err := handleVoteError(err, voteData, electionIDbytes); err != nil {
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	go func() {
		if !v.db.UserExists(voteData.FID) {
			if err := v.db.AddUser(voteData.FID, "", "", []string{}, []string{}, "", 0); err != nil {
				log.Errorw(err, "failed to add user to database")
			}
		}
		if err := v.db.IncreaseVoteCount(voteData.FID, electionIDbytes, voteData.Proof.LeafWeight); err != nil {
			log.Errorw(err, "failed to increase vote count")
		}

		// wait until voteCount increases or timeout
		// if voteCount increases, update the election cache and generate the new results image
		// TODO: check this is actually useful to increase the cache hit rate
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			c, err := v.cli.ElectionVoteCount(electionIDbytes)
			if err != nil {
				log.Warnw("failed to fetch vote count", "error", err)
			}
			if c > voteCount {
				election, err := v.election(electionIDbytes)
				if err != nil {
					log.Warnw("failed to fetch election", "error", err)
					break
				}
				_, err = v.updateAndFetchResultsFromDatabase(electionIDbytes, election)
				if err != nil {
					log.Warnw("failed to update results", "error", err)
				}
			}
		}
	}()

	// wait some time so the vote is processed and the results are updated
	time.Sleep(2 * time.Second)

	response := strings.ReplaceAll(frame(frameAfterVote), "{nullifier}", fmt.Sprintf("%x", voteData.Nullifier))
	response = strings.ReplaceAll(response, "{title}", metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", electionID)
	png := imageframe.AfterVoteImage()
	response = strings.ReplaceAll(response, "{image}", imageLink(png))
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func extractVoteDataAndCheckIfEligible(packet *FrameSignaturePacket, electionID types.HexBytes, root []byte, cli *apiclient.HTTPclient) (*voteData, error) {
	messageBytes, err := hex.DecodeString(packet.TrustedData.MessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message bytes: %w", err)
	}
	actionMessage, pubKey, fid, err := farcasterproof.VerifyFrameSignature(messageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to verify frame signature: %w", err)
	}

	// compute the voterID, based on the public key
	voterID := state.NewFarcasterVoterID(pubKey, fid)

	log.Infow("received vote request",
		"electionID", electionID,
		"voterID", fmt.Sprintf("%x", voterID.Address()),
		"fid", fid,
		"pubkey", fmt.Sprintf("%x", pubKey),
		"button", actionMessage.ButtonIndex,
		"url", string(actionMessage.Url),
		"state", string(actionMessage.State),
	)

	// compute the nullifier for the vote (a hash of the voterID and the electionID)
	nullifier := farcasterproof.GenerateNullifier(fid, electionID)

	// construct the vote data
	data := &voteData{
		Nullifier: nullifier,
		VoterID:   voterID,
		FID:       fid,
		PubKey:    pubKey,
	}

	// check if the voter is elegible to vote (in the census)
	data.Proof, err = cli.CensusGenProof(root, voterID.Address())
	if err != nil {
		return data, ErrNotInCensus
	}

	// check if the voter already voted
	_, code, err := cli.Request("GET", nil, "votes", "verify", electionID.String(), fmt.Sprintf("%x", nullifier))
	if err != nil {
		return data, fmt.Errorf("could not verify vote: %w", err)
	}
	if code == http.StatusOK {
		return data, ErrAlreadyVoted
	}
	return data, nil
}

// handleVoteError handles the error returned by the extractVoteDataAndCheckIfEligible function.
// Returns nil if the error is nil, otherwise returns the HTTP error message and the error itself.
// The error message is a HTML page with an image and a message that can be displayed to the user.
func handleVoteError(err error, voteData *voteData, electionID types.HexBytes) ([]byte, error) {
	// handle the vote result
	if errors.Is(err, ErrNotInCensus) {
		log.Debugw("participant not in the census",
			"electionID", electionID.String(),
			"fid", voteData.FID,
			"pubkey", fmt.Sprintf("%x", voteData.PubKey),
			"voterID", fmt.Sprintf("%x", voteData.VoterID.Address()),
		)
		png := imageframe.NotElegibleImage()
		response := strings.ReplaceAll(frame(frameNotElegible), "{image}", imageLink(png))
		response = strings.ReplaceAll(response, "{processID}", electionID.String())
		return []byte(response), ErrNotInCensus
	}

	if errors.Is(err, ErrAlreadyVoted) {
		log.Debugw("participant already voted",
			"electionID", electionID.String(),
			"fid", voteData.FID,
			"pubkey", fmt.Sprintf("%x", voteData.PubKey),
			"nullifier", fmt.Sprintf("%x", voteData.Nullifier),
		)
		png := imageframe.AlreadyVotedImage()
		response := strings.ReplaceAll(frame(frameAlreadyVoted), "{image}", imageLink(png))
		response = strings.ReplaceAll(response, "{nullifier}", fmt.Sprintf("%x", voteData.Nullifier))
		response = strings.ReplaceAll(response, "{processID}", electionID.String())
		return []byte(response), ErrAlreadyVoted
	}

	if errors.Is(err, ErrVoteDelegated) {
		log.Debugw("participant already delegated vote",
			"electionID", electionID.String(),
			"fid", voteData.FID,
		)
		png := imageframe.AlreadyDelegated()
		response := strings.ReplaceAll(frame(frameDelegatedVote), "{image}", imageLink(png))
		response = strings.ReplaceAll(response, "{processID}", electionID.String())
		return []byte(response), ErrVoteDelegated
	}

	if err != nil {
		log.Warnw("failed to vote", "error", err)
		png, err2 := imageframe.ErrorImage(err.Error())
		if err2 != nil {
			return nil, fmt.Errorf("failed to create image: %w", err2)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", imageLink(png))
		response = strings.ReplaceAll(response, "{processID}", electionID.String())
		return []byte(response), err
	}
	return nil, nil
}

// vote creates a vote transaction, including the frame signature packet and sends it to the vochain.
// It returns the nullifier of the vote (which is the unique identifier of the vote), the voterID and an error.
func vote(packet *FrameSignaturePacket, electionID types.HexBytes, root []byte, cli *apiclient.HTTPclient) (*voteData, error) {
	voteData, err := extractVoteDataAndCheckIfEligible(packet, electionID, root, cli)
	if err != nil {
		return nil, err
	}

	// build the vote package
	votePackage := &state.VotePackage{
		Votes: []int{packet.UntrustedData.ButtonIndex - 1},
	}
	votePackageBytes, err := votePackage.Encode()
	if err != nil {
		return voteData, fmt.Errorf("failed to encode vote package: %w", err)
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
		return voteData, fmt.Errorf("failed to decode frame signed message: %w", err)
	}

	arboProof := &models.ProofArbo{
		Type:            models.ProofArbo_BLAKE2B,
		Siblings:        voteData.Proof.Proof,
		AvailableWeight: voteData.Proof.LeafValue,
		KeyType:         voteData.Proof.KeyType,
		VoteWeight:      voteData.Proof.LeafValue,
	}

	vote.Proof = &models.Proof{
		Payload: &models.Proof_FarcasterFrame{
			FarcasterFrame: &models.ProofFarcasterFrame{
				SignedFrameMessageBody: frameSignedMessage,
				PublicKey:              voteData.PubKey,
				CensusProof:            arboProof,
			},
		},
	}

	// sign and send the vote transaction
	tx, err := proto.Marshal(&models.Tx{Payload: &models.Tx_Vote{Vote: vote}})
	if err != nil {
		return voteData, fmt.Errorf("failed to marshal vote transaction: %w", err)
	}

	txHash, nullifier, err := cli.SignAndSendTx(tx)
	if err != nil {
		return voteData, fmt.Errorf("failed to sign and send vote transaction: %w", err)
	}

	log.Infow("vote transaction sent", "txHash", fmt.Sprintf("%x", txHash), "nullifier", fmt.Sprintf("%x", nullifier))
	return voteData, nil
}
