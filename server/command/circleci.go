package command

import (
	"fmt"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/service"

	circleci "github.com/jszwedko/go-circleci"
	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

const (
	invalidCommand = "Invalid command parameters. Please use `/circleci help` for more information."
)

var CircleCICommandHandler = Handler{
	Command: &model.Command{
		Trigger:          "circleci",
		Description:      "Integration with CircleCI.",
		DisplayName:      "CircleCI",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect <token>, disconnect, me, projects, build, recent builds.",
		AutoCompleteHint: "[command]",
		Username:         config.BotUserName,
		IconURL:          config.BotIconURL,
	},
	handlers: map[string]HandlerFunc{
		"connect":            executeConnect,
		"disconnect":         executeDisconnect,
		"me":                 executeMe,
		"subscribe":          executeSubscribe,
		"unsubscribe":        executeUnsubscribe,
		"list/subscriptions": executeListSubscriptions,
		"projects":           executeListProjects,
		"build":              executeBuild,
		"recent/builds":      executeListRecentBuilds,
	},
	defaultHandler: func(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
		return util.SendEphemeralCommandResponse(invalidCommand)
	},
}

func executeSubscribe(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) != 3 {
		return util.SendEphemeralCommandResponse("Invalid number of arguments. syntax: `/circleci subscribe [vcs-alias] [org-name] [repo-name]`")
	}

	vcs, err := service.GetVCS(args[0])
	if err != nil {
		return util.SendEphemeralCommandResponse(err.Error())
	}

	newSubscription := serializer.Subscription{
		VCSType:   vcs.Type,
		BaseURL:   vcs.BaseURL,
		OrgName:   args[1],
		RepoName:  args[2],
		ChannelID: context.ChannelId,
	}

	if err := service.AddSubscription(newSubscription); err != nil {
		return util.SendEphemeralCommandResponse("Failed to add subscription. Please try again later. If the problem persists, contact your system administrator.")
	}

	return util.SendEphemeralCommandResponse("Subscription added successfully.")
}

func executeUnsubscribe(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) != 3 {
		return util.SendEphemeralCommandResponse("Invalid number of arguments. syntax: `/circleci unsubscribe [vcs-alias] [org-name] [repo-name]`")
	}

	vcs, err := service.GetVCS(args[0])
	if err != nil {
		return util.SendEphemeralCommandResponse(err.Error())
	}

	subscription := serializer.Subscription{
		VCSType:   vcs.Type,
		BaseURL:   vcs.BaseURL,
		OrgName:   args[1],
		RepoName:  args[2],
		ChannelID: context.ChannelId,
	}

	if err := service.RemoveSubscription(subscription); err != nil {
		return util.SendEphemeralCommandResponse("Failed to remove subscription. Please try again later. If the problem persists, contact your system administrator.")
	}

	return util.SendEphemeralCommandResponse("Subscription removed successfully.")
}

func executeListSubscriptions(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	subscriptions, err := service.ListSubscriptions(context.ChannelId)
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to fetch the list of subscriptions. Please use `/circleci subscribe` to create a subscription.")
	}

	if len(subscriptions) == 0 {
		return util.SendEphemeralCommandResponse("You have no notifications subscribed to this channel.\nUse `/circleci subscribe` to create a subscription.")
	}

	message := "| VcsType | BaseURL | Organisation | Repository |\n| :-- | --: | :-- | :-- |\n"
	for _, s := range subscriptions {
		message += fmt.Sprintf("| %s | %s | %s | %s |\n", s.VCSType, s.BaseURL, s.OrgName, s.RepoName)
	}

	return util.SendEphemeralCommandResponse(message)
}

