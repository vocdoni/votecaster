package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vocdoni/vote-frame/communityhub"
	"github.com/vocdoni/vote-frame/helpers"
	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

var resultsPNGgenerationMutex = sync.Mutex{}

// checkIfElectionFinishedAndHandle checks if the election is finished and if so, sends the final results.
// Returns true if the election is finished and the response was sent, false otherwise.
// The caller should return immediately after this function returns true.
func (v *vocdoniHandler) checkIfElectionFinishedAndHandle(electionID types.HexBytes, ctx *httprouter.HTTPContext) bool {
	pngResults := v.db.FinalResultsPNG(electionID)
	if pngResults == nil {
		return false
	}
	response := strings.ReplaceAll(frame(frameFinalResults), "{image}", imageLink(imageframe.AddImageToCache(pngResults)))
	response = strings.ReplaceAll(response, "{processID}", electionID.String())
	response = strings.ReplaceAll(response, "{title}", "Final results")

	ctx.SetResponseContentType("text/html; charset=utf-8")
	if err := ctx.Send([]byte(response), http.StatusOK); err != nil {
		log.Warnw("failed to send response", "error", err)
		return true
	}
	return true
}

func (v *vocdoniHandler) results(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID := ctx.URLParam("electionID")
	if len(electionID) == 0 {
		return errorImageResponse(ctx, fmt.Errorf("invalid electionID"))
	}
	log.Infow("received results request", "electionID", electionID)
	electionIDbytes, err := hex.DecodeString(electionID)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to decode electionID: %w", err))
	}
	// check if the election is finished and if so, send the final results as a static PNG
	if v.checkIfElectionFinishedAndHandle(electionIDbytes, ctx) {
		return nil
	}

	// get the election from the vochain and create a PNG image with the results
	election, err := v.cli.Election(electionIDbytes)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to fetch election: %w", err))
	}
	metadata := helpers.UnpackMetadata(election.Metadata)
	if election.Results == nil || len(election.Results) == 0 {
		return errorImageResponse(ctx, fmt.Errorf("election results not ready"))
	}

	electiondb, err := v.db.Election(electionIDbytes)
	if err != nil {
		log.Warnw("failed to fetch election from database", "error", err)
	}

	// if final results, create the static PNG image with the results
	if election.FinalResults {
		id, err := v.finalizeElectionResults(election, electiondb)
		if err != nil {
			return errorImageResponse(ctx, fmt.Errorf("failed to create final results: %w", err))
		}
		response := strings.ReplaceAll(frame(frameFinalResults), "{image}", imageLink(id))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", "Final results")

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	} else {
		_, err := v.updateAndFetchResultsFromDatabase(electionIDbytes, election)
		if err != nil {
			return fmt.Errorf("failed to update/fetch results: %w", err)
		}
	}

	totalWeightStr := ""
	census, err := v.db.CensusFromElection(electionIDbytes)
	if err == nil {
		totalWeightStr = census.TotalWeight
	}

	// if not final results, create the dynamic PNG image with the results
	response := strings.ReplaceAll(frame(frameResults), "{image}", resultsPNGfile(election, electiondb, totalWeightStr))
	response = strings.ReplaceAll(response, "{title}", metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

// finalizeElectionResults creates the final results image and stores it in the database.
// election is mandatory but electiondb is optional. Some features may not work if electiondb is nil.
// Returns the imageID of the final results image.
func (v *vocdoniHandler) finalizeElectionResults(election *api.Election, electiondb *mongo.Election) (imageID string, err error) {
	if election == nil || election.Metadata == nil {
		return "", fmt.Errorf("nil election or missing parameters")
	}
	if !election.FinalResults {
		return "", fmt.Errorf("election not finalized")
	}
	totalWeightStr := ""
	census, err := v.db.CensusFromElection(election.ElectionID)
	if err == nil {
		totalWeightStr = census.TotalWeight
	}

	id, err := imageframe.ResultsImage(election, electiondb, totalWeightStr)
	if err != nil {
		return "", fmt.Errorf("failed to create image: %w", err)
	}
	go func() {
		choices, votes := helpers.ExtractResults(election, 0)
		if err := v.db.AddFinalResults(election.ElectionID, imageframe.FromCache(id), choices, helpers.BigIntsToStrings(votes)); err != nil {
			log.Errorw(err, "failed to add final results to database")
			return
		}
		if electiondb != nil {
			if err := v.settleResultsIntoCommunityHub(electiondb, choices, votes); err != nil {
				log.Errorw(err, "failed to settle results into community hub")
			}
		}
	}()
	return id, nil
}

func (v *vocdoniHandler) settleResultsIntoCommunityHub(electiondb *mongo.Election, choices []string, votes []*big.Int) error {
	if len(votes) == 0 || len(choices) == 0 {
		return fmt.Errorf("invalid votes/choices")
	}

	if electiondb == nil {
		return fmt.Errorf("nil electiondb")
	}

	// check if the election is from a community, else return silently
	if electiondb.Community == nil {
		return nil
	}

	log.Debugw("settling results into community hub", "electionID", electiondb.ElectionID, "communityID", electiondb.Community.ID)

	// send the final results to the community hub if electiondb is not nil
	// check if community exists in the smart contract
	comm, err := v.comhub.Community(electiondb.Community.ID)
	if err != nil || comm == nil {
		return fmt.Errorf("failed to fetch community from the community hub: %w", err)
	}
	electionID, err := hex.DecodeString(electiondb.ElectionID)
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// Extract the list of participants from the database
	voters, err := v.db.VotersOfElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch voters from the database: %w", err)
	}
	participants := []*big.Int{}
	for _, voter := range voters {
		participants = append(participants, new(big.Int).SetUint64(voter.UserID))
	}

	// Transform the results into a format suitable for the community hub
	tally := [][]*big.Int{votes}

	// We need the census to calculate the turnout
	census, err := v.db.CensusFromElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch census from the database for election %x: %w", electionID, err)
	}

	root, err := hex.DecodeString(census.Root)
	if err != nil {
		return fmt.Errorf("failed to decode census root: %w", err)
	}

	// Create a new big.Int from the truncated float
	turnout := big.NewInt(int64(math.Trunc(float64(helpers.CalculateTurnout(census.TotalWeight, electiondb.CastedWeight)))))

	// Extract the choices from the election results
	totalVotingPower, _ := new(big.Int).SetString(census.TotalWeight, 10)
	hubResults := &communityhub.HubResults{
		Question:         electiondb.Question,
		Options:          choices,
		Date:             time.Now().String(),
		Tally:            tally,
		Turnout:          turnout,
		TotalVotingPower: totalVotingPower,
		Participants:     participants,
		CensusRoot:       root,
		CensusURI:        census.URL,
		VoteCount:        new(big.Int).SetUint64(electiondb.CastedVotes),
	}
	log.Infow("sending results transaction to community hub smart contract",
		"electionID", electiondb.ElectionID,
		"communityID", electiondb.Community.ID,
		"hubResults", hubResults)
	if err := v.comhub.SetResults(electiondb.Community.ID, electionID, hubResults); err != nil {
		return fmt.Errorf("failed to set results on the community hub: %w", err)
	}
	return nil
}

