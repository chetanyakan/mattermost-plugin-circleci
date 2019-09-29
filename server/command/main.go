package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

type Context struct {
	CommandArgs *model.CommandArgs
	Props       map[string]interface{}
}

type Config struct {
	Command  *model.Command
	HelpText string
	Execute  func([]string, Context) (*model.CommandResponse, *model.AppError)
	Validate func([]string, Context) (*model.CommandResponse, *model.AppError)
}

func (c *Config) Syntax() string {
	return fmt.Sprintf("/%s %s", c.Command.Trigger, c.Command.AutoCompleteHint)
}

var commands = map[string]*Config{
	commandConnect().Command.Trigger:    commandConnect(),
	commandDisconnect().Command.Trigger: commandDisconnect(),
	commandMe().Command.Trigger:         commandMe(),
	// commandRemoveMembers().Command.Trigger: commandRemoveMembers(),
	// commandStandup().Command.Trigger:       commandStandup(),
	// commandHelp().Command.Trigger:          commandHelp(),
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
