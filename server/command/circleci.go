package command

import (
	"fmt"
	circleci2 "github.com/TomTucka/go-circleci/circleci"
	"github.com/jszwedko/go-circleci"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
	"time"

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
		"connect":       executeConnect,
		"disconnect":    executeDisconnect,
		"me":            executeMe,
		"projects":      executeListProjects,
		"build":         executeBuild,
		"recent/builds": executeListRecentBuilds,
	},
	defaultHandler: func(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
		return util.SendEphemeralCommandResponse(invalidCommand)
	},
}

func executeConnect(ctx *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	// we need the auth token
	if len(args) < 1 {
		return util.SendEphemeralCommandResponse("Please specify the auth token.")
	}

	authToken := args[0]
	conf := circleci2.NewConfiguration()
	conf.AddDefaultHeader("Circle-Token", authToken)
	client := circleci2.NewAPIClient(conf)
	user, _, err := client.PreviewApi.GetCurrentUser(nil)
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is valid. Error: " + err.Error())
	}

	if err := config.Mattermost.KVSet(ctx.UserId+"_auth_token", []byte(authToken)); err != nil {
		config.Mattermost.LogError("Unable to save auth token to KVStore. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse(fmt.Sprintf("Successfully connected to circleci account: %s as %s", user.Name, user.Login))
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

	client := util.GetCircleciClient(string(authToken))
	user, _, err := client.PreviewApi.GetCurrentUser(nil)
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. " + err.Error())
	}

	attachment := &model.SlackAttachment{
		Color:    "#7FC1EE",
		Pretext:  fmt.Sprintf("Initiated by CircleCI user: %s", user.Login),
		ThumbURL: "/plugins/" + config.PluginName + "/static/circleci-blue.png",
		Title:    "CircleCI",
		Fields: []*model.SlackAttachmentField{
			{
				Title: "Name",
				Value: user.Name,
				Short: true,
			},
			{
				Title: "Username",
				Value: user.Login,
				Short: true,
			},
			{
				Title: "ID",
				Value: user.Id,
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
	client := util.GetCircleciClient(string(authToken))

	account, repo, branch := args[0], args[1], args[2]
	builds, _, err := client.InsightsApi.GetProjectWorkflowRuns(nil, account+"/"+repo, branch, utils.Yesterday(), utils.Yesterday().Add(2*24*time.Hour), nil)
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. Error: " + err.Error())
	}

	workflowCache := map[string]circleci2.Workflow{}
	pipelineCache := map[string]circleci2.Pipeline{}
	userCache := map[string]circleci2.User{}

	title := "Recent Builds"
	attachments := []*model.SlackAttachment{}
	for _, build := range builds.Items[:10] {
		var workflow circleci2.Workflow

		if _, exists := workflowCache[build.Id]; exists {
			workflow = workflowCache[build.Id]
		} else {
			workflow, _, err = client.WorkflowApi.GetWorkflowById(nil, build.Id)
			if err != nil {
				config.Mattermost.LogError("Unable to fetch workflow for ID: " + build.Id + ". Error: " + err.Error())
				return util.SendEphemeralCommandResponse("Unable to fetch data from CircleCI. Please try again.")
			}
		}

		var pipeline circleci2.Pipeline

		if _, exists := pipelineCache[workflow.PipelineId]; exists {
			pipeline = pipelineCache[workflow.PipelineId]
		} else {
			pipeline, _, err = client.PipelineApi.GetPipelineById(nil, workflow.PipelineId)
			if err != nil {
				config.Mattermost.LogError("Unable to fetch pipeline for ID: " + workflow.PipelineId + ". Error: " + err.Error())
				return util.SendEphemeralCommandResponse("Unable to fetch data from CircleCI. Please try again.")
			}
		}

		var triggeredBy circleci2.User

		if _, exists := userCache[workflow.StartedBy]; exists {
			triggeredBy = userCache[workflow.StartedBy]
		} else {
			triggeredBy, _, err = client.UserApi.GetUser(nil, workflow.StartedBy)
			if err != nil {
				config.Mattermost.LogError("Unable to fetch user for ID: " + workflow.StartedBy + ". Error: " + err.Error())
				return util.SendEphemeralCommandResponse("Unable to fetch data from CircleCI. Please try again.")
			}
		}

		attachments = append(attachments, &model.SlackAttachment{
			Title: fmt.Sprintf("https://app.circleci.com/pipelines/%s/%s/473/workflows/%s/", account, repo, workflow.Id),
			Color: "#04AA51",
			ThumbURL: "/plugins/" + config.PluginName + "/static/circleci-blue.png",
			Fields: []*model.SlackAttachmentField{
				{
					Short: true,
					Title: "Triggered By",
					Value: triggeredBy.Login,
				},
				{
					Short: true,
					Title: "Status",
					Value: build.Status,
				},
				{
					Short: true,
					Title: "VCS",
					Value: pipeline.Vcs.OriginRepositoryUrl,
				},
				{
					Short: true,
					Title: "Branch",
					Value: pipeline.Vcs.Branch,
				},
				{
					Short: true,
					Title: "Pipeline Number",
					Value: workflow.PipelineNumber,
				},
				{
					Short: true,
					Title: "Branch",
					Value: pipeline.Vcs.Branch,
				},
			},
		})
	}

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text: title,
		Attachments: attachments,
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
