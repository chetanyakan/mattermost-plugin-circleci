package service

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

func SendWebhookNotifications(circleCIWebhook serializer.CircleCIWebhookRequest) error {
	b, err := config.Mattermost.KVGet(store.SubscriptionsKey)
	if err != nil {
		config.Mattermost.LogError("failed to get the list of subscriptions", "Error", err.Error())
		return err
	}

	subscriptions, appErr := serializer.SubscriptionsFromJSON(b)
	if appErr != nil {
		config.Mattermost.LogError("failed to deserialize the list of subscriptions", "Error", appErr.Error())
		return err
	}

	channelIDs := subscriptions.GetChannelIDs(circleCIWebhook.GetSubscription())
	if len(channelIDs) == 0 {
		config.Mattermost.LogWarn("Received CircleCI Webhook request, but it is not subscribed to any channels")
		return nil
	}

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

	for _, channelID := range channelIDs {
		post.ChannelId = channelID
		if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
			config.Mattermost.LogError("Failed to CircleCI status create the post in the channel.", "Error", appErr.Error(), "ChannelID", channelID)
		}
	}

	return nil
}
