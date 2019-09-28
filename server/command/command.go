package command

import (
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "circleci",
		Description:      "A Mattermost plugin to interact with CircleCI",
		DisplayName:      "CircleCI",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect, disconnect, me, help",
		AutoCompleteHint: "[command]",
	}
}

func postCommandResponse(args *model.CommandArgs, text string) {
	botUserID := config.GetConfig().BotUserID
	post := &model.Post{
		UserId:    botUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = config.Mattermost.SendEphemeralPost(args.UserId, post)
}
