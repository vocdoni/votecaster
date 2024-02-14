package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vocdoni/farcaster-poc/mongo"
	"go.vocdoni.io/dvote/apiclient"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"go.vocdoni.io/dvote/util"
)

type vocdoniHandler struct {
	cli       *apiclient.HTTPclient
	census    *CensusInfo
	webappdir string
	db        *mongo.MongoStorage
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

	// Create the account if it doesn't exist and return the handler
	return &vocdoniHandler{
		cli:       cli,
		census:    census,
		webappdir: webappdir,
		db:        db,
	}, ensureAccountExist(cli)
}

func (v *vocdoniHandler) landing(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	election, err := v.cli.Election(electionID)
	if err != nil {
		return fmt.Errorf("failed to get election: %w", err)
	}

	if len(election.Metadata.Questions) == 0 {
		return fmt.Errorf("election has no questions")
	}
	question := election.Metadata.Questions[0].Title["default"]

	png, err := textToImage(fmt.Sprintf("> %s", question), backgrounds[BackgroundGeneric])
	if err != nil {
		return err
	}
	response := strings.ReplaceAll(frameMain, "{server}", serverURL)
	response = strings.ReplaceAll(response, "{processID}", ctx.URLParam("electionID"))
	response = strings.ReplaceAll(response, "{image}", base64.StdEncoding.EncodeToString(png))
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) showElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}
	log.Infow("received show election request", "electionID", ctx.URLParam("electionID"))

	// create a PNG image with the election description
	election, err := v.cli.Election(electionID)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	png, err := generateElectionImage(election.Metadata.Title["default"])
	if err != nil {
		return fmt.Errorf("failed to generate image: %v", err)
	}
	// send the response
	response := strings.ReplaceAll(frame(frameVote), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{processID}", ctx.URLParam("electionID"))

	r := election.Metadata.Questions[0].Choices
	for i := 0; i < 4; i++ {
		if len(r) > i {
			response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), r[i].Title["default"])
			continue
		}
		response = strings.ReplaceAll(response, fmt.Sprintf("{option%d}", i), "")
	}

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func generateElectionImage(title string) ([]byte, error) {
	text := strings.Builder{}
	text.WriteString(title)
	return textToImage(text.String(), backgrounds[BackgroundGeneric])
}

