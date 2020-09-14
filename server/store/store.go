package store

import (
	"encoding/json"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
)

const (
	subscriptionsKey = "subscriptions"
)

func getBytes(s interface{}) []byte {
	b, _ := json.Marshal(s)
	return b
}

func GetSubscriptions() (*serializer.Subscriptions, error) {
	b, err := config.Mattermost.KVGet(subscriptionsKey)
	if err != nil {
		config.Mattermost.LogError("failed to get the list of subscriptions", "Error", err.Error())
		return nil, err
	}

	s := serializer.Subscriptions{}
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func SaveSubscriptions(s *serializer.Subscriptions) error {
	// TODO: Check if we should use KVCompareAndSet to prevent race conditions
	if err := config.Mattermost.KVSet(subscriptionsKey, getBytes(s)); err != nil {
		config.Mattermost.LogError("failed to save the list of subscriptions in KVStore", "Error", err.Error())
		return err
	}
	return nil
}
