package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/flag"

	urlapi "go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"google.golang.org/protobuf/proto"
	"lukechampine.com/blake3"

	farcasterpb "github.com/vocdoni/farcaster-poc/proto"
)

const (
	frameHashSize = 20
)

var frameHTML1 = `
<html lang="en">
      <head>
        <meta property="fc:frame" content="vNext" />
        <meta property="fc:frame:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmSazhXeuFEPSnosTFFS8Yi6KfV1tjUCPxDLSnFY2wVZEJ" />
        <meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/poll" />
        <meta property="fc:frame:button:1" content="Vote" />
        <title>Vocdoni Frame</title>
      </head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
</html>
`

var frameHTML2 = `
<html lang="en">
      <head>
        <meta property="fc:frame" content="vNext" />
        <meta property="fc:frame:image" content="https://app.vocdoni.io/assets/banner.png" />
        <meta property="fc:frame:post_url" content="https://celoni.vocdoni.net/vote" />
        <meta property="fc:frame:button:1" content="Green" />
        <meta property="fc:frame:button:2" content="Purple" />
        <meta property="fc:frame:button:3" content="Red" />
        <meta property="fc:frame:button:4" content="Blue" />
        <title>Vocdoni Frame</title>
      </head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
</html>
`

var frameResponse = `
    <!DOCTYPE html>
    <html>
      <head>
				<meta property="fc:frame" content="vNext" />
				<meta property="fc:frame:image" content="https://black-glamorous-rabbit-362.mypinata.cloud/ipfs/QmVyhAuvdLQgWZ7xog2WtXP88B7TswChCqZdKxVUR5rDUq" />
      </head>
      <body>
        <h1>Hello Farcaster! this is <a href="https://vocdoni.io">Vocdoni</a></h1>
      </body>
    </html>
`

// FrameSignaturePacket mirrors the JSON structure received by the Frame server.
type FrameSignaturePacket struct {
	UntrustedData struct {
		FID         int64  `json:"fid"`
		URL         string `json:"url"`
		MessageHash string `json:"messageHash"`
		Timestamp   int64  `json:"timestamp"`
		Network     int    `json:"network"`
		ButtonIndex int    `json:"buttonIndex"`
		InputText   string `json:"inputText"`
		CastID      struct {
			FID  int64  `json:"fid"`
			Hash string `json:"hash"`
		} `json:"castId"`
	} `json:"untrustedData"`
	TrustedData struct {
		MessageBytes string `json:"messageBytes"`
	} `json:"trustedData"`
}

// VerifyFrameSignature validates the frame message and returns de deserialized frame action and public key.
func VerifyFrameSignature(packet *FrameSignaturePacket) (*farcasterpb.FrameActionBody, ed25519.PublicKey, error) {
	// Decode the message bytes from hex
	messageBytes, err := hex.DecodeString(packet.TrustedData.MessageBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode message bytes: %w", err)
	}

	msg := farcasterpb.Message{}
	if err := proto.Unmarshal(messageBytes, &msg); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal Message: %w", err)
	}
	log.Debugf("farcaster signed message: %s", log.FormatProto(&msg))

	if msg.Data == nil {
		return nil, nil, fmt.Errorf("invalid message data")
	}
	if msg.SignatureScheme != farcasterpb.SignatureScheme_SIGNATURE_SCHEME_ED25519 {
		return nil, nil, fmt.Errorf("invalid signature scheme")
	}
	if msg.Data.Type != farcasterpb.MessageType_MESSAGE_TYPE_FRAME_ACTION {
		return nil, nil, fmt.Errorf("invalid message type, got %s", msg.Data.Type.String())
	}
	var pubkey ed25519.PublicKey = msg.GetSigner()

	if pubkey == nil {
		return nil, nil, fmt.Errorf("signer is nil")
	}

	// Verify the hash and signature
	msgDataBytes, err := proto.Marshal(msg.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal message data: %w", err)
	}

	log.Debugw("verifying message signature", "size", len(msgDataBytes))
	h := blake3.New(160, nil)
	if _, err := h.Write(msgDataBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to hash message: %w", err)
	}
	hashed := h.Sum(nil)[:frameHashSize]

	if !bytes.Equal(msg.Hash, hashed) {
		return nil, nil, fmt.Errorf("hash mismatch (got %x, expected %x)", hashed, msg.Hash)
	}

	if !ed25519.Verify(pubkey, hashed, msg.GetSignature()) {
		return nil, nil, fmt.Errorf("signature verification failed")
	}
	actionBody := msg.Data.GetFrameActionBody()
	if actionBody == nil {
		return nil, nil, fmt.Errorf("invalid action body")
	}

	return actionBody, pubkey, nil
}

func main() {
	tlsDomain := flag.String("tlsDomain", "", "The domain to use for the TLS certificate")
	tlsDirCert := flag.String("tlsDirCert", "", "The directory to use for the TLS certificate")
	host := flag.String("listenHost", "", "The host to listen on")
	port := flag.Int("listenPort", 0, "The port to listen on")
	dataDir := flag.String("dataDir", "", "The directory to use for the data")
	// Parse the command line flags
	flag.Parse()
	log.Init("debug", "stdout", nil)

	// Simulating receiving a frame signature packet
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

	if err := uAPI.Endpoint.RegisterMethod("/", http.MethodGet, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(frameHTML1), http.StatusOK)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/poll", http.MethodPost, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(frameHTML2), http.StatusOK)
	}); err != nil {
		log.Fatal(err)
	}

	if err := uAPI.Endpoint.RegisterMethod("/vote", http.MethodPost, "public", func(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
		packet := &FrameSignaturePacket{}
		if err := json.Unmarshal(msg.Data, packet); err != nil {
			return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
		}
		fmt.Printf("Received frame signature packet:\n%+v\n", packet)
		action, pubkey, err := VerifyFrameSignature(packet)
		if err != nil {
			return fmt.Errorf("failed to verify frame signature: %w", err)
		}
		log.Infow("successfully verified frame signature", "pubkey", pubkey, "action", log.FormatProto(action))

		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(frameResponse), http.StatusOK)
	}); err != nil {
		log.Fatal(err)
	}

	log.Infof("startup complete at %s", time.Now().Format(time.RFC850))

	// close if interrupt received
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Warnf("received SIGTERM, exiting at %s", time.Now().Format(time.RFC850))
}
