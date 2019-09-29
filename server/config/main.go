package config

import (
	"github.com/mattermost/mattermost-server/plugin"
	"go.uber.org/atomic"
)

const (
	CommandPrefix             = PluginName
	URLMappingKeyPrefix       = "url_"
	ServerExeToWebappRootPath = "/../webapp"

	URLPluginBase = "/plugins/" + PluginName
	URLStaticBase = URLPluginBase + "/static"

	HeaderMattermostUserID = "Mattermost-User-Id"

	BotUserName    = "circleci"
	BotDisplayName = "CircleCI"
	BotDescription = "Created by the CircleCI Plugin."
)

var (
	config     atomic.Value
	Mattermost plugin.API
)

type Configuration struct {
	BotUserID string `json:"BotUserID,omitempty"`
}

func GetConfig() *Configuration {
	return config.Load().(*Configuration)
}

func SetConfig(c *Configuration) {
	config.Store(c)
}

func (c *Configuration) ProcessConfiguration() error {
	// any post-processing on configurations goes here

	return nil
}

func (c *Configuration) IsValid() error {
	// Add config validations here.
	// Check for required fields, formats, etc.

	return nil
}
