package controller

import (
	"encoding/json"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/service"
	"net/http"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
)

var circleCIBuildFinished = &Endpoint{
	Path:         "/webhook",
	Method:       http.MethodPost,
	Execute:      handleCircleCIBuildFinished,
	RequiresAuth: false,
}

func handleCircleCIBuildFinished(w http.ResponseWriter, r *http.Request) {
	var cwReq serializer.CircleCIWebhookRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&cwReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		config.Mattermost.LogError("Failed to decode request body.", "Error", err.Error())
		return
	}

	if err := service.SendWebhookNotifications(cwReq); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	returnStatusOK(w)
}
