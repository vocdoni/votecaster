package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vocdoni/vote-frame/imageframe"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/types"
)

func (v *vocdoniHandler) notificationsHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	png := imageframe.NotificationsImage()
	response := strings.ReplaceAll(frame(frameNotifications), "{image}", imageLink(png))

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) notificationsResponseHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}

	actionMessage, m, _, err := VerifyFrameSignature(packet)
	if err != nil {
		return ErrFrameSignature
	}

	if actionMessage.ButtonIndex == 3 {
		// User has clicked the "filter by user" button
		png := imageframe.NotificationsManageImage()
		response := strings.ReplaceAll(frame(frameNotificationsManager), "{image}", imageLink(png))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return ctx.Send([]byte(response), http.StatusOK)
	}

	allowNotifications := actionMessage.ButtonIndex == 1

	var png string
	if err := v.db.SetNotificationsAcceptedForUser(m.Data.Fid, allowNotifications); err != nil {
		return fmt.Errorf("failed to update notifications: %w", err)
	}
	log.Infow("notifications updated", "fid", m.Data.Fid, "allow", allowNotifications)
	if allowNotifications {
		png = imageframe.NotificationsAcceptedImage()
	} else {
		png = imageframe.NotificationsDeniedImage()
	}
	response := strings.ReplaceAll(frame(frameNotificationsResponse), "{image}", imageLink(png))

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

func (v *vocdoniHandler) notificationsFilterByUserHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	packet := &FrameSignaturePacket{}
	if err := json.Unmarshal(msg.Data, packet); err != nil {
		return fmt.Errorf("failed to unmarshal frame signature packet: %w", err)
	}

	actionMessage, m, _, err := VerifyFrameSignature(packet)
	if err != nil {
		return ErrFrameSignature
	}

	setErrorResponse := func(err error) []byte {
		log.Warnw("failed to filter by user", "error", err)
		png := imageframe.NotificationsErrorImage()
		response := strings.ReplaceAll(frame(frameNotificationsResponse), "{image}", imageLink(png))
		ctx.SetResponseContentType("text/html; charset=utf-8")
		return []byte(response)
	}

	if len(actionMessage.InputText) == 0 {
		return ctx.Send(setErrorResponse(fmt.Errorf("missing input text")), http.StatusOK)
	}

	// Get the filtered user by username from the database
	user, err := v.db.UserByUsername(string(actionMessage.InputText))
	if err != nil {
		return ctx.Send(setErrorResponse(err), http.StatusOK)
	}

	allowNotifications := actionMessage.ButtonIndex == 1

	if !allowNotifications {
		if err := v.db.AddNotificationMutedUser(m.Data.Fid, user.UserID); err != nil {
			return ctx.Send(setErrorResponse(err), http.StatusOK)
		}
	} else {
		if err := v.db.DelNotificationMutedUser(m.Data.Fid, user.UserID); err != nil {
			return ctx.Send(setErrorResponse(err), http.StatusOK)
		}
	}

	var png string
	log.Infow("notifications updated by user", "fid", m.Data.Fid, "filtered user", user.Username, "allow", allowNotifications)
	if allowNotifications {
		png = imageframe.NotificationsAcceptedImage()
	} else {
		png = imageframe.NotificationsDeniedImage()
	}
	response := strings.ReplaceAll(frame(frameNotificationsResponse), "{image}", imageLink(png))

	ctx.SetResponseContentType("text/html; charset=utf-8")
	return ctx.Send([]byte(response), http.StatusOK)
}

