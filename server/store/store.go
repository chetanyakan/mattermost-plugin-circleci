package store

import (
	"encoding/json"

	"github.com/pkg/errors"

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
		return nil, errors.New(err.Error())
	}

	s := serializer.Subscriptions{}
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func SaveSubscriptions(originalSubscriptions, newSubscriptions *serializer.Subscriptions) (bool, error) {
	success, err := config.Mattermost.KVCompareAndSet(subscriptionsKey, getBytes(originalSubscriptions), getBytes(newSubscriptions))
	if err != nil {
		config.Mattermost.LogError("failed to save the list of subscriptions in KVStore", "Error", err.Error())
		return success, errors.New(err.Error())
	}
	return success, nil
}
