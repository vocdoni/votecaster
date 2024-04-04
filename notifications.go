package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vocdoni/vote-frame/imageframe"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
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
		png := imageframe.NotificationsImage()
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
