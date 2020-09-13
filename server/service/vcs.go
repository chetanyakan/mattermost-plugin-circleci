package service

import (
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/store"
)

var (
	defaultVCSList = map[string]serializer.VCS{
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
		return &vcs, nil
	}

	vcs, err := store.GetVCS(alias)
	if err != nil {
		return nil, err
	}

	return vcs, nil
}
