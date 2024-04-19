package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strings"
	"sync"

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

	// get the election from the database
	electiondb, err := v.db.Election(electionIDbytes)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to fetch election: %w", err))
	}

	// Update the results on the database
	choices, votes := extractResults(election, electiondb.CensusERC20TokenDecimals)
	if err := v.db.SetPartialResults(electionIDbytes, choices, votes); err != nil {
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
// Returns the imageID of the final results image.
func (v *vocdoniHandler) finalizeElectionResults(election *api.Election, electiondb *mongo.Election) (imageID string, err error) {
	if election == nil || electiondb == nil {
		return "", fmt.Errorf("nil election or electiondb")
	}
	if !election.FinalResults {
		return "", fmt.Errorf("election not finalized")
	}
	id, err := imageframe.ResultsImage(election, electiondb.CensusERC20TokenDecimals)
	if err != nil {
		return "", fmt.Errorf("failed to create image: %w", err)
	}
	choices, votes := extractResults(election, electiondb.CensusERC20TokenDecimals)
	go func() {
		if err := v.db.AddFinalResults(election.ElectionID, imageframe.FromCache(id), choices, votes); err != nil {
			log.Errorw(err, "failed to add final results to database")
			return
		}
		log.Infow("final results image built ondemand", "electionID", election.ElectionID.String())
	}()
	return id, nil
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

func extractResults(election *api.Election, censusTokenDecimals uint32) (choices []string, results []string) {
	for _, option := range election.Metadata.Questions[0].Choices {
		choices = append(choices, option.Title["default"])
		value := ""
		if censusTokenDecimals > 0 {
			resultsValueFloat := new(big.Float).Quo(
				new(big.Float).SetInt(election.Results[0][option.Value].MathBigInt()),
				new(big.Float).SetInt(big.NewInt(int64(math.Pow(10, float64(censusTokenDecimals))))),
			)
			value = fmt.Sprintf("%.2f", resultsValueFloat)
		} else {
			value = election.Results[0][option.Value].MathBigInt().String()
		}
		results = append(results, value)
	}
	return choices, results
}