func resultsPNGfile(election *api.Election, electiondb *mongo.Election, totalWeightStr string) string {
	resultsPNGgenerationMutex.Lock()
	defer resultsPNGgenerationMutex.Unlock()
	id, err := imageframe.ResultsImage(election, electiondb, totalWeightStr)
	if err != nil {
		log.Warnw("failed to create results image", "error", err)
		return imageLink(imageframe.NotFoundImage())
	}
	return imageLink(id)
}

// updateAndFetchResultsFromDatabase updates the results on the database and returns them updated.
// It also updates the LRU cached election.
// If election is nil, it fetches it from the vochain API.
func (v *vocdoniHandler) updateAndFetchResultsFromDatabase(
	electionID types.HexBytes,
	election *api.Election,
) (*mongo.Results, error) {
	if election == nil {
		var err error
		election, err = v.cli.Election(electionID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch election: %w", err)
		}
	}

	// Update LRU cached election
	_ = v.electionLRU.Add(fmt.Sprintf("%x", electionID), election)

	// Update the results on the database
	choices, votes := helpers.ExtractResults(election, 0)
	votesString := helpers.BigIntsToStrings(votes)
	log.Infow("updating partial results", "electionID", electionID.String(), "choices", choices, "votes", votesString)
	if err := v.db.SetPartialResults(electionID, choices, votesString); err != nil {
		return nil, fmt.Errorf("failed to update results: %w", err)
	}

	// Fetch results from the database to return them in the response
	results, err := v.db.Results(electionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch results: %w", err)
	}

	return results, nil
}
