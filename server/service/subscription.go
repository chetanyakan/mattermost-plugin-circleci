package service

import (
	"fmt"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

func AddSubscription(newSubscription serializer.Subscription) error {
	subscriptionsList, err := store.GetSubscriptions()
	if err != nil {
		subscriptionsList = serializer.NewSubscriptions()
	}
	subscriptionsList.Add(newSubscription)

	if err := store.SaveSubscriptions(subscriptionsList); err != nil {
		return err
	}

	return nil
}

func RemoveSubscription(subscription serializer.Subscription) error {
	subscriptionsList, err := store.GetSubscriptions()
	if err != nil {
		return err
	}

	subscriptionsList.Remove(subscription)
	if err := store.SaveSubscriptions(subscriptionsList); err != nil {
		return err
	}

	return nil
}

func getFormattedMessage(subscriptions []serializer.Subscription) string {
	if len(subscriptions) == 0 {
		return ""
	}

	message := "| VcsType | BaseURL | Organisation | Repository |\n| :-- | --: | :-- | :-- |\n"
	for _, s := range subscriptions {
		message += fmt.Sprintf("| %s | %s | %s | %s |\n", s.VCSType, s.BaseURL, s.OrgName, s.RepoName)
	}
	return message
}

func ListSubscriptions(channelID string) (string, error) {
	subscriptions, err := store.GetSubscriptions()
	if err != nil {
		return "", err
	}

	return getFormattedMessage(subscriptions.List(channelID)), nil
}
