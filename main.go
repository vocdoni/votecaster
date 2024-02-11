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

var serverURL = "http://localhost:8888"

func main() {
	server := flag.String("server", serverURL, "The full URL of the server (http or https)")
	tlsDomain := flag.Bool("tlsDomain", false, "Should a TLS certificate be fetched from letsencrypt for the domain? (requires port 443)")
	tlsDirCert := flag.String("tlsDirCert", "", "The directory to use to store the TLS certificate")
	host := flag.String("listenHost", "0.0.0.0", "The host to listen on")
	port := flag.Int("listenPort", 8888, "The port to listen on")
	dataDir := flag.String("dataDir", "", "The directory to use for the data")
	apiEndpoint := flag.String("apiEndpoint", "https://api-dev.vocdoni.net/v2", "The Vocdoni API endpoint to use")
	vocdoniPrivKey := flag.String("vocdoniPrivKey", "", "The Vocdoni private key to use for orchestrating the election (hex)")
	censusFromFile := flag.String("censusFromFile", "farcaster_census.json", "Take census details from JSON file")
	logLevel := flag.String("logLevel", "info", "The log level to use")
	webAppDir := flag.String("web", "fc-create-election/dist", "The path where the static web app is located")

	// Parse the command line flags
	flag.Parse()
	log.Init(*logLevel, "stdout", nil)

	// check the server URL is http or https and extract the domain
	if !strings.HasPrefix(*server, "http://") && !strings.HasPrefix(*server, "https://") {
		log.Fatal("server URL must start with http:// or https://")
	}
	serverURL = *server
	domain := strings.Split(serverURL, "/")[2]
	log.Infow("server URL", "URL", serverURL, "domain", domain)

	// Create or load the census
	censusInfo := &CensusInfo{}
	if *censusFromFile != "" {
		if err := censusInfo.FromFile(*censusFromFile); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("censusFromFile is required")
	}

	// Create the Vocdoni handler
	handler, err := NewVocdoniHandler(*apiEndpoint, *vocdoniPrivKey, censusInfo)
	if err != nil {
		log.Fatal(err)
	}

	// Create the HTTP API router
	router := new(httprouter.HTTProuter)
	if *tlsDomain {
		router.TLSdomain = domain
	}
	router.TLSdirCert = *tlsDirCert
	if err := router.Init(*host, *port); err != nil {
		log.Fatal(err)
	}

	// Add handler to serve the static files
	router.AddRawHTTPHandler("/app/*", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		log.Infow("serving static file", "path", r.URL.Path)
		http.ServeFile(w, r, *webAppDir)
	})
	router.AddRawHTTPHandler("/app", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		log.Infow("serving static file", "path", r.URL.Path)
		http.ServeFile(w, r, *webAppDir)
	})

	// Create the API handler
	uAPI, err := urlapi.NewAPI(router, "/", *dataDir, db.TypePebble)
	if err != nil {
		log.Fatal(err)
	}

	// The root endpoint redirects to /app
	if err := uAPI.Endpoint.RegisterMethod("/", http.MethodGet, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		ctx.Writer.Header().Add("Location", "/app")
		return ctx.Send([]byte("Redirecting to /app"), http.StatusTemporaryRedirect)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/", http.MethodPost, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		ctx.Writer.Header().Add("Location", "/app")
		return ctx.Send([]byte("Redirecting to /app"), http.StatusTemporaryRedirect)
	}); err != nil {
		log.Fatal(err)
	}

	// Register the API methods
	if err := uAPI.Endpoint.RegisterMethod("/router/{electionID}", http.MethodPost, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		electionID := ctx.URLParam("electionID")
		packet := &FrameSignaturePacket{}
		if err := json.Unmarshal(msg.Data, packet); err != nil {
			return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
		}
		redirectURL := ""
		switch packet.UntrustedData.ButtonIndex {
		case 1:
			redirectURL = fmt.Sprintf(serverURL+"/poll/results/%s", electionID)
		case 2:
			redirectURL = fmt.Sprintf(serverURL+"/poll/%s", electionID)
		default:
			redirectURL = serverURL + "/"
		}
		log.Infow("received router request", "electionID", electionID, "buttonIndex", packet.UntrustedData.ButtonIndex, "redirectURL", redirectURL)
		ctx.Writer.Header().Add("Location", redirectURL)
		return ctx.Send([]byte(redirectURL), http.StatusTemporaryRedirect)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll/results/{electionID}", http.MethodGet, "public", handler.results); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll/results/{electionID}", http.MethodPost, "public", handler.results); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll/{electionID}", http.MethodGet, "public", handler.showElection); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll/{electionID}", http.MethodPost, "public", handler.showElection); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/vote/{electionID}", http.MethodPost, "public", handler.vote); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/create", http.MethodPost, "public", handler.createElection); err != nil {
		log.Fatal(err)
	}

	// close if interrupt received
	log.Infof("startup complete at %s", time.Now().Format(time.RFC850))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Warnf("received SIGTERM, exiting at %s", time.Now().Format(time.RFC850))
}
