package service

import (
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

func ListSubscriptions(channelID string) ([]serializer.Subscription, error) {
	subscriptions, err := store.GetSubscriptions()
	if err != nil {
		return nil, err
	}

	return subscriptions.List(channelID), nil
}
