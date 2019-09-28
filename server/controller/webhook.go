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

func executeSubmitSurveyResponse(w http.ResponseWriter, r *http.Request) error {
	var response struct {
		ChannelID string `json:"channel_id"`
		Status    string `json:"status"`
		Build     string `json:"build"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&response); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return errors.Wrap(err, "failed to decode request body into survey response object")
	}

	fmt.Println(response)
	conf := config.GetConfig()

	// slackAttachmentFields := []*model.SlackAttachmentField{
	// 	{
	// 		Title: "",
	// 		Value: "",
	// 		Short: false,
	// 	},
	// 	{
	// 		Title: "",
	// 		Value: "",
	// 		Short: false,
	// 	},
	// 	{
	// 		Title: "",
	// 		Value: "",
	// 		Short: true,
	// 	},
	// 	{
	// 		Title: "",
	// 		Value: "",
	// 		Short: true,
	// 	},
	// }

	attachment := &model.SlackAttachment{
		Color: colors[response.Status],
		Text:  texts[response.Status],
		// Fields:     slackAttachmentFields,
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