func executeConnect(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	// we need the auth token
	if len(args) < 1 {
		return util.SendEphemeralCommandResponse("Please specify the auth token.")
	}

	authToken := args[0]
	client := &circleci.Client{Token: authToken}
	user, err := client.Me()
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is valid. Error: " + err.Error())
	}

	if err := config.Mattermost.KVSet(context.UserId+"_auth_token", []byte(authToken)); err != nil {
		config.Mattermost.LogError("Unable to save auth token to KVStore. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse("Successfully connected to circleci account: " + user.Login)
}

func executeDisconnect(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	if err := config.Mattermost.KVDelete(context.UserId + "_auth_token"); err != nil {
		config.Mattermost.LogError("Unable to disconnect. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse("Successfully disconnected.")
}

func executeMe(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	client := &circleci.Client{Token: string(authToken)}
	user, err := client.Me()
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. " + err.Error())
	}

	attachment := &model.SlackAttachment{
		Color:    "#7FC1EE",
		Pretext:  fmt.Sprintf("Initiated by CircleCI user: %s", user.Login),
		ThumbURL: user.AvatarURL,
		Fields: []*model.SlackAttachmentField{
			{
				Title: "Name",
				Value: user.Name,
				Short: true,
			},
			{
				Title: "Email",
				Value: user.SelectedEmail,
				Short: true,
			},
		},
	}

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}

func executeListProjects(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	client := &circleci.Client{Token: string(authToken)}
	projects, err := client.ListProjects()
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. Error: " + err.Error())
	}

	text := fmt.Sprintf("Here's a list of projects you follow on CircleCI:\n\n| Project | URL | OSS | ENV VARS |\n| :---- | :----- | :---- | :---- |\n")
	for _, project := range projects {
		envVars, err := client.ListEnvVars(project.Username, project.Reponame)
		if err != nil {
			return util.SendEphemeralCommandResponse(fmt.Sprintf("Problem listing env vars for %s/%s: %v", project.Username, project.Reponame, err))
		}

		circleURL := fmt.Sprintf("https://circleci.com/gh/%s/%s", project.Username, project.Reponame)
		text += fmt.Sprintf("| [%s/%s](%s) | %s | %t | %t |\n", project.Username, project.Reponame, project.VCSURL, circleURL, project.FeatureFlags.OSS, len(envVars) > 0)
	}

	attachment := &model.SlackAttachment{
		Color: "#7FC1EE",
		Text:  text,
	}

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}

func executeListRecentBuilds(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}
	client := &circleci.Client{Token: string(authToken)}

	var builds []*circleci.Build
	var err error

	text := "Recent Builds:\n"
	if len(args) == 3 {
		account, repo, branch := args[0], args[1], args[2]
		builds, err = client.ListRecentBuildsForProject(account, repo, branch, "", 30, 0)
	} else {
		builds, err = client.ListRecentBuilds(30, 0)
	}
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. Error: " + err.Error())
	}

	for _, build := range builds {
		text += fmt.Sprintf("- [%s/%s](%s): %s. Build: [%d](%s). Status: %s\n", build.Username, build.Reponame, build.VCSURL, build.Branch, build.BuildNum, build.BuildURL, build.Status)
	}

	attachment := &model.SlackAttachment{
		Color: "#7FC1EE",
		Text:  text,
	}

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}

func executeBuild(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	// we need the auth token
	if len(args) < 3 {
		return util.SendEphemeralCommandResponse("Please specify the account, repo and branch names.")
	}

	account, repo, branch := args[0], args[1], args[2]
	client := &circleci.Client{Token: string(authToken)}
	build, err := client.Build(account, repo, branch)
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. Error: " + err.Error())
	}

	attachment := &model.SlackAttachment{
		Color:   "#7FC1EE",
		Pretext: fmt.Sprintf("CircleCI build %d initiated successfully.", build.BuildNum),
		Text:    fmt.Sprintf("CircleCI build [%d](%s) initiated successfully.", build.BuildNum, build.BuildURL),
		Fields: []*model.SlackAttachmentField{
			{
				Title: "Commit Details",
				Value: build.AllCommitDetails,
				Short: false,
			},
			{
				Title: "User",
				Value: build.User.Login,
				Short: false,
			},
			{
				Title: "Account",
				Value: build.Username,
				Short: false,
			},
			{
				Title: "Repo",
				Value: build.Reponame,
				Short: false,
			},
			{
				Title: "Branch",
				Value: build.Branch,
				Short: false,
			},
		},
	}

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}
