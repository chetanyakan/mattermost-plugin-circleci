package service

import (
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

const (
    // TODO: Configurable retry count
	kvStoreMaxRetryLimit = 3
)

func AddSubscription(newSubscription serializer.Subscription) error {
	for i := 0; i < kvStoreMaxRetryLimit; i++ {
		subscriptionsList, err := store.GetSubscriptions()
		if err != nil {
			// Get subscriptions failed. Try to create the first subscription
			subscriptionsList = nil
		}

		newSubscriptionsList := serializer.NewSubscriptions()
		if err := copier.Copy(&newSubscriptionsList, &subscriptionsList); err != nil {
			config.Mattermost.LogError("Failed to copy subscriptions list", "Error", err.Error())
			return err
		}

		newSubscriptionsList.Add(newSubscription)

		success, err := store.SaveSubscriptions(subscriptionsList, newSubscriptionsList)
		if err != nil {
			return err
		}
		if success {
			return nil
		}
	}

	return errors.New("Failed to add a subscription. Max retry limit reached.")
}

func RemoveSubscription(subscription serializer.Subscription) error {
	for i := 0; i < kvStoreMaxRetryLimit; i++ {
		subscriptionsList, err := store.GetSubscriptions()
		if err != nil {
			// Cannot remove subscription if it does not already exist
			return err
		}

		newSubscriptionsList := serializer.NewSubscriptions()
		if err := copier.Copy(&newSubscriptionsList, &subscriptionsList); err != nil {
			config.Mattermost.LogError("Failed to copy subscriptions list", "Error", err.Error())
			return err
		}

		newSubscriptionsList.Remove(subscription)

		success, err := store.SaveSubscriptions(subscriptionsList, newSubscriptionsList)
		if err != nil {
			return err
		}
		if success {
			return nil
		}
	}

	return errors.New("Failed to add a subscription. Max retry limit reached.")
}

func ListSubscriptions(channelID string) ([]serializer.Subscription, error) {
	subscriptions, err := store.GetSubscriptions()
	if err != nil {
		return nil, err
	}

	return subscriptions.List(channelID), nil
}
