package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"lukechampine.com/blake3"
)

func (v *vocdoniHandler) imagesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	id := ctx.URLParam("id")
	data := v.imageFromCache(id)
	if data != nil {
		return imageResponse(ctx, data)
	}
	idSplit := strings.Split(id, "_")
	if len(idSplit) != 2 {
		log.Debugw("access to old PNG", "requestURI", ctx.Request.RequestURI, "url", ctx.Request.URL, "headers", ctx.Request.Header)
		return errorImageResponse(ctx, fmt.Errorf("nothing here... click results"))
	}
	electionID, err := hex.DecodeString(idSplit[0])
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to decode id: %w", err))
	}
	// check if the election is finished and if so, send the final results as a static PNG
	if v.checkIfElectionFinishedAndHandle(electionID, ctx) {
		return nil
	}
	// as fallback, get the election and return the landing png
	election, err := v.election(electionID)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to get election: %w", err))
	}
	png, err := buildLandingPNG(election)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to build landing: %w", err))
	}
	return imageResponse(ctx, png)
}

// addImageToCache adds an image to the LRU cache.
// Returns the full URL (absolute) of the image.
// If electionID is nil, the image is not associated with any election.
func (v *vocdoniHandler) addImageToCache(data []byte, electionID types.HexBytes) string {
	id := blake3.Sum256(data)
	var idstr string
	if electionID == nil {
		idstr = hex.EncodeToString(id[:])
	} else {
		idstr = fmt.Sprintf("%s_%s", electionID.String(), hex.EncodeToString(id[:8]))
	}
	v.imagesLRU.Add(idstr, data)
	return fmt.Sprintf("%s%s/%s.png", serverURL, imageHandlerPath, idstr)
}

// imageFromCache gets an image from the LRU cache.
// Returns nil if the image is not in the cache.
func (v *vocdoniHandler) imageFromCache(id string) []byte {
	data, ok := v.imagesLRU.Get(id)
	if !ok {
		return nil
	}
	return data
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

	png, err := textToImage(electionImageContents(election), frames[BackgroundGeneric])
	if err != nil {
		return errorImageResponse(ctx, err)
	}

	// set png headers and return response as is
	return imageResponse(ctx, png)
}

func imageResponse(ctx *httprouter.HTTPContext, png []byte) error {
	log.Debugw("sending image response", "size", len(png))
	defer ctx.Request.Body.Close()
	if ctx.Request.Context().Err() != nil {
		// The connection was closed, so don't try to write to it.
		return fmt.Errorf("connection is closed")
	}
	ctx.SetResponseContentType("image/png")
	return ctx.Send(png, 200)
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
		response := strings.ReplaceAll(frame(testImageHTML), "{image}", v.addImageToCache(png, nil))
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
	jresponse, err := json.Marshal(map[string]string{"image": v.addImageToCache(png, nil)})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send(jresponse, http.StatusOK)
}
