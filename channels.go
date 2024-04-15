package main

import (
	"encoding/json"
	"net/http"

	"github.com/vocdoni/vote-frame/farcasterapi"
	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

func (v *vocdoniHandler) channelHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	channelID := ctx.URLParam("channelID")
	if channelID == "" {
		return ctx.Send([]byte("no channel id provided"), http.StatusBadRequest)
	}
	ch, err := v.fcapi.Channel(ctx.Request.Context(), channelID)
	if err != nil {
		if err == farcasterapi.ErrChannelNotFound {
			return ctx.Send([]byte("channel not found"), http.StatusNotFound)
		}
		return ctx.Send([]byte(err.Error()), apirest.HTTPstatusInternalErr)
	}
	res, err := json.Marshal(map[string]interface{}{
		"id":            ch.ID,
		"name":          ch.Name,
		"description":   ch.Description,
		"followerCount": ch.Followers,
		"image":         ch.Image,
		"url":           ch.URL,
	})
	if err != nil {
		return ctx.Send([]byte("error encoding channel details"), apirest.HTTPstatusInternalErr)
	}
	return ctx.Send(res, http.StatusOK)
}

func (v *vocdoniHandler) findChannelHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	query := ctx.Request.URL.Query().Get("q")
	if query == "" {
		return ctx.Send([]byte("query parameter not provided"), http.StatusBadRequest)
	}
	channels, err := v.fcapi.FindChannel(ctx.Request.Context(), query)
	if err != nil {
		log.Errorw(err, "failed to list channels")
		return ctx.Send([]byte("error getting list of channels"), http.StatusInternalServerError)
	}
	res := map[string][]map[string]interface{}{"channels": {}}
	for _, ch := range channels {
		res["channels"] = append(res["channels"], map[string]interface{}{
			"id":            ch.ID,
			"name":          ch.Name,
			"description":   ch.Description,
			"followerCount": ch.Followers,
			"image":         ch.Image,
			"url":           ch.URL,
		})
	}
	bRes, err := json.Marshal(res)
	if err != nil {
		log.Errorw(err, "failed to marshal channels")
		return ctx.Send([]byte("error marshaling channels"), http.StatusInternalServerError)
	}
	return ctx.Send(bRes, http.StatusOK)
}
