package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/circleci"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

func commandMe() *Config {
	return &Config{
		Command: &model.Command{
			Trigger:          "me",
			AutoComplete:     false,
			AutoCompleteDesc: "Details about connected user.",
		},
		HelpText: "",
		Validate: validateMe,
		Execute:  executeMe,
	}
}

func validateMe(args []string, context Context) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.CommandArgs.UserId + "_auth_token")
	if appErr != nil {
		return util.SendEphemeralText(appErr.Error())
	}
	if string(authToken) == "" {
		return util.SendEphemeralText("Not connected. Please connect and try again later.")
	}
	return nil, nil
}

func executeMe(args []string, context Context) (*model.CommandResponse, *model.AppError) {
	authToken, _ := config.Mattermost.KVGet(context.CommandArgs.UserId + "_auth_token")
	client := &circleci.Client{Token: string(authToken)}
	user, err := client.Me()
	if err != nil {
		return util.SendEphemeralText("Unable to connect to circleci. Make sure the auth token is still valid. " + err.Error())
	}

	attachment := &model.SlackAttachment{
		Color:    "#7FC1EE",
		Pretext:  fmt.Sprintf("Initiated by CircleCI user: %s", user.Login),
		ThumbURL: user.AvatarURL,
		Fields: []*model.SlackAttachmentField{
			{
				Title: "Name",
				Value: user.Name,
				Short: true,
			},
			{
				Title: "Email",
				Value: user.SelectedEmail,
				Short: true,
			},
		},
	}

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     "https://circleci.zendesk.com/system/brands/0011/9868/circleci-1_thumb.png",
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}
