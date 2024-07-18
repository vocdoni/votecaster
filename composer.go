package main

import (
	"encoding/json"
	"net/http"

	"go.vocdoni.io/dvote/httprouter"
	"go.vocdoni.io/dvote/httprouter/apirest"
)

type ComposerResponse struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (v *vocdoniHandler) composer(msg *apirest.APIdata, ctx *httprouter.HTTPContext) error {
	response := ComposerResponse{
		Type:  "form",
		Title: "Create a blockchain Poll",
		Url:   serverURL + "/app#/composer",
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	return ctx.Send(data, http.StatusOK)
}
