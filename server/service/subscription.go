package service

import (
	"encoding/json"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

func AddSubscription(newSubscription serializer.Subscription) error {
	err := store.AtomicModify(store.SubscriptionsKey, func(initialBytes []byte) ([]byte, error) {
		subscriptions, err := serializer.SubscriptionsFromJSON(initialBytes)
		if err != nil {
			return nil, err
		}

		subscriptions.Add(newSubscription)
		modifiedBytes, marshalErr := json.Marshal(subscriptions)
		if marshalErr != nil {
			return nil, marshalErr
		}
		return modifiedBytes, nil
	})

	if err != nil {
		config.Mattermost.LogError("Failed to Add subscription.", "Error", err.Error())
		return err
	}

	return nil
}

func RemoveSubscription(subscription serializer.Subscription) error {
	err := store.AtomicModify(store.SubscriptionsKey, func(initialBytes []byte) ([]byte, error) {
		subscriptions, err := serializer.SubscriptionsFromJSON(initialBytes)
		if err != nil {
			return nil, err
		}

		subscriptions.Remove(subscription)
		modifiedBytes, marshalErr := json.Marshal(subscriptions)
		if marshalErr != nil {
			return nil, marshalErr
		}
		return modifiedBytes, nil
	})

	if err != nil {
		config.Mattermost.LogError("Failed to Remove subscription.", "Error", err.Error())
		return err
	}

	return nil
}

func ListSubscriptions(channelID string) ([]serializer.Subscription, error) {
	b, err := config.Mattermost.KVGet(store.SubscriptionsKey)
	if err != nil {
		config.Mattermost.LogError("failed to get the list of subscriptions", "Error", err.Error())
		return nil, err
	}

	subscriptions, appErr := serializer.SubscriptionsFromJSON(b)
	if appErr != nil {
		config.Mattermost.LogError("failed to deserialize the list of subscriptions", "Error", appErr.Error())
		return nil, err
	}

	return subscriptions.List(channelID), nil
}
