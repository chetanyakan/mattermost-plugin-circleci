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
	// commandHelp().Command.Trigger:          commandHelp(),
}

func sendEphemeralMessage(userID, channelID, message string) {
	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: channelID,
		Message:   message,
	}
	_ = config.Mattermost.SendEphemeralPost(userID, post)
}
