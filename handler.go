package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
	c3cli "github.com/vocdoni/census3/apiclient"
	"github.com/vocdoni/vote-frame/airstack"
	"github.com/vocdoni/vote-frame/communityhub"
	"github.com/vocdoni/vote-frame/farcasterapi"
	"github.com/vocdoni/vote-frame/helpers"
	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"github.com/vocdoni/vote-frame/reputation"
	"github.com/vocdoni/vote-frame/shortener"
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
	cliToken      *uuid.UUID
	apiEndpoint   *url.URL
	defaultCensus *CensusInfo
	webappdir     string
	db            *mongo.MongoStorage
	electionLRU   *lru.Cache[string, *api.Election]
	fcapi         farcasterapi.API
	airstack      *airstack.Airstack
	census3       *c3cli.HTTPclient
	comhub        *communityhub.CommunityHub
	repUpdater    *reputation.Updater

	backgroundQueue  sync.Map
	addAuthTokenFunc func(uint64, string)
	adminFID         uint64
}

func NewVocdoniHandler(
	apiEndpoint,
	accountPrivKey string,
	census *CensusInfo,
	webappdir string,
	db *mongo.MongoStorage,
	ctx context.Context,
	fcapi farcasterapi.API,
	token *uuid.UUID,
	airstack *airstack.Airstack,
	census3 *c3cli.HTTPclient,
	comhub *communityhub.CommunityHub,
	repUpdater *reputation.Updater,
	adminFID uint64,
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
	if token == nil {
		token = new(uuid.UUID)
		*token = uuid.New()
	}
	cli, err := apiclient.NewWithBearer(hostURL.String(), token)
	if err != nil {
		return nil, fmt.Errorf("failed to create apiclient: %w", err)
	}
	log.Infow("using bearer token", "token", token.String())

	if err := cli.SetAccount(accountPrivKey); err != nil {
		return nil, fmt.Errorf("failed to set account: %w", err)
	}

	vh := &vocdoniHandler{
		cli:           cli,
		cliToken:      token,
		apiEndpoint:   hostURL,
		defaultCensus: census,
		webappdir:     webappdir,
		db:            db,
		fcapi:         fcapi,
		airstack:      airstack,
		census3:       census3,
		comhub:        comhub,
		repUpdater:    repUpdater,
		adminFID:      adminFID,
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
	go finalizeElectionsAtBackround(ctx, vh)
	return vh, ensureAccountExist(cli)
}

// AddAuthTokenFunc sets the function to add an authentication token to the farcaster API.
func (v *vocdoniHandler) AddAuthTokenFunc(f func(uint64, string)) {
	v.addAuthTokenFunc = f
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

	// validate the frame package to airstack
	if v.airstack != nil {
		airstack.ValidateFrameMessage(msg.Data, v.airstack.ApiKey())
	}

	election, err := v.election(electionID)
	if err != nil {
		return fmt.Errorf("failed to get election: %w", err)
	}
	metadata := helpers.UnpackMetadata(election.Metadata)
	if len(metadata.Questions) == 0 {
		return fmt.Errorf("election has no questions")
	}

	response := strings.ReplaceAll(frame(frameMain), "{processID}", election.ElectionID.String())
	response = strings.ReplaceAll(response, "{title}", metadata.Title["default"])
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
	// validate the frame package to airstack
	if v.airstack != nil {
		airstack.ValidateFrameMessage(msg.Data, v.airstack.ApiKey())
	}

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

	text := []string{}
	title := ""
	dbElection, err := v.db.Election(electionIDbytes)
	if err != nil {
		// election not found in the database, so we use just the information from the vochain API
		election, err := v.election(electionIDbytes)
		if err != nil {
			return fmt.Errorf("failed to fetch election: %w", err)
		}
		metadata := helpers.UnpackMetadata(election.Metadata)
		censusUserCount := election.Census.MaxCensusSize
		text = append(text, fmt.Sprintf("\nStarted at %s UTC", election.StartDate.Format("2006-01-02 15:04:05")))
		if !election.FinalResults {
			text = append(text, fmt.Sprintf("Remaining time %s", time.Until(election.EndDate).Round(time.Minute).String()))
		} else {
			text = append(text, fmt.Sprintf("The poll finalized at %s", election.EndDate.Format("2006-01-02 15:04:05")))
		}
		text = append(text, fmt.Sprintf("Poll id %x...", election.ElectionID[:16]))
		text = append(text, fmt.Sprintf("Executed on network %s", v.cli.ChainID()))
		text = append(text, fmt.Sprintf("Census hash %x...", election.Census.CensusRoot[:12]))
		if censusUserCount >= uint64(maxElectionSize) {
			text = append(text, fmt.Sprintf("Allowed voters %d", censusUserCount))
		} else {
			text = append(text, fmt.Sprintf("Census size %d", censusUserCount))
		}
		title = metadata.Title["default"]
	} else {
		// election found in the database, so we use the information from the database
		text = append(text, fmt.Sprintf("\nStarted at %s UTC", dbElection.CreatedTime.Format("2006-01-02 15:04:05")))
		if time.Now().Before(dbElection.EndTime) {
			text = append(text, fmt.Sprintf("Remaining time: %s", time.Until(dbElection.EndTime).Round(time.Minute).String()))
		} else {
			text = append(text, fmt.Sprintf("The poll finalized at %s", dbElection.EndTime.Format("2006-01-02 15:04:05")))
		}
		if dbElection.Community != nil {
			text = append(text, fmt.Sprintf("Community: %s", dbElection.Community.Name))
		}
		owner, err := v.db.User(dbElection.UserID)
		if err == nil {
			text = append(text, fmt.Sprintf("Owner: %s", owner.Username))
		}
		text = append(text, fmt.Sprintf("Executed on network: %s", v.cli.ChainID()))

		if dbElection.FarcasterUserCount > 0 {
			text = append(text, fmt.Sprintf("Elegible users: %d", dbElection.FarcasterUserCount))
		}
		text = append(text, fmt.Sprintf("Last vote: %s", dbElection.LastVoteTime.Format("2006-01-02 15:04:05")))
		text = append(text, fmt.Sprintf("Cast votes: %d", dbElection.CastedVotes))

		title = dbElection.Question
	}

	png, err := imageframe.InfoImage(text)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	// send the response
	response := strings.ReplaceAll(frame(frameInfo), "{image}", imageLink(png))
	response = strings.ReplaceAll(response, "{title}", title)
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
	// import the database in the background to avoid issues when the request
	// times out and the import is not finished (for large databases)
	go func() {
		if err := v.db.Import(msg.Data); err != nil {
			log.Errorf("failed to import database: %v", err)
		}
	}()
	return ctx.Send(nil, http.StatusOK)
}

// whitelistHandler is a handler to whitelist a user.
func (v *vocdoniHandler) whitelistHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	fid, err := strconv.ParseUint(ctx.URLParam("fid"), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse fid: %w", err)
	}
	if err := v.db.SetWhiteListedForUser(fid, true); err != nil {
		return fmt.Errorf("failed to get whitelist: %w", err)
	}
	return ctx.Send(nil, http.StatusOK)
}

// shortURLHanlder is a handler to shorten a URL.
func (v *vocdoniHandler) shortURLHanlder(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	url := ctx.Request.URL.Query().Get("url")
	if url == "" {
		return ctx.Send([]byte("missing URL"), http.StatusBadRequest)
	}
	shortURL, err := shortener.ShortURL(ctx.Request.Context(), url)
	if err != nil {
		return fmt.Errorf("failed to shorten URL: %w", err)
	}
	res, err := json.Marshal(map[string]string{"result": shortURL})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	return ctx.Send(res, http.StatusOK)
}
