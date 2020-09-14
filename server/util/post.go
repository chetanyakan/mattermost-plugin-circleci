package util

import (
	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/mattermost/mattermost-server/model"
)

func BaseSlackAttachment() *model.SlackAttachment {
	return &model.SlackAttachment{
		AuthorIcon: config.BotIconURL,
		AuthorName: config.BotDisplayName,
		Color:      "#7FC1EE",
		ThumbURL:   config.BotThumbnail,
	}
}
