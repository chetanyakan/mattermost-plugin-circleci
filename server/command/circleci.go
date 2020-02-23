package command

import (
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

const (
	invalidCommand = "Invalid command parameters. Please use `/circleci help` for more information."
)

var CircleCICommandHandler = Handler{
	Command: &model.Command{
		Trigger:          "circleci",
		Description:      "Integration with CircleCI.",
		DisplayName:      "CircleCI",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect <token>, disconnect, me.",
		AutoCompleteHint: "[command]",
		Username:         config.BotUserName,
		IconURL:          config.BotIconURL,
	},
	handlers: map[string]HandlerFunc{
		"connect":    executeConnect,
		"disconnect": executeDisconnect,
		"me":         executeMe,
	},
	defaultHandler: func(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
		return util.SendEphemeralCommandResponse(invalidCommand)
	},
}
