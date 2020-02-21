package command

import (
	circleci "github.com/jszwedko/go-circleci"
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

func executeConnect(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	// we need the auth token
	if len(args) < 1 {
		return util.SendEphemeralCommandResponse("Please specify the auth token.")
	}

	authToken := args[0]
	client := &circleci.Client{Token: authToken}
	user, err := client.Me()
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is valid. Error: " + err.Error())
	}

	if err := config.Mattermost.KVSet(context.UserId+"_auth_token", []byte(authToken)); err != nil {
		config.Mattermost.LogError("Unable to save auth token to KVStore. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse("Successfully connected to circleci account: " + user.Login)
}
