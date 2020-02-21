package command

import (
	"fmt"

	circleci "github.com/jszwedko/go-circleci"
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

func executeMe(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	client := &circleci.Client{Token: string(authToken)}
	user, err := client.Me()
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. " + err.Error())
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
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}
