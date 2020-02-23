package controller

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/model"

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

	var post *model.Post
	if cwReq.Status == "failure" {
		post = cwReq.GenerateFailurePost()
	} else if cwReq.Status == "success" {
		post = cwReq.GenerateSuccessPost()
	}

	if post != nil {
		if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
			config.Mattermost.LogError(appErr.Error())
		}
	}

	returnStatusOK(w)
}
