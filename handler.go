package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"github.com/vocdoni/farcaster-poc/mongo"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/util"
)

var (
	ErrElectionUnknown = fmt.Errorf("electionID unknown")
)

type vocdoniHandler struct {
	cli         *apiclient.HTTPclient
	census      *CensusInfo
	webappdir   string
	db          *mongo.MongoStorage
	electionLRU *lru.Cache
}

func NewVocdoniHandler(apiEndpoint, accountPrivKey string, census *CensusInfo, webappdir string, db *mongo.MongoStorage) (*vocdoniHandler, error) {
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

	if err := loadImages(); err != nil {
		return nil, fmt.Errorf("failed to load images: %w", err)
	}

	vh := &vocdoniHandler{
		cli:       cli,
		census:    census,
		webappdir: webappdir,
		db:        db,
		electionLRU: func() *lru.Cache {
			lru, err := lru.New(100)
			if err != nil {
				log.Fatal(err)
			}
			return lru
		}(),
	}

	// Add the election callback to the mongo database to fetch the election information
	db.AddElectionCallback(vh.election)

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

	png, err := textToImage(electionImageContents(election), backgrounds[BackgroundGeneric])
	if err != nil {
		return err
	}
	response := strings.ReplaceAll(frame(frameMain), "{processID}", ctx.URLParam("electionID"))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{image}", base64.StdEncoding.EncodeToString(png))
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) showElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	electionIDbytes, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	// check if the election is finished and if so, send the final results
	if v.checkIfElectionFinishedAndHandle(electionIDbytes, ctx) {
		return nil
	}

	log.Infow("received show election request", "electionID", ctx.URLParam("electionID"))

	// create a PNG image with the election description
	election, err := v.election(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	png, err := textToImage(electionImageContents(election), backgrounds[BackgroundGeneric])
	if err != nil {
		return fmt.Errorf("failed to generate image: %v", err)
	}
	// send the response
	response := strings.ReplaceAll(frame(frameVote), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", ctx.URLParam("electionID"))

	r := election.Metadata.Questions[0].Choices
	for i := 0; i < 4; i++ {
		if len(r) > i {
			opt := ""
			switch i {
			case 0:
				opt = "1️⃣"
			case 1:
				opt = "2️⃣"
			case 2:
				opt = "3️⃣"
			case 3:
				opt = "4️⃣"
			}
			response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), opt)
			continue
		}
		response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), "")
	}

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func generateElectionImage(title string) ([]byte, error) {
	return textToImage(textToImageContents{title: title}, backgrounds[BackgroundGeneric])
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

	title := "Vocdoni is the blockchain for voting"
	text := []string{}
	text = append(text, fmt.Sprintf("\nStarted %s ago", time.Since(election.StartDate).Round(time.Minute).String()))
	if !election.FinalResults {
		text = append(text, fmt.Sprintf("Remaining time %s", time.Until(election.EndDate).Round(time.Minute).String()))
	}
	text = append(text, fmt.Sprintf("Poll id %x...", election.ElectionID[:20]))
	text = append(text, fmt.Sprintf("Executed on network %s", v.cli.ChainID()))
	text = append(text, fmt.Sprintf("Census hash %x...", election.Census.CensusRoot[:20]))
	text = append(text, fmt.Sprintf("Max allowed votes %d", election.Census.MaxCensusSize))
	png, err := textToImage(textToImageContents{title: title, results: text}, backgrounds[BackgroundInfo])
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	// send the response
	response := strings.ReplaceAll(frame(frameInfo), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
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
	if err != nil {
		log.Warnw("failed to fetch election", "error", err)
		png, err := errorImage(err)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
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
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
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
		png, err := textToImage(textToImageContents{title: ""}, backgrounds[BackgroundNotElegible])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameNotElegible), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if errors.Is(err, ErrAlreadyVoted) {
		log.Infow("participant already voted", "voterID", fmt.Sprintf("%x", voterID))
		png, err := textToImage(textToImageContents{title: ""}, backgrounds[BackgroundAlreadyVoted])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameAlreadyVoted), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{nullifier}", fmt.Sprintf("%x", nullifier))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if err != nil {
		log.Warnw("failed to vote", "error", err)
		png, err := textToImage(textToImageContents{title: fmt.Sprintf("Error: %s", err.Error())}, backgrounds[BackgroundGeneric])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	go func() {
		if !v.db.UserExists(fid) {
			if err := v.db.AddUser(fid, "", []string{}, 0); err != nil {
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
	png, err := textToImage(textToImageContents{title: ""}, backgrounds[BackgroundAfterVote])
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	response = strings.ReplaceAll(response, "{image}", base64.StdEncoding.EncodeToString(png))
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

// checkIfElectionFinishedAndHandle checks if the election is finished and if so, sends the final results.
// Returns true if the election is finished and the response was sent, false otherwise.
// The caller should return immediately after this function returns true.
func (v *vocdoniHandler) checkIfElectionFinishedAndHandle(electionID types.HexBytes, ctx *httprouter.HTTPContext) bool {
	pngResults := v.db.FinalResultsPNG(electionID)
	if len(pngResults) == 0 {
		return false
	}
	response := strings.ReplaceAll(frame(frameFinalResults), "{image}", base64.StdEncoding.EncodeToString(pngResults))
	response = strings.ReplaceAll(response, "{processID}", electionID.String())
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
	// Update LRU cached election
	evicted := v.electionLRU.Add(electionID, election)
	log.Debugw("updated election cache", "electionID", electionID, "evicted", evicted)

	buildResultsPNG := func(election *api.Election) ([]byte, error) {
		castedVotes := uint64(0)
		for i := 0; i < len(election.Results[0]); i++ {
			castedVotes += (election.Results[0][i].MathBigInt().Uint64())
		}
		var text []string
		var logResults []uint64
		title := fmt.Sprintf("> %s", election.Metadata.Questions[0].Title["default"])
		// Check for division by zero error
		if castedVotes == 0 {
			text = []string{"No votes casted yet..."}
		} else {
			text = []string{fmt.Sprintf("Total votes casted: %d\n", castedVotes)}
			for i, r := range election.Metadata.Questions[0].Choices {
				votesForOption := election.Results[0][r.Value].MathBigInt().Uint64()
				percentage := float64(votesForOption) * 100 / float64(castedVotes)
				text = append(text, (fmt.Sprintf("%d. %s",
					i+1,
					r.Title["default"],
				)))
				text = append(text, generateProgressBar(percentage))
				logResults = append(logResults, votesForOption)
			}
		}
		log.Debugw("election results", "castedVotes", castedVotes, "results", logResults)

		png, err := textToImage(textToImageContents{title: title, body: text}, backgrounds[BackgroundResults])
		if err != nil {
			return nil, fmt.Errorf("failed to create image: %w", err)
		}
		return png, nil
	}

	// if final results, create the static PNG image with the results
	if election.FinalResults {
		png, err := buildResultsPNG(election)
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		if err := v.db.AddFinalResults(electionIDbytes, png); err != nil {
			return fmt.Errorf("failed to add final results to database: %w", err)
		}
		if v.checkIfElectionFinishedAndHandle(electionIDbytes, ctx) {
			return nil
		}
	}
	// if not final results, create the dynamic PNG image with the results
	png, err := buildResultsPNG(election)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	response := strings.ReplaceAll(frame(frameResults), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{title}", election.Metadata.Title["default"])
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) createElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var req *ElectionCreateRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal election request: %w", err)
	}

	if req.Duration == 0 {
		req.Duration = time.Hour * 24
	} else {
		req.Duration *= time.Hour
		if req.Duration > maxElectionDuration {
			return fmt.Errorf("election duration too long")
		}
	}

	electionID, err := createElection(v.cli, &req.ElectionDescription, v.census)
	if err != nil {
		return fmt.Errorf("failed to create election: %v", err)
	}

	go func() {
		election, err := waitForElection(v.cli, electionID)
		if err != nil {
			log.Errorw(err, "failed to create election")
			return
		}
		// add the election to the LRU cache and the database
		v.electionLRU.Add(electionID.String(), election)
		if err := v.db.AddElection(electionID, req.Profile.FID); err != nil {
			log.Errorw(err, "failed to add election to database")
		}
		u, err := v.db.User(req.Profile.FID)
		if err != nil {
			if !errors.Is(err, mongo.ErrUserUnknown) {
				log.Errorw(err, "failed to get user from database")
				return
			}
			if err := v.db.AddUser(req.Profile.FID, req.Profile.Username, req.Profile.Verifications, 1); err != nil {
				log.Errorw(err, "failed to add user to database")
			}
			return
		}
		u.Addresses = req.Profile.Verifications
		u.Username = req.Profile.Username
		u.ElectionCount++
		if err := v.db.UpdateUser(u); err != nil {
			log.Errorw(err, "failed to update user in database")
		}
	}()

	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send([]byte(electionID.String()), http.StatusOK)
}

func (v *vocdoniHandler) checkElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	if electionID == nil {
		return ctx.Send(nil, http.StatusNoContent)
	}
	_, ok := v.electionLRU.Get(fmt.Sprintf("%x", electionID))
	if !ok {
		return ctx.Send(nil, http.StatusNoContent)
	}
	return ctx.Send(nil, http.StatusOK)
}

func (v *vocdoniHandler) preview(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	election, err := v.election(electionID)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to get election: %w", err))
	}

	if len(election.Metadata.Questions) == 0 {
		return errorImageResponse(ctx, fmt.Errorf("election has no questions"))
	}

	png, err := textToImage(electionImageContents(election), backgrounds[BackgroundGeneric])
	if err != nil {
		return errorImageResponse(ctx, err)
	}

	// set png headers and return response as is
	return imageResponse(ctx, png)
}

func imageResponse(ctx *httprouter.HTTPContext, png []byte) error {
	ctx.Writer.Header().Set("Content-Type", "image/png")
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(png)))
	_, err := ctx.Writer.Write(png)

	return err
}

