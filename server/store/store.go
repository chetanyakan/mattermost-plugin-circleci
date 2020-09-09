package store

import (
	"encoding/json"
	"fmt"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
)

const (
	subscriptionsKey = "subscriptions"
	vcsKeyPrefix = "vcs_"
)

func getBytes(s interface{}) []byte {
	b, _ := json.Marshal(s)
	return b
}

func GetSubscriptions() (*serializer.Subscriptions, error) {
	b, err := config.Mattermost.KVGet(subscriptionsKey)
	if err != nil {
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
		return err
	}
	return nil
}

func GetVCS(alias string) (*serializer.VCS, error) {
	key := vcsKeyPrefix + alias
	data, err := config.Mattermost.KVGet(key)
	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Error occurred fetching VCS from KV store. VCS alias: [%s], error: [%s]", alias, err.Error()))
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var vcs *serializer.VCS
	if err := json.Unmarshal(data, &vcs); err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Error occiurred unmarshaling VCS fetched from KV store. VCS alias: [%s], error: [%s]", alias, err.Error()))
		return nil, err
	}

	return vcs, nil
}

func SaveVCS(vcs *serializer.VCS) error {
	key := vcsKeyPrefix + vcs.Alias

	data, err := json.Marshal(vcs)
	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Error occurred marshaling VCS. Alias: [%s], base URL: [%s], error: [%s]", vcs.Alias, vcs.BaseURL, err.Error()))
		return err
	}

	if err := config.Mattermost.KVSet(key, data); err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Error occurred saving VCS in KV store. Alias: [%s], base URL: [%s], error: [%s]", vcs.Alias, vcs.BaseURL, err.Error()))
		return err
	}

	return nil

}

func DeleteVCS(alias string) error {
	key := vcsKeyPrefix + alias
	err := config.Mattermost.KVDelete(key)
	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Error occurred deleting VCS from KV store. Alias: [%s], error: [%s]", alias, err.Error()))
		return err
	}

	return nil
}
