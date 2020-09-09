package service

import (
	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func SendWebhookNotifications(req serializer.CircleCIWebhookRequest) error {
	var post *model.Post
	if req.Status == "failure" {
		post = req.GenerateFailurePost()
	} else if req.Status == "success" {
		post = req.GenerateSuccessPost()
	}

	if post == nil {
		config.Mattermost.LogError("failed to generate post from webhook")
		return errors.New("failed to generate post from webhook")
	}

	subscriptions, err := store.GetSubscriptions()
	if err != nil {
		config.Mattermost.LogError("failed to get the list of subscriptions", "Error", err.Error())
		return err
	}
	channelIDs := subscriptions.GetChannelIDs(req.GetSubscription())

	for _, channelID := range channelIDs {
		post.ChannelId = channelID
		if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
			config.Mattermost.LogError(appErr.Error())
		}
	}

	return nil
}
