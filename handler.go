package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

var ErrElectionUnknown = fmt.Errorf("electionID unknown")

type vocdoniHandler struct {
	cli           *apiclient.HTTPclient
	defaultCensus *CensusInfo
	webappdir     string
	db            *mongo.MongoStorage
	electionLRU   *lru.Cache[string, *api.Election]
	fcapi         farcasterapi.API

	censusCreationMap sync.Map
}

func NewVocdoniHandler(
	apiEndpoint,
	accountPrivKey string,
	census *CensusInfo,
	webappdir string,
	db *mongo.MongoStorage,
	ctx context.Context,
	fcapi farcasterapi.API,
) (*vocdoniHandler, error) {
	// Get the vocdoni account
	if accountPrivKey == "" {
		accountPrivKey = util.RandomHex(32)
		log.Infow("generated new vocdoni account", "privkey", accountPrivKey)
	}

	// Create the Vocdoni API client
	hostURL, err := url.Parse(apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse apiEndpoint: %w", err)
	}
	log.Debugf("connecting to %s", hostURL.String())
	token := uuid.New()
	cli, err := apiclient.NewHTTPclient(hostURL, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to create apiclient: %w", err)
	}
	log.Infow("using bearer token", "token", token.String())

	if err := cli.SetAccount(accountPrivKey); err != nil {
		return nil, fmt.Errorf("failed to set account: %w", err)
	}

	vh := &vocdoniHandler{
		cli:           cli,
		defaultCensus: census,
		webappdir:     webappdir,
		db:            db,
		fcapi:         fcapi,
		electionLRU: func() *lru.Cache[string, *api.Election] {
			lru, err := lru.New[string, *api.Election](100)
			if err != nil {
				log.Fatal(err)
			}
			return lru
		}(),
	}

	// Add the election callback to the mongo database to fetch the election information
	db.AddElectionCallback(vh.election)
	go vh.finalizeElectionsAtBackround(ctx)
	return vh, ensureAccountExist(cli)
}

func (v *vocdoniHandler) landing(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// check if the election is finished and if so, send the final results
	if v.checkIfElectionFinishedAndHandle(electionID, ctx) {
		return nil
	}

	election, err := v.election(electionID)
	if err != nil {
		return fmt.Errorf("failed to get election: %w", err)
	}

	if len(election.Metadata.Questions) == 0 {
		return fmt.Errorf("election has no questions")
	}

	response := strings.ReplaceAll(frame(frameMain), "{processID}", election.ElectionID.String())
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{image}", landingPNGfile(election))

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func landingPNGfile(election *api.Election) string {
	pngFile, err := imageframe.QuestionImage(election)
	if err != nil {
		log.Warnw("failed to create landing image", "error", err)
		return imageLink(imageframe.NotFoundImage())
	}
	return imageLink(pngFile)
}

func (v *vocdoniHandler) info(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the electionID from the URL and fetch the election from the vochain
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
		return fmt.Errorf("failed to fetch election: %w", err)
	}

	text := []string{}
	text = append(text, fmt.Sprintf("\nStarted at %s UTC", election.StartDate.Format("2006-01-02 15:04:05")))
	if !election.FinalResults {
		text = append(text, fmt.Sprintf("Remaining time %s", time.Until(election.EndDate).Round(time.Minute).String()))
	} else {
		text = append(text, fmt.Sprintf("The poll finalized at %s", election.EndDate.Format("2006-01-02 15:04:05")))
	}
	text = append(text, fmt.Sprintf("Poll id %x...", election.ElectionID[:16]))
	text = append(text, fmt.Sprintf("Executed on network %s", v.cli.ChainID()))
	text = append(text, fmt.Sprintf("Census hash %x...", election.Census.CensusRoot[:12]))
	if election.Census.MaxCensusSize >= uint64(maxElectionSize) {
		text = append(text, fmt.Sprintf("Allowed voters %d", election.Census.MaxCensusSize))
	} else {
		text = append(text, fmt.Sprintf("Census size %d", election.Census.MaxCensusSize))
	}

	png, err := imageframe.InfoImage(text)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	// send the response
	response := strings.ReplaceAll(frame(frameInfo), "{image}", imageLink(png))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) staticHandler(w http.ResponseWriter, r *http.Request) {
	var filePath string
	if r.URL.Path == "/app" || r.URL.Path == "/app/" {
		filePath = path.Join(v.webappdir, "index.html")
	} else {
		filePath = path.Join(v.webappdir, strings.TrimPrefix(path.Clean(r.URL.Path), "/app"))
	}
	// Serve the file using http.ServeFile
	http.ServeFile(w, r, filePath)
}

// dumpDB is a handler to dump the database contents.
func (v *vocdoniHandler) dumpDB(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	return ctx.Send([]byte(v.db.String()), http.StatusOK)
}

// importDB is a handler to import the database contents produced by dumpDB.
func (v *vocdoniHandler) importDB(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	if err := v.db.Import(msg.Data); err != nil {
		return fmt.Errorf("failed to import database: %w", err)
	}
	return ctx.Send(nil, http.StatusOK)
}
