package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/flag"

	urlapi "go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

func main() {
	tlsDomain := flag.String("tlsDomain", "", "The domain to use for the TLS certificate")
	tlsDirCert := flag.String("tlsDirCert", "", "The directory to use for the TLS certificate")
	host := flag.String("listenHost", "", "The host to listen on")
	port := flag.Int("listenPort", 0, "The port to listen on")
	dataDir := flag.String("dataDir", "", "The directory to use for the data")
	apiEndpoint := flag.String("apiEndpoint", "https://api-dev.vocdoni.net/v2", "The Vocdoni API endpoint to use")
	vocdoniPrivKey := flag.String("vocdoniPrivKey", "", "The Vocdoni private key to use for orchestrating the election (hex)")
	logLevel := flag.String("logLevel", "info", "The log level to use")

	// Parse the command line flags
	flag.Parse()
	log.Init(*logLevel, "stdout", nil)

	handler, err := NewVocdoniHandler(*apiEndpoint, *vocdoniPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	// Create the test census
	censusInfo, err := createTestCensus(handler.cli)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new test election
	electionID, err := createElection(handler.cli, &electionDescription{
		question: "How do you take kiwi?",
		choices:  []string{"skin on", "skin off", "I don't like kiwi"},
		duration: 24 * time.Hour,
	}, censusInfo)
	if err != nil {
		log.Fatal(err)
	}

	// Create the HTTP API router
	router := new(httprouter.HTTProuter)
	router.TLSdomain = *tlsDomain
	router.TLSdirCert = *tlsDirCert
	if err := router.Init(*host, *port); err != nil {
		log.Fatal(err)
	}

	uAPI, err := urlapi.NewAPI(router, "/", *dataDir, db.TypePebble)
	if err != nil {
		log.Fatal(err)
	}

	// Register the API methods
	if err := uAPI.Endpoint.RegisterMethod("/", http.MethodGet, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(strings.ReplaceAll(frameMain, "{processID}", electionID.String())), http.StatusOK)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/", http.MethodPost, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(strings.ReplaceAll(frameMain, "{processID}", electionID.String())), http.StatusOK)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/router/{electionID}", http.MethodPost, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		electionID := ctx.URLParam("electionID")
		packet := &FrameSignaturePacket{}
		if err := json.Unmarshal(msg.Data, packet); err != nil {
			return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
		}
		redirectURL := ""
		switch packet.UntrustedData.ButtonIndex {
		case 1:
			redirectURL = fmt.Sprintf("https://celoni.vocdoni.net/poll/results/%s", electionID)
		case 2:
			redirectURL = fmt.Sprintf("https://celoni.vocdoni.net/poll/%s", electionID)
		default:
			redirectURL = "https://celoni.vocdoni.net/"
		}
		log.Infow("received router request", "electionID", electionID, "buttonIndex", packet.UntrustedData.ButtonIndex, "redirectURL", redirectURL)
		ctx.Writer.Header().Add("Location", redirectURL)
		return ctx.Send([]byte(redirectURL), http.StatusTemporaryRedirect)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll/results/{electionID}", http.MethodPost, "public", handler.results); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll/{electionID}", http.MethodPost, "public", handler.showElection); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/vote/{electionID}", http.MethodPost, "public", handler.vote); err != nil {
		log.Fatal(err)
	}

	// close if interrupt received
	log.Infof("startup complete at %s", time.Now().Format(time.RFC850))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Warnf("received SIGTERM, exiting at %s", time.Now().Format(time.RFC850))
}
