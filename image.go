package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/vocdoni/vote-frame/imageframe"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/types"
)

func (v *vocdoniHandler) imagesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	id := ctx.URLParam("id")
	data := imageframe.FromCache(id)
	if data != nil {
		return imageResponse(ctx, data)
	}
	idSplit := strings.Split(id, "_")
	var electionID types.HexBytes
	var err error

	electionID, err = hex.DecodeString(idSplit[0])
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to decode id: %w", err))
	}
	// check if the election is finished and if so, send the final results as a static PNG
	pngResults := v.db.FinalResultsPNG(electionID)
	if pngResults != nil {
		// for future requests, add the image to the cache with the given id
		imageframe.AddImageToCacheWithID(id, pngResults)
		return imageResponse(ctx, pngResults)
	}

	// as fallback, get the election and return the landing png
	election, err := v.election(electionID)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to get election: %w", err))
	}
	png, err := imageframe.QuestionImage(election)
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to build landing: %w", err))
	}
	return imageResponse(ctx, imageframe.FromCache(png))
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

	png, err := imageframe.QuestionImage(election)
	if err != nil {
		return errorImageResponse(ctx, err)
	}

	// set png headers and return response as is
	return imageResponse(ctx, imageframe.FromCache(png))
}

func imageResponse(ctx *httprouter.HTTPContext, png []byte) error {
	defer ctx.Request.Body.Close()
	if ctx.Request.Context().Err() != nil {
		// The connection was closed, so don't try to write to it.
		return fmt.Errorf("connection is closed")
	}
	ctx.SetResponseContentType("image/png")
	return ctx.Send(png, 200)
}

func errorImageResponse(ctx *httprouter.HTTPContext, err error) error {
	png, err := imageframe.ErrorImage(err.Error())
	if err != nil {
		return err
	}
	return imageResponse(ctx, imageframe.FromCache(png))
}

// imageLink returns the URL for the image with the given key.
func imageLink(imageKey string) string {
	return fmt.Sprintf("%s/images/%s.png", serverURL, imageKey)
}
