# mattermost-plugin-circleci

![Go Version](https://img.shields.io/github/go-mod/go-version/chetanyakan/mattermost-plugin-circleci)
[![Go Report Card](https://goreportcard.com/badge/github.com/chetanyakan/mattermost-plugin-circleci)](https://goreportcard.com/report/github.com/chetanyakan/mattermost-plugin-circleci)
[![Build Status](https://img.shields.io/circleci/project/github/chetanyakan/mattermost-plugin-circleci/master)](https://circleci.com/gh/chetanyakan/mattermost-plugin-circleci)
[![Release](https://img.shields.io/github/v/release/chetanyakan/mattermost-plugin-circleci?include_prereleases)](https://github.com/chetanyakan/mattermost-plugin-circleci/releases/latest)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fchetanyakan%2Fmattermost-plugin-circleci.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fchetanyakan%2Fmattermost-plugin-circleci?ref=badge_shield)

### Platform & tools

- Make sure you have following components installed:

  - Go - v1.15 - [Getting Started](https://golang.org/doc/install)
    > **Note:** If you have installed Go to a custom location, make sure the `$GOROOT` variable is set properly. Refer [Installing to a custom location](https://golang.org/doc/install#install).
  - Make

## Setting up CircleCI

Set up CircleCI to run the build job for each branch and build-and-release for each tag.

1. Go to [CircleCI Dashboard](https://circleci.com/dashboard).
2. In the top left, you will find the Org switcher. Select your Organisation.
3. If this is your first project on CircleCI, go to the Projects page, click the **Add Projects** button, then click the **Set Up Project** button next to your project. You may also click **Start Building** to manually trigger your first build.
4. To manage GitHub releases using CircleCI, you need to add your github personal access token to your project's environment variables.
   - Follow the instructions [here](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to create a personal access token. For CircleCI releases, you would need the `repo` scope.
   - Add the environment variable to your project as `GITHUB_TOKEN` by following the instructions [here](https://circleci.com/docs/2.0/env-vars/#setting-an-environment-variable-in-a-project).
   
#### Usage Instructions

Bot User: CircleCI

Slash Commands:

- /connect
- /disconnect
- /me

Webhooks:

- For all finished builds (successful or failed)
