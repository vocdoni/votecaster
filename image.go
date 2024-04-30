package main

import (
	"crypto/md5"
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

var (
	imageTypeRgx   = regexp.MustCompile(`^data:(image/[a-z]+);base64,`)
	isAvatarURLRgx = regexp.MustCompile(`.+/images/avatar/(.+)\.jpg$`)
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
	// upload the avatar and return the URL
	avatarURL, err := v.uploadAvatar(req.AvatarID, userFID, req.CommunityID, req.Data)
	if err != nil {
		return fmt.Errorf("cannot upload avatar: %w", err)
	}
	// return the URL of the uploaded avatar
	result := map[string]string{"logoURL": avatarURL}
	res, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("cannot marshal result: %w", err)
	}
	return ctx.Send(res, 200)
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

// uploadAvatar uploads the avatar image with the given avatarID, associated to
// the user with the given userFID and the community with the given communityID.
// If the avatarID is empty, it calculates the avatarID from the data. It returns
// the URL of the uploaded avatar image. It stores the avatar in the database.
// If an error occurs, it returns an empty string and the error.
func (v *vocdoniHandler) uploadAvatar(avatarID string, userFID, communityID uint64, data string) (string, error) {
	if !isBase64Image(data) {
		return "", fmt.Errorf("image is not base64 encoded")
	}
	var err error
	if avatarID == "" {
		if avatarID, err = calculateAvatarID([]byte(data)); err != nil {
			return "", fmt.Errorf("error calculating avatarID: %w", err)
		}
	}
	imageTypeResults := imageTypeRgx.FindStringSubmatch(data)
	if len(imageTypeResults) != 2 {
		return "", fmt.Errorf("bad formatted image")
	}
	prefix, contentType := imageTypeResults[0], imageTypeResults[1]
	// remove the prefix from the image data
	data, _ = strings.CutPrefix(data, prefix)
	// store the avatar in the database
	if err := v.db.SetAvatar(avatarID, []byte(data), userFID, communityID, contentType); err != nil {
		return "", fmt.Errorf("cannot set avatar: %w", err)
	}
	return avatarURL(serverURL, avatarID), nil
}

// avatarURL returns the URL for the avatar with the given avatarID.
func avatarURL(baseURL, avatarID string) string {
	return fmt.Sprintf("%s/images/avatar/%s.jpg", baseURL, avatarID)
}

// avatarIDfromURL returns the avatarID from the given URL. If the URL is not an
// avatar URL, it returns an empty string and false.
func avatarIDfromURL(url string) (string, bool) {
	avatarID := isAvatarURLRgx.FindStringSubmatch(url)
	if len(avatarID) != 2 {
		return "", false
	}
	return avatarID[1], true
}

// isBase64Image returns true if the given data is a base64 encoded image.
func isBase64Image(data string) bool {
	return imageTypeRgx.MatchString(data)
}

// calculateAvatarID calculates the avatarID from the given data. The avatarID
// is the first 12 bytes of the md5 hash of the data. If an error occurs, it
// returns an empty string and the error.
func calculateAvatarID(data []byte) (string, error) {
	md5hash := md5.New()
	if _, err := md5hash.Write(data); err != nil {
		return "", fmt.Errorf("cannot calculate hash: %w", err)
	}
	bhash := md5hash.Sum(nil)[:12]
	return hex.EncodeToString(bhash), nil
}
