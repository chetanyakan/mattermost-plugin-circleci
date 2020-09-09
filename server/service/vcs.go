package service

import (
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/pkg/errors"
)

var (
	vcsList = map[string]serializer.VCS{
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
	vcs, found := vcsList[alias]
	if !found {
		return nil, errors.New("Invalid VCS alias. Please use 'github' or 'bitbucket'.")
	}
	return &vcs, nil
}
