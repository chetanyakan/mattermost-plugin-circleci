package command

import (
	circleci "github.com/jszwedko/go-circleci"
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

func commandConnect() *Config {
	return &Config{
		Command: &model.Command{
			Trigger:          "connect",
			AutoComplete:     true,
			AutoCompleteDesc: "Connect to your circleci account.",
			AutoCompleteHint: "<auth-token>",
		},
		HelpText: "* auth token can be specified here.",
		Validate: validateConnect,
		Execute:  executeConnect,
	}
}

func validateConnect(args []string, context Context) (*model.CommandResponse, *model.AppError) {
	// we need the auth token
	if len(args) < 1 {
		return util.SendEphemeralText("Please specify the auth token")
	}

	return nil, nil
}

func executeConnect(args []string, context Context) (*model.CommandResponse, *model.AppError) {
	authToken := args[0]
	client := &circleci.Client{Token: authToken}
	user, err := client.Me()
	if err != nil {
		return util.SendEphemeralText("Unable to connect to circleci. Make sure the auth token is valid. " + err.Error())
	}

	if err := config.Mattermost.KVSet(context.CommandArgs.UserId+"_auth_token", []byte(authToken)); err != nil {
		return util.SendEphemeralText("Unable to save auth token to KVStore" + err.Error())
	}

	return util.SendEphemeralText("Successfully connected to circleci account: " + user.Login)
}
