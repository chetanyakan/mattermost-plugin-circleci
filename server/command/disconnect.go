package command

import (
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

func executeDisconnect(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	if err := config.Mattermost.KVDelete(context.UserId + "_auth_token"); err != nil {
		config.Mattermost.LogError("Unable to disconnect. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse("Successfully disconnected.")
}