// notificationsForceSendHandler enqueue the notifications for the given election and its users.
// The election should be already created and the census stored in the database.
// Returns the list of usernames that will receive the notifications.
func (v *vocdoniHandler) notificationsSendHandler(_ *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	electionID, err := hex.DecodeString(ctx.URLParam("electionID"))
	if err != nil {
		return fmt.Errorf("failed to decode election ID: %w", err)
	}

	election, err := v.cli.Election(electionID)
	if err != nil {
		return fmt.Errorf("failed to get election: %w", err)
	}

	census, err := v.db.CensusFromElection(electionID)
	if err != nil {
		return fmt.Errorf("failed to get census: %w", err)
	}

	userProfile, err := v.db.User(census.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	expiration := election.EndDate
	if time.Until(expiration) < time.Hour*3 {
		expiration = expiration.Add(-time.Minute * 10)
	} else {
		expiration = expiration.Add(-time.Hour * 3)
	}

	usernames := []string{}
	for participant := range census.Participants {
		usernames = append(usernames, participant)
	}

	if err := v.createNotifications(
		electionID,
		census.CreatedBy,
		userProfile.Username,
		usernames,
		fmt.Sprintf("%s/%x", serverURL, electionID),
		"",
		expiration,
	); err != nil {
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	data, err := json.Marshal(usernames)
	if err != nil {
		return fmt.Errorf("failed to marshal usernames: %w", err)
	}
	return ctx.Send(data, http.StatusOK)
}

// createNotifications creates a notification for each user in the census.
func (v *vocdoniHandler) createNotifications(electionID types.HexBytes, ownerFID uint64,
	ownerName string, usernames []string, frameURL, customText string, deadline time.Time,
) error {
	log.Infow("enqueue notifications",
		"owner", ownerName,
		"electionID", electionID.String(),
		"userCount", len(usernames),
		"frameURL", frameURL,
	)
	// Get the election from the database to get the community ID and name
	election, err := v.db.Election(electionID)
	if err != nil {
		return fmt.Errorf("failed to get election: %w", err)
	}
	if election.Community == nil {
		return fmt.Errorf("election has no community")
	}
	// Add a notification for each user in the census to the database
	for _, username := range usernames {
		user, err := v.db.UserByUsername(username)
		if err != nil {
			if errors.Is(err, mongo.ErrUserUnknown) {
				log.Warnw("user not found", "username", username)
				continue
			}
			return fmt.Errorf("failed to get user: %w", err)
		}
		if _, err := v.db.AddNotifications(mongo.NotificationTypeNewElection, electionID.String(),
			user.UserID, ownerFID, election.Community.ID, username, ownerName, election.Community.Name,
			frameURL, customText, deadline,
		); err != nil {
			return fmt.Errorf("failed to add notification: %w", err)
		}
	}
	return nil
}

func (v *vocdoniHandler) directNotificationsHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	token := msg.AuthToken
	if token == "" {
		return fmt.Errorf("missing auth token header")
	}
	auth, err := v.db.UpdateActivityAndGetData(token)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// get access profile to use the warpcast api key of the current user
	accessProfile, err := v.db.UserAccessProfile(auth.UserID)
	if err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// check if the user has a configured warpcast api key
	if accessProfile == nil || accessProfile.WarpcastAPIKey == "" {
		return ctx.Send([]byte("no warpcast api key configured"), http.StatusBadRequest)
	}
	// decode the request body
	var directNotification DirectNotification
	if err := json.Unmarshal(msg.Data, &directNotification); err != nil {
		return ctx.Send([]byte(err.Error()), http.StatusInternalServerError)
	}
	// get users to notify checking if they have already voted
	toNotify := []uint64{}
	electionID := types.HexStringToHexBytes(directNotification.ElectionID)
	for _, userFID := range directNotification.FIDs {
		alreadyVoted, err := v.db.HasAlreadyVoted(userFID, electionID)
		if err != nil {
			log.Warnf("failed to check if user has already voted: %v", err)
			continue
		}
		if !alreadyVoted {
			toNotify = append(toNotify, userFID)
		}
	}
	internalCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// send the direct notifications
	failed := 0
	for _, userFID := range toNotify {
		if err := v.fcapi.DirectMessage(internalCtx, accessProfile.WarpcastAPIKey, directNotification.Content, userFID); err != nil {
			log.Warnf("failed to send direct notification: %v", err)
			failed++
		}
	}
	if failed > 0 {
		return ctx.Send([]byte(fmt.Sprintf("failed to send %d direct notifications", failed)), http.StatusInternalServerError)
	}
	return ctx.Send([]byte("ok"), http.StatusOK)
}
