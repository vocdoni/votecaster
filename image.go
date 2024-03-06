package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/vocdoni/vote-frame/imageframe"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
	"lukechampine.com/blake3"
)

var oldImagesHandlerMap = map[string]string{
	"08d56a13ee330b927b7d181a3ee17de580e265ca51ac5b0728c2a272712585bd": "4ae20a8eb4caa52f5588f7bb9f3c6d6b7cf003a5b03f4589edea1000000000a2",
}

func (v *vocdoniHandler) imagesHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	id := ctx.URLParam("id")
	data := v.imageFromCache(id)
	if data != nil {
		return imageResponse(ctx, data)
	}
	idSplit := strings.Split(id, "_")
	var electionID types.HexBytes
	var err error
	if len(idSplit) != 2 {
		// for backwards compatibility, check if the id is in the oldImagesHandlerMap
		// remove this code after some weeks
		if eid, ok := oldImagesHandlerMap[id]; ok {
			log.Warnw("old PNG match", "id", id)
			electionID, err = hex.DecodeString(eid)
			if err != nil {
				return errorImageResponse(ctx, fmt.Errorf("nothing here... click results"))
			}
			election, err := v.election(electionID)
			if err != nil {
				return errorImageResponse(ctx, fmt.Errorf("id not found... click results"))
			}
			png, err := buildResultsPNG(election)
			if err != nil {
				return errorImageResponse(ctx, fmt.Errorf("failed to build results: %w", err))
			}
			return imageResponse(ctx, png)
		} else {
			log.Debugw("access to old PNG", "requestURI", ctx.Request.RequestURI, "url", ctx.Request.URL)
			return errorImageResponse(ctx, fmt.Errorf("nothing here... click results"))
		}
	}

	electionID, err = hex.DecodeString(idSplit[0])
	if err != nil {
		return errorImageResponse(ctx, fmt.Errorf("failed to decode id: %w", err))
	}
	// check if the election is finished and if so, send the final results as a static PNG
	pngResults := v.db.FinalResultsPNG(electionID)
	if pngResults != nil {
		return imageResponse(ctx, pngResults)
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

	png, err := buildLandingPNG(election)
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
	png, err := imageframe.ErrorImage(err.Error())
	if err != nil {
		return err
	}
	return imageResponse(ctx, png)
}
