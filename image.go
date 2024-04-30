package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/types"
)

var imageTypeRgx = regexp.MustCompile(`^data:(image/[a-z]+);base64,`)

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

// avatarHandler returns the avatar image with the given avatarID.
func (v *vocdoniHandler) avatarHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	avatarID := ctx.URLParam("avatarID")
	if avatarID == "" {
		return ctx.Send([]byte{}, 400)
	}
	avatar, err := v.db.Avatar(avatarID)
	if err != nil {
		if err == mongo.ErrAvatarUnknown {
			return ctx.Send([]byte{}, 404)
		}
		return fmt.Errorf("failed to get avatar: %w", err)
	}
	// return the avatar image as is
	ctx.SetResponseContentType(avatar.ContentType)
	// decode base64 image and return it
	png, err := base64.StdEncoding.DecodeString(string(avatar.Data))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}
	return ctx.Send(png, 200)
}

// uploadAvatarHandler uploads the avatar image with the given avatarID,
// associated to the user that is making the request and the communityID
// provided.
func (v *vocdoniHandler) updloadAvatarHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	// extract userFID from auth token
	userFID, err := v.db.UserFromAuthToken(msg.AuthToken)
	if err != nil {
		return fmt.Errorf("cannot get user from auth token: %w", err)
	}
	req := struct {
		AvatarID    string `json:"id"`
		CommunityID uint64 `json:"communityID"`
		Data        string `json:"data"`
	}{}
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return fmt.Errorf("cannot parse request: %w", err)
	}
	imageTypeResults := imageTypeRgx.FindStringSubmatch(req.Data)
	if len(imageTypeResults) != 2 {
		return ctx.Send([]byte("bad formatted image"), 400)
	}
	prefix, contentType := imageTypeResults[0], imageTypeResults[1]
	// remove the prefix from the image data
	req.Data, _ = strings.CutPrefix(req.Data, prefix)
	if err := v.db.SetAvatar(req.AvatarID, []byte(req.Data), userFID, req.CommunityID, contentType); err != nil {
		return fmt.Errorf("cannot set avatar: %w", err)
	}
	return ctx.Send([]byte{}, 200)
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
