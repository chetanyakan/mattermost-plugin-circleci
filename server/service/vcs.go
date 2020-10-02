package service

import (
	"errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

func GetVCS(alias string) (*serializer.VCS, error) {
	// first we check for default VCS, then in custom VCS
	if vcs, found := serializer.DefaultVCSList[alias]; found {
		return vcs, nil
	}

	vcs, err := store.GetVCS(alias)
	if err != nil {
		return nil, err
	}

	return vcs, nil
}

func AddVCS(vcs *serializer.VCS) error {
	if _, exists := serializer.DefaultVCSList[vcs.Alias]; exists {
		return errors.New("VCS alias already exists")
	}

	return store.SaveVCS(vcs)
}

func DeleteVCS(alias string) error {
	if _, exists := serializer.DefaultVCSList[alias]; exists {
		return errors.New("specified VCS is a system VCS and cannot be deleted")
	}

	return store.DeleteVCS(alias)
}

func GetVCSList() ([]*serializer.VCS, error) {
	vcsList, err := store.GetVCSList()
	if err != nil {
		return nil, err
	}

	for _, systemVCS := range serializer.DefaultVCSList {
		vcsList = append(vcsList, systemVCS)
	}

	return vcsList, nil
}