func errorImageResponse(ctx *httprouter.HTTPContext, err error) error {
	png, err := errorImage(err)
	if err != nil {
		return err
	}

	return imageResponse(ctx, png)
}

func electionImageContents(election *api.Election) textToImageContents {
	title := election.Metadata.Questions[0].Title["default"]
	var questions []string
	for k, option := range election.Metadata.Questions[0].Choices {
		questions = append(questions, fmt.Sprintf("%d. %s", k+1, option.Title["default"]))
	}
	return textToImageContents{title: title, body: questions}
}

func (v *vocdoniHandler) testImage(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	if ctx.Request.Method == http.MethodGet {
		png, err := generateElectionImage("How would you like to take kiwi in Mumbai?")
		if err != nil {
			return err
		}
		response := strings.ReplaceAll(frame(testImageHTML), "{image}", base64.StdEncoding.EncodeToString(png))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}
	description := &ElectionDescription{}
	if err := json.Unmarshal(msg.Data, description); err != nil {
		return fmt.Errorf("failed to unmarshal election description: %w", err)
	}
	png, err := generateElectionImage(description.Question)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	jresponse, err := json.Marshal(map[string]string{"image": base64.StdEncoding.EncodeToString(png)})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

// Note: I know this is not the way to serve static files... the http.ServeFile function should be used.
// however for some reason it does not work for the index.html file (it does for any other file!).
// So I'm using this workaround for now...
func (v *vocdoniHandler) staticHandler(w http.ResponseWriter, r *http.Request) {
	var p string
	if r.URL.Path == "/app" || r.URL.Path == "/app/" {
		p = path.Join(v.webappdir, "index.html")
	} else {
		p = path.Join(v.webappdir, strings.TrimPrefix(path.Clean(r.URL.Path), "/app"))
	}

	// Open the file
	file, err := os.Open(p)
	if err != nil {
		// If the file does not exist or there's an error, return 404
		http.Error(w, "Nothing here...", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set the Content-Type header
	var contentType string
	switch {
	case strings.HasSuffix(p, ".js"):
		contentType = "application/javascript"
	case strings.HasSuffix(p, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(p, ".html"):
		contentType = "text/html"
	case strings.HasSuffix(p, ".svg"):
		contentType = "image/svg+xml"
	default:
		// Read the first 512 bytes to pass to DetectContentType
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			// If there's an error reading the file, return an internal server error
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Reset the read pointer back to the start of the file
		if _, err := file.Seek(0, 0); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Detect the content type and set the Content-Type header
		contentType = http.DetectContentType(buf[:n])
	}

	w.Header().Set("Content-Type", contentType)

	// Write the file content to the response
	_, err = io.Copy(w, file)
	if err != nil {
		log.Warnf("error writing file to response: %v", err.Error())
	}
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

func (v *vocdoniHandler) rankingByElectionsCreated(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	users, err := v.db.UsersByElectionNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"users": users,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) rankingByVotes(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	users, err := v.db.UsersByVoteNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"users": users,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) rankingOfElections(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	elections, err := v.db.ElectionsByVoteNumber()
	if err != nil {
		return fmt.Errorf("failed to get ranking: %w", err)
	}
	jresponse, err := json.Marshal(map[string]any{
		"polls": elections,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send(jresponse, http.StatusOK)
}

func (v *vocdoniHandler) election(electionID types.HexBytes) (*api.Election, error) {
	electionCached, ok := v.electionLRU.Get(electionID.String())
	if ok {
		return electionCached.(*api.Election), nil
	}
	election, err := v.cli.Election(electionID)
	if err != nil {
		return nil, ErrElectionUnknown
	}
	evicted := v.electionLRU.Add(electionID.String(), election)
	log.Debugw("added election to cache", "electionID", electionID, "evicted", evicted)
	return election, nil
}
