package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vocdoni/vote-frame/communityhub"
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
	if election.Results == nil || len(election.Results) == 0 {
		return errorImageResponse(ctx, fmt.Errorf("election results not ready"))
	}

	// get the election from the database if exist, else just use tokenDecimals = 0
	tokenDecimals := uint32(0)
	electiondb, err := v.db.Election(electionIDbytes)
	if err == nil {
		tokenDecimals = electiondb.CensusERC20TokenDecimals
	}

	// Update the results on the database
	choices, votes := extractResults(election, tokenDecimals)
	if err := v.db.SetPartialResults(electionIDbytes, choices, bigIntsToStrings(votes)); err != nil {
		return fmt.Errorf("failed to update results: %w", err)
	}

	// Update LRU cached election
	evicted := v.electionLRU.Add(electionID, election)
	log.Debugw("updated election cache", "electionID", electionID, "evicted", evicted)

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
	}

	// if not final results, create the dynamic PNG image with the results
	response := strings.ReplaceAll(frame(frameResults), "{image}", resultsPNGfile(election, electiondb.CensusERC20TokenDecimals))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
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
	erc20TokenDecimals := uint32(0)
	if electiondb != nil {
		erc20TokenDecimals = electiondb.CensusERC20TokenDecimals
	}
	id, err := imageframe.ResultsImage(election, erc20TokenDecimals)
	if err != nil {
		return "", fmt.Errorf("failed to create image: %w", err)
	}
	go func() {
		choices, votes := extractResults(election, erc20TokenDecimals)
		if err := v.db.AddFinalResults(election.ElectionID, imageframe.FromCache(id), choices, bigIntsToStrings(votes)); err != nil {
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

	// check if the election is from a community, else return silently
	if electiondb.Community == nil {
		return nil
	}

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
		return fmt.Errorf("failed to fetch census from the database: %w", err)
	}

	root, err := hex.DecodeString(census.Root)
	if err != nil {
		return fmt.Errorf("failed to decode census root: %w", err)
	}

	// Extract the choices from the election results
	totalVotingPower, _ := new(big.Int).SetString(census.TotalWeight, 10)
	hubResults := &communityhub.HubResults{
		Question:         electiondb.Question,
		Options:          choices,
		Date:             time.Now().String(),
		Tally:            tally,
		Turnout:          calculateTurnout(census.TotalWeight, electiondb.CastedWeight),
		TotalVotingPower: totalVotingPower,
		Participants:     participants,
		CensusRoot:       root,
		CensusURI:        census.URL,
	}
	log.Infow("results",
		"electionID", electiondb.ElectionID,
		"communityID", electiondb.Community.ID,
		"hubResults", hubResults)
	if err := v.comhub.SetResults(electiondb.Community.ID, electionID, hubResults); err != nil {
		return fmt.Errorf("failed to set results on the community hub: %w", err)
	}
	log.Infow("final results sent to the community hub",
		"communityID", electiondb.Community.ID,
		"electionID", electiondb.ElectionID)

	return nil
}

func resultsPNGfile(election *api.Election, tokenDecimals uint32) string {
	resultsPNGgenerationMutex.Lock()
	defer resultsPNGgenerationMutex.Unlock()
	id, err := imageframe.ResultsImage(election, tokenDecimals)
	if err != nil {
		log.Warnw("failed to create results image", "error", err)
		return imageLink(imageframe.NotFoundImage())
	}
	return imageLink(id)
}

func extractResults(election *api.Election, censusTokenDecimals uint32) (choices []string, results []*big.Int) {
	if election == nil || election.Metadata == nil || election.Results == nil {
		return nil, nil // Return nil if the main structures are nil
	}

	questions := election.Metadata.Questions
	if len(questions) == 0 || len(questions[0].Choices) == 0 || len(election.Results) == 0 {
		return nil, nil // Return nil if there are no questions or choices or results
	}

	firstQuestionChoices := questions[0].Choices
	firstResult := election.Results[0]
	if len(firstResult) == 0 {
		return nil, nil // Return nil if the results for the first question are empty
	}

	for _, option := range firstQuestionChoices {
		defaultTitle, ok := option.Title["default"]
		if !ok {
			continue // Skip if there's no default title
		}
		choices = append(choices, defaultTitle)

		if option.Value >= uint32(len(firstResult)) || firstResult[option.Value] == nil {
			results = append(results, nil) // Append nil if the index is out of range or the result is nil
			continue
		}

		bigIntResult := firstResult[option.Value].MathBigInt()
		if censusTokenDecimals > 0 {
			// Scale the result down based on the number of decimals
			scalingFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(censusTokenDecimals)), nil)
			bigIntResult = new(big.Int).Div(bigIntResult, scalingFactor)
		}
		results = append(results, bigIntResult)
	}
	return choices, results
}

// calculateTurnout computes the turnout percentage from two big.Int strings.
// If the strings are not valid numbers, it returns zero.
func calculateTurnout(totalWeightStr, castedWeightStr string) *big.Int {
	totalWeight := new(big.Int)
	castedWeight := new(big.Int)

	_, ok := totalWeight.SetString(totalWeightStr, 10)
	if !ok {
		return big.NewInt(0)
	}

	_, ok = castedWeight.SetString(castedWeightStr, 10)
	if !ok {
		return big.NewInt(0)
	}

	// Multiply castedWeight by 100 to preserve integer properties during division
	castedWeightMul := new(big.Int).Mul(castedWeight, big.NewInt(100))

	// Compute the turnout percentage as an integer if the total weight is not zero
	if totalWeight.Cmp(big.NewInt(0)) == 0 {
		log.Error("total weight is zero")
		return big.NewInt(0)
	}
	turnoutPercentage := new(big.Int).Div(castedWeightMul, totalWeight)

	return turnoutPercentage
}
