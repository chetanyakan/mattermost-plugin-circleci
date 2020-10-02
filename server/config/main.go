package config

import (
	"errors"
	"strings"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"go.uber.org/atomic"
)

const (
	CommandPrefix = "circleci"

	URLPluginBase = "/plugins/" + PluginName
	URLStaticBase = URLPluginBase + "/static"

	HeaderMattermostUserID = "Mattermost-User-Id"

	BotUserName    = "circleci"
	BotDisplayName = "CircleCI"
	BotDescription = "Created by the CircleCI Plugin."

	BotIconURL        = URLStaticBase + "/circleci.png"
	BotThumbnail      = URLStaticBase + "/circleci-blue.png"
	BotIconURLSuccess = URLStaticBase + "/circleci_success.png"
	BotIconURLFailed  = URLStaticBase + "/circeci_failed.png"

	WorkflowSuccessIconURL = URLStaticBase + "/icon_success.png"
	WorkflowFailedIconURL  = URLStaticBase + "/icon_failed.png"

	// TODO: Configurable retry count
	KVCompareAndSetMaxRetries = 5
)

var (
	config     atomic.Value
	Mattermost plugin.API
	BotUserID  string
)

type Configuration struct {
	Secret string `json:"Secret"`
}

func GetConfig() *Configuration {
	return config.Load().(*Configuration)
}

func SetConfig(c *Configuration) {
	config.Store(c)
}

// ProcessConfiguration is used for post-processing on configuration.
func (c *Configuration) ProcessConfiguration() error {
	c.Secret = strings.TrimSpace(c.Secret)

	return nil
}

// IsValid is used for config validations.
func (c *Configuration) IsValid() error {
	if c.Secret == "" {
		return errors.New("please provide the Webhook Secret")
	}

	return nil
}
