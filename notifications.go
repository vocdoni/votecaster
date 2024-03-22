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