func (v *vocdoniHandler) info(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// get the electionID from the URL and fetch the election from the vochain
	electionID := ctx.URLParam("electionID")
	electionIDbytes, err := hex.DecodeString(electionID)
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}

	election, err := v.cli.Election(electionIDbytes)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}

	text := strings.Builder{}
	text.WriteString(" Vocdoni is the blockchain for voting\n")
	text.WriteString("--------------------------------------------\n\n")
	text.WriteString(fmt.Sprintf("> Started %s ago\n", time.Since(election.StartDate).Round(time.Minute).String()))
	text.WriteString(fmt.Sprintf("> Remaining time %s\n", time.Until(election.EndDate).Round(time.Minute).String()))
	text.WriteString(fmt.Sprintf("> Poll id %x...\n", election.ElectionID[0:16]))
	text.WriteString(fmt.Sprintf("> Census hash %x...\n", election.Census.CensusRoot[0:16]))
	text.WriteString(fmt.Sprintf("> Executed on network %s\n", v.cli.ChainID()))
	png, err := textToImage(text.String(), backgrounds[BackgroundInfo])
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	// send the response
	response := strings.ReplaceAll(frame(frameInfo), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{processID}", electionID)
	response = strings.ReplaceAll(response, "{onvote}", onvoteURL)
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

	election, err := v.cli.Election(electionIDbytes)
	if err != nil {
		log.Warnw("failed to fetch election", "error", err)
		png, err := textToImage(fmt.Sprintf("Error: %s", err.Error()), backgrounds[BackgroundGeneric])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
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
		png, err := textToImage("", backgrounds[BackgroundNotElegible])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameNotElegible), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if errors.Is(err, ErrAlreadyVoted) {
		log.Infow("participant already voted", "voterID", fmt.Sprintf("%x", voterID))
		png, err := textToImage("", backgrounds[BackgroundAlreadyVoted])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameAlreadyVoted), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{nullifier}", fmt.Sprintf("%x", nullifier))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		response = strings.ReplaceAll(response, "{explorer}", explorerURL)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	if err != nil {
		log.Warnw("failed to vote", "error", err)
		png, err := textToImage(fmt.Sprintf("Error: %s", err.Error()), backgrounds[BackgroundGeneric])
		if err != nil {
			return fmt.Errorf("failed to create image: %w", err)
		}
		response := strings.ReplaceAll(frame(frameError), "{image}", base64.StdEncoding.EncodeToString(png))
		response = strings.ReplaceAll(response, "{processID}", electionID)
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	go func() {
		if !v.db.Exists(fid) {
			if err := v.db.AddUser(fid, "", []string{}); err != nil {
				log.Errorw(err, "failed to add user to database")
			}
		}
		if err := v.db.IncreaseVoteCount(fid, electionIDbytes); err != nil {
			log.Errorw(err, "failed to increase vote count")
		}
	}()

	response := strings.ReplaceAll(frame(frameAfterVote), "{nullifier}", fmt.Sprintf("%x", nullifier))
	response = strings.ReplaceAll(response, "{processID}", electionID)
	response = strings.ReplaceAll(response, "{explorer}", explorerURL)
	png, err := textToImage("", backgrounds[BackgroundAfterVote])
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	response = strings.ReplaceAll(response, "{image}", base64.StdEncoding.EncodeToString(png))
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) results(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID := ctx.URLParam("electionID")
	log.Infow("received results request", "electionID", electionID)
	electionIDbytes, err := hex.DecodeString(electionID)
	if err != nil {
		return fmt.Errorf("failed to decode electionID: %w", err)
	}

	// get the election from the vochain and create a PNG image with the results
	election, err := v.cli.Election(electionIDbytes)
	if err != nil {
		return fmt.Errorf("failed to fetch election: %w", err)
	}
	if election.Results == nil || len(election.Results) == 0 {
		return fmt.Errorf("election results not ready")
	}

	castedVotes := new(types.BigInt)
	for i := 0; i < len(election.Results[0]); i++ {
		castedVotes.Add(castedVotes, election.Results[0][i])
	}

	text := strings.Builder{}

	// Check for division by zero error
	if castedVotes.MathBigInt().Cmp(big.NewInt(0)) == 0 {
		text.WriteString("No votes casted yet")
	} else {
		text.WriteString(fmt.Sprintf("Total votes casted: %s\n\n", castedVotes.String()))
		for _, r := range election.Metadata.Questions[0].Choices {
			votesForOption := election.Results[0][r.Value]
			percentage := new(big.Float).Quo(new(big.Float).SetInt(votesForOption.MathBigInt()), new(big.Float).SetInt(castedVotes.MathBigInt()))
			percentageFloat, _ := percentage.Float64()             // Convert to float64 for rounding
			roundedPercentage := math.Round(percentageFloat * 100) // Round to nearest whole number and multiply by 100 for percentage

			_, err := text.WriteString(fmt.Sprintf("> %s: %.0f%%\n",
				r.Title["default"],
				roundedPercentage, // Use rounded percentage here
			))
			if err != nil {
				return fmt.Errorf("failed to write results: %w", err)
			}
		}
	}

	png, err := textToImage(text.String(), backgrounds[BackgroundResults])
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	response := strings.ReplaceAll(frame(frameResults), "{image}", base64.StdEncoding.EncodeToString(png))
	response = strings.ReplaceAll(response, "{processID}", electionID)
	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

type FarcasterProfile struct {
	Bio           string
	Custody       string // hex address actually
	DisplayName   string `json:"displayName"`
	FID           uint64 `json:"fid"` // BigInt probably
	Avatar        string `json:"pfpUrl"`
	Username      string
	Verifications []string // hex addresses too
}

type ElectionCreateRequest struct {
	ElectionDescription // embedded... not sure how to handle this with json unless I Unmarshal it myself
	Profile             FarcasterProfile
}

func (v *vocdoniHandler) createElection(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	var req *ElectionDescription
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal election request: %w", err)
	}
	if req.Duration == 0 {
		req.Duration = time.Hour * 24
	} else {
		req.Duration *= time.Hour
	}
	electionID, err := createElection(v.cli, req, v.census)
	if err != nil {
		return fmt.Errorf("failed to create election: %v", err)
	}

	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send([]byte(electionID.String()), http.StatusOK)
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
	if strings.HasSuffix(p, ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(p, ".html") {
		contentType = "text/html"
	} else {
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
