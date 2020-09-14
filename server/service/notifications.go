package service

import (
	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
	"github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"
)

func SendWebhookNotifications(circleCIWebhook serializer.CircleCIWebhookRequest) error {
	var post *model.Post
	if circleCIWebhook.Status == "failure" {
		post = circleCIWebhook.GenerateFailurePost()
	} else if circleCIWebhook.Status == "success" {
		post = circleCIWebhook.GenerateSuccessPost()
	}

	if post == nil {
		config.Mattermost.LogError("failed to generate post from webhook")
		return errors.New("failed to generate post from webhook")
	}

	subscriptions, err := store.GetSubscriptions()
	if err != nil {
		return err
	}
	channelIDs := subscriptions.GetChannelIDs(circleCIWebhook.GetSubscription())

	for _, channelID := range channelIDs {
		post.ChannelId = channelID
		if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
			config.Mattermost.LogError("Failed to CircleCI status create the post in the channel.", "Error", appErr.Error(), "ChannelID", channelID)
		}
	}

	return nil
}
