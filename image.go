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
	"lukechampine.com/blake3"
)

func (v *vocdoniHandler) imagesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	log.Debugw("received images request", "method", ctx.Request.Method, "url", ctx.Request.URL)
	id := ctx.URLParam("id")
	data := v.imageFromCache(id)
	log.Debugw("fetched image from cache", "id", id, "datasize", len(data))
	if data == nil {
		log.Debugw("image not found in cache", "id", id)
		return ctx.Send(nil, http.StatusNotFound)
	}
	log.Debugw("image found, calling imageResponse", "id", id, "datasize", len(data))
	return imageResponse(ctx, data)
}

// addImageToCache adds an image to the LRU cache.
// Returns the full URL (absolute) of the image.
func (v *vocdoniHandler) addImageToCache(data []byte) string {
	log.Debugw("adding image to cache", "data", len(data))

	id := blake3.Sum256(data)
	idstr := hex.EncodeToString(id[:])
	// check if the image is already in the cache and return it
	if v.imageFromCache(idstr) != nil {
		return fmt.Sprintf("%s%s/%s.png", serverURL, imageHandlerPath, idstr)
	}
	evicted := v.imagesLRU.Add(idstr, data)
	log.Debugw("added image to cache", "id", idstr, "evicted", evicted)
	return fmt.Sprintf("%s%s/%s.png", serverURL, imageHandlerPath, idstr)
}

// imageFromCache gets an image from the LRU cache.
// Returns nil if the image is not in the cache.
func (v *vocdoniHandler) imageFromCache(id string) []byte {
	log.Debugw("getting image from cache", "id", id)
	data, ok := v.imagesLRU.Get(id)
	if !ok {
		log.Debugw("image not found in cache", "id", id)
		return nil
	}
	log.Debugw("image found in cache", "id", id)
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

	png, err := textToImage(electionImageContents(election), backgrounds[BackgroundGeneric])
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
	ctx.Writer.Header().Set("Content-Type", "image/png")
	ctx.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(png)))
	log.Debugw("writing image response", "size", len(png))
	_, err := ctx.Writer.Write(png)
	log.Debugw("wrote image response", "size", len(png), "error", err)
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
		response := strings.ReplaceAll(frame(testImageHTML), "{image}", v.addImageToCache(png))
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
	jresponse, err := json.Marshal(map[string]string{"image": v.addImageToCache(png)})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	ctx.Writer.Header().Set("Content-Type", "application/json")
	return ctx.Send(jresponse, http.StatusOK)
}
