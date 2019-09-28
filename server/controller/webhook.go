package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

var colors = map[string]string{
	"test":    "#AAAAAA",
	"success": "#41aa58",
	"failure": "#d10c20",
}

var texts = map[string]string{
	"test":    "Hello from CircleCI",
	"success": "CircleCI build success",
	"failure": "CircleCI build failure",
}

var submitSurveyResponse = &Endpoint{
	Path:         "/submit",
	Method:       http.MethodPost,
	Execute:      executeSubmitSurveyResponse,
	RequiresAuth: false,
}

type APIResponse struct {
	ChannelID string `json:"channel_id"`
	Status    string `json:"status"`
	BuildNum  string `json:"build_num"`
	RepoName  string `json:"repo_name"`
	Tag       string `json:"tag"`
	Commit    string `json:"commit"`
	BuildURL  string `json:"build_url"`
}

func executeSubmitSurveyResponse(w http.ResponseWriter, r *http.Request) error {
	var response APIResponse
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return errors.Wrap(err, "failed to decode request body into survey response object")
	}

	fmt.Println(response)

	var post *model.Post
	if response.Status == "failure" {
		post = generateFailurePost(response)
	} else if response.Status == "success" {
		post = generateSuccessPost(response)
	}

	post, appErr := config.Mattermost.CreatePost(post)
	if appErr != nil {
		config.Mattermost.LogError(appErr.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(model.MapToJson(map[string]string{"status": "ok"}))); err != nil {
		return errors.Wrap(err, "failed to write data to HTTP response")
	}

	return nil
}

func generateFailurePost(response APIResponse) *model.Post {
	conf := config.GetConfig()
	slackAttachmentFields := []*model.SlackAttachmentField{
		{
			Title: "Link",
			Value: response.BuildURL,
			Short: false,
		},
		{
			Title: "Tag",
			Value: response.Tag,
			Short: true,
		},
		{
			Title: "Commit",
			Value: response.Commit,
			Short: true,
		},
	}

	attachment := &model.SlackAttachment{
		Color:    "#d10c20",
		Title:    fmt.Sprintf("Oops. Build %s failed.", response.BuildNum),
		Text:     response.RepoName,
		Fields:   slackAttachmentFields,
		ThumbURL: "https://png.pngtree.com/svg/20170406/icon_failed__1325447.png",
	}

	post := &model.Post{
		UserId:    conf.BotUserID,
		ChannelId: response.ChannelID,
		Props: model.StringInterface{
			"from_webhook":      "true",
			"override_icon_url": "https://circleci.zendesk.com/system/brands/0011/9868/circleci-1_thumb.png",
			"attachments":       []*model.SlackAttachment{attachment},
		},
	}

	return post
}

func generateSuccessPost(response APIResponse) *model.Post {
	conf := config.GetConfig()
	slackAttachmentFields := []*model.SlackAttachmentField{
		{
			Title: "Link",
			Value: response.BuildURL,
			Short: false,
		},
		{
			Title: "Tag",
			Value: response.Tag,
			Short: true,
		},
		{
			Title: "Commit",
			Value: response.Commit,
			Short: true,
		},
	}

	attachment := &model.SlackAttachment{
		Color:    "#41aa58",
		Title:    fmt.Sprintf("Build %s passed.", response.BuildNum),
		Text:     response.RepoName,
		Fields:   slackAttachmentFields,
		ThumbURL: "https://png.pngtree.com/svg/20170510/success_404253.png",
	}

	post := &model.Post{
		UserId:    conf.BotUserID,
		ChannelId: response.ChannelID,
		Props: model.StringInterface{
			"from_webhook":      "true",
			"override_icon_url": "https://circleci.zendesk.com/system/brands/0011/9868/circleci-1_thumb.png",
			"attachments":       []*model.SlackAttachment{attachment},
		},
	}

	return post
}
