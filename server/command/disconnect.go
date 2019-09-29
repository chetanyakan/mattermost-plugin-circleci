package command

import (
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

func commandDisconnect() *Config {
	return &Config{
		Command: &model.Command{
			Trigger:          "disconnect",
			AutoComplete:     false,
			AutoCompleteDesc: "",
			AutoCompleteHint: "",
		},
		HelpText: "",
		Validate: validateDisconnect,
		Execute:  executeDisconnect,
	}
}

func validateDisconnect(args []string, context Context) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.CommandArgs.UserId + "_auth_token")
	if appErr != nil {
		return util.SendEphemeralText(appErr.Error())
	}
	if string(authToken) == "" {
		return util.SendEphemeralText("Not connected. Please connect and try again later.")
	}

	return nil, nil
}

func executeDisconnect(args []string, context Context) (*model.CommandResponse, *model.AppError) {
	if err := config.Mattermost.KVDelete(context.CommandArgs.UserId + "_auth_token"); err != nil {
		return util.SendEphemeralText("Unable to disconnect. Error: " + err.Error())
	}

	return util.SendEphemeralText("Successfully disconnected.")
}
