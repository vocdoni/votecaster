package main

import (
	"encoding/json"
	"net/http"

	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
	"go.vocdoni.io/dvote/log"
)

func (v *vocdoniHandler) channelsHandler(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
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
