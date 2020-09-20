package service

import (
	"errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

var (
	defaultVCSList = map[string]*serializer.VCS{
		"github": {
			Alias:   serializer.VCSTypeGithub,
			BaseURL: "https://github.com",
		},
		"bitbucket": {
			Alias:   serializer.VCSTypeBitbucket,
			BaseURL: "https://bitbucket.org",
		},
	}
)

func GetVCS(alias string) (*serializer.VCS, error) {
	// first we check for default VCS, then in custom VCS
	if vcs, found := defaultVCSList[alias]; found {
		return vcs, nil
	}

	vcs, err := store.GetVCS(alias)
	if err != nil {
		return nil, err
	}

	return vcs, nil
}

func AddVCS(vcs *serializer.VCS) error {
	if _, exists := defaultVCSList[vcs.Alias]; exists {
		return errors.New("VCS alias already exists")
	}

	return store.SaveVCS(vcs)
}

func DeleteVCS(alias string) error {
	if _, exists := defaultVCSList[alias]; exists {
		return errors.New("specified VCS is a system VCS and cannot be deleted")
	}

	return store.DeleteVCS(alias)
}

func GetVCSList() ([]*serializer.VCS, error) {
	vcsList, err := store.GetVCSList()
	if err != nil {
		return nil, err
	}

	for _, systemVCS := range defaultVCSList {
		vcsList = append(vcsList, systemVCS)
	}

	return vcsList, nil
}
