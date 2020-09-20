package store

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

const (
	subscriptionsKey = "subscriptions"
	vcsKeyPrefix     = "vcs_"
	listVCSKey       = "vcs_list"
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
		config.Mattermost.LogError(fmt.Sprintf("Error occurred unmarshaling VCS fetched from KV store. VCS alias: [%s], error: [%s]", alias, err.Error()))
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

	if err := addToVCSList(*vcs); err != nil {
		return err
	}

	return nil
}

func GetVCSList() ([]*serializer.VCS, error) {
	data, appErr := config.Mattermost.KVGet(listVCSKey)
	if appErr != nil {
		config.Mattermost.LogError("Failed to fetch list of VCS from KV store. Error: " + appErr.Error())
		return nil, errors.New(appErr.Error())
	}

	var vcsList []*serializer.VCS
	if err := json.Unmarshal(data, &vcsList); err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Failed to unmarshal VCS list. Error: %s", err.Error()))
		return nil, err
	}

	return vcsList, nil
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

func addToVCSList(vcs serializer.VCS) error {
	vcsListData, appErr := config.Mattermost.KVGet(listVCSKey)
	if appErr != nil {
		config.Mattermost.LogError("Failed to fetch list of VCS from KV store. Error: " + appErr.Error())
		return appErr
	}

	var dataToUnmarshal []byte
	if len(vcsListData) == 0 {
		dataToUnmarshal = []byte("[]")
	} else {
		dataToUnmarshal = vcsListData
	}

	var vcsList *[]serializer.VCS
	if err := json.Unmarshal(dataToUnmarshal, &vcsList); err != nil {
		config.Mattermost.LogError("Failed to unmarshal VCS list. Error: " + err.Error())
		return err
	}

	config.Mattermost.LogInfo(fmt.Sprintf("%v", vcsList))

	*vcsList = append(*vcsList, vcs)

	updatedVCSListData, vcsListErr := json.Marshal(*vcsList)
	if vcsListErr != nil {
		config.Mattermost.LogError("Failed to marshal updated VCS list. Error: " + vcsListErr.Error())
		return vcsListErr
	}

	err := util.KVCompareAndSet(listVCSKey, vcsListData, updatedVCSListData, func(oldData []byte) ([]byte, error) {
		var oldList []serializer.VCS
		if err := json.Unmarshal(oldData, &oldList); err != nil {
			config.Mattermost.LogError("Failed to unmarshal VCS list.", "Error", err.Error())
			return nil, err
		}

		vcsList := append(oldList, vcs)
		newList, err := json.Marshal(vcsList)
		if err != nil {
			config.Mattermost.LogError("Failed to marshal updated VCS list.", "Error", err.Error())
			return nil, err
		}

		return newList, nil
	})

	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Failed to save VCS to KV store by compare and set. VCS: [%v], error: [%s]", vcs, err.Error()))
		return err
	}

	return nil
}
