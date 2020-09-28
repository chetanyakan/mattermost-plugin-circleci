package command

import (
	"context"
	"fmt"
	"github.com/thoas/go-funk"
	"net/http"
	"strings"
	"time"

	circleci2 "github.com/TomTucka/go-circleci/circleci"
	"github.com/antihax/optional"
	"github.com/dustin/go-humanize"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/serializer"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/service"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

const (
	invalidCommand = "Invalid command parameters. Please use `/circleci help` for more information."
	HeadTypeBranch = "branch"
	HeadTypeTag    = "tag"
)

type command struct {
	Execute          HandlerFunc
	AutocompleteData *model.AutocompleteData
}

var commandConnect = &command{
	Execute: executeConnect,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "connect",
		HelpText: "Connect with CircleCI account",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "CircleCI Auth Token",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "CircleCI Auth Token",
					Pattern: ".+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandDisconnect = &command{
	Execute: executeDisconnect,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "disconnect",
		HelpText: "Disconnect your CircleCI account",
	},
}

var commandSubscribe = &command{
	Execute: executeSubscribe,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "subscribe",
		HelpText: "Subscribe to specified CircleCI notifications in the current channel",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandUnsubscribe = &command{
	Execute: executeUnsubscribe,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "unsubscribe",
		HelpText: "Unsubscribe to specified CircleCI notifications in the current channel",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandListSubscriptions = &command{
	Execute: executeListSubscriptions,
	AutocompleteData: &model.AutocompleteData{
		Trigger:     "list subscriptions",
		HelpText:    "Get list of CircleCI notifications subscribed in the current channel",
		Arguments:   nil,
		SubCommands: nil,
	},
}

var commandBuild = &command{
	Execute: executeBuild,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "build",
		HelpText: "Trigger the specified build.",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Head Type",
				Type:     model.AutocompleteArgTypeStaticList,
				Required: true,
				Data: &model.AutocompleteStaticListArg{
					PossibleArguments: []model.AutocompleteListItem{
						{
							Item:     "branch",
							HelpText: "Build against a branch",
						},
						{
							Item:     "tag",
							HelpText: "Build against a tag",
						},
					},
				},
			},
			{
				HelpText: "Head to build against",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Branch or tag name to build against.",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandRecentBuilds = &command{
	Execute: executeListRecentBuilds,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "recent builds",
		HelpText: "List recent builds of specified pipeline",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Workflow name to list recent builds of",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Workflow name to list recent builds of. Example - `build`, `release`.",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandAddVCS = &command{
	Execute: executeAddVCS,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "add vcs",
		HelpText: "Add new VCS alias",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Name to be used as VCS alias",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Base URL of the VCS. This is the URL you see in CircleCI when viewing a build. For example - `github` is the base URL in `https://app.circleci.com/pipelines/github/foobar`",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS base URL",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandDeleteVCS = &command{
	Execute: executeDeleteVCS,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "delete vcs",
		HelpText: "Delete an existing VCS alias",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Name to be used as VCS alias",
					Pattern: ".+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandListVCS = &command{
	Execute: executeListVCS,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "list vcs",
		HelpText: "List all available VCS.",
	},
}

var commandProjectSummary = &command{
	Execute: executeProjectSummary,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "project summary",
		HelpText: "Show project summary",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "Project",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list projects` to view available projects",
					Pattern: ".+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandGetPipelineByNumber = &command{
	Execute: executeGetPipelineByNumber,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "pipeline",
		HelpText: "Get details of a pipeline.",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Pipeline Number",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Pipeline number",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandGetEnvironmentVariables = &command{
	Execute: executeGetAllEnvironmentVariables,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "environment",
		HelpText: "Get masked environment variables for a project.",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var commandRecentWorkflowRuns = &command{
	Execute: executeRecentWorkflowRuns,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "workflow insights",
		HelpText: "Get insight for a workflow's recent runs.",
		Arguments: []*model.AutocompleteArg{
			{
				HelpText: "VCS Alias",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "VCS Alias. Use `/circle list vcs` to view available VCS",
					Pattern: ".+",
				},
			},
			{
				HelpText: "Org Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Org name on the VCS. For example org name for `github.com/foo/bar` would be `foo`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Repository Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Repository name on the VCS. For example repository name for `github.com/foo/bar` would be `bar`.",
					Pattern: "._+",
				},
			},
			{
				HelpText: "Workflow Name",
				Type:     model.AutocompleteArgTypeText,
				Required: true,
				Data: &model.AutocompleteTextArg{
					Hint:    "Workflow's name",
					Pattern: "._+",
				},
			},
		},
		SubCommands: nil,
	},
}

var CircleCICommandHandler = Handler{
	Command: &model.Command{
		Trigger:          "circleci",
		Description:      "Integration with CircleCI.",
		DisplayName:      "CircleCI",
		AutoComplete:     true,
		Username:         config.BotUserName,
		IconURL:          config.BotIconURL,
		AutocompleteData: &model.AutocompleteData{
			Trigger:  "circleci",
			HelpText: "interact with CircleCI right from with Mattermost",
			SubCommands: []*model.AutocompleteData{
				commandConnect.AutocompleteData,
				commandDisconnect.AutocompleteData,
				commandSubscribe.AutocompleteData,
				commandUnsubscribe.AutocompleteData,
				commandListSubscriptions.AutocompleteData,
				commandBuild.AutocompleteData,
				commandRecentBuilds.AutocompleteData,
				// These can be used later when adding Github and Bitbucket on-premise support
				//commandAddVCS.AutocompleteData,
				//commandDeleteVCS.AutocompleteData,
				//commandListVCS.AutocompleteData,
				commandProjectSummary.AutocompleteData,
				commandGetPipelineByNumber.AutocompleteData,
				commandGetEnvironmentVariables.AutocompleteData,
				commandRecentWorkflowRuns.AutocompleteData,
			},
		},
	},
	handlers: map[string]HandlerFunc{
		"connect":            commandConnect.Execute,
		"disconnect":         commandDisconnect.Execute,
		"subscribe":          commandSubscribe.Execute,
		"unsubscribe":        commandUnsubscribe.Execute,
		"list/subscriptions": commandListSubscriptions.Execute,
		"build":              commandBuild.Execute,
		"recent/builds":      commandRecentBuilds.Execute,
		// These can be used later when adding Github and Bitbucket on-premise support
		//"add/vcs":            commandAddVCS.Execute,
		//"delete/vcs":         commandDeleteVCS.Execute,
		//"list/vcs":           commandListVCS.Execute,
		"project/summary":    commandProjectSummary.Execute,
		"pipeline":           commandGetPipelineByNumber.Execute,
		"environment":        commandGetEnvironmentVariables.Execute,
		"workflow/insights":  commandRecentWorkflowRuns.Execute,
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
		VCSType:   vcs.Alias,
		BaseURL:   vcs.BaseURL,
		OrgName:   args[1],
		RepoName:  args[2],
		ChannelID: context.ChannelId,
	}

	if err := newSubscription.Validate(); err != nil {
		return util.SendEphemeralCommandResponse(fmt.Sprintf("Failed to validate subscription details. Error: %v", err.Error()))
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
		VCSType:   vcs.Alias,
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

	message := "| VcsType | BaseURL | Organization | Repository |\n| :-- | --: | :-- | :-- |\n"
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
	conf := circleci2.NewConfiguration()
	conf.AddDefaultHeader("Circle-Token", authToken)

	if err := config.Mattermost.KVSet(context.UserId+"_auth_token", []byte(authToken)); err != nil {
		config.Mattermost.LogError("Unable to save auth token to KVStore. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse("Successfully saved auth token.")
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

func executeListRecentBuilds(ctx *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(ctx.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}
	client := util.GetCircleciClient(string(authToken))

	vcsAlias, org, repo, workflow := args[0], args[1], args[2], args[3]

	vcs, err := service.GetVCS(vcsAlias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to get VCS details. Please try again later. If the problem persists, contact your system administrator.")
	}

	builds, resp, err := client.InsightsApi.GetProjectWorkflowRuns(
		nil,
		fmt.Sprintf("%s/%s/%s", vcs.Type, org, repo), workflow, utils.Yesterday(), utils.Yesterday().Add(2*24*time.Hour),
		nil,
	)

	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. Error: " + err.Error())
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	workflowCache := map[string]circleci2.Workflow{}
	pipelineCache := map[string]circleci2.Pipeline{}
	userCache := map[string]circleci2.User{}

	attachment := util.BaseSlackAttachment()

	for _, build := range builds.Items[:10] {
		var workflow circleci2.Workflow

		if _, exists := workflowCache[build.Id]; exists {
			workflow = workflowCache[build.Id]
		} else {
			workflow, resp, err = client.WorkflowApi.GetWorkflowById(context.TODO(), build.Id)
			if err != nil {
				config.Mattermost.LogError("Unable to fetch workflow for ID: " + build.Id + ". Error: " + err.Error())
				return util.SendEphemeralCommandResponse("Unable to fetch data from CircleCI. Please try again.")
			}
			if resp != nil {
				defer resp.Body.Close()
			}
		}

		var pipeline circleci2.Pipeline

		if _, exists := pipelineCache[workflow.PipelineId]; exists {
			pipeline = pipelineCache[workflow.PipelineId]
		} else {
			pipeline, resp, err = client.PipelineApi.GetPipelineById(context.TODO(), workflow.PipelineId)
			if err != nil {
				config.Mattermost.LogError("Unable to fetch pipeline for ID: " + workflow.PipelineId + ". Error: " + err.Error())
				return util.SendEphemeralCommandResponse("Unable to fetch data from CircleCI. Please try again.")
			}
			if resp != nil {
				defer resp.Body.Close()
			}
		}

		var triggeredBy circleci2.User

		if _, exists := userCache[workflow.StartedBy]; exists {
			triggeredBy = userCache[workflow.StartedBy]
		} else {
			triggeredBy, resp, err = client.UserApi.GetUser(context.TODO(), workflow.StartedBy)
			if err != nil {
				config.Mattermost.LogError("Unable to fetch user for ID: " + workflow.StartedBy + ". Error: " + err.Error())
				return util.SendEphemeralCommandResponse("Unable to fetch data from CircleCI. Please try again.")
			}
			if resp != nil {
				defer resp.Body.Close()
			}
		}

		attachment.Fields = append(attachment.Fields, []*model.SlackAttachmentField{
			{
				Short: false,
				Title: "",
				Value: "***",
			},
			{
				Short: false,
				Title: fmt.Sprintf("Build #%d", workflow.PipelineNumber),
				Value: fmt.Sprintf("https://app.circleci.com/pipelines/%s/%s/%d/workflows/%s/", vcs, repo, workflow.PipelineNumber, workflow.Id),
			},
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
				Value: fmt.Sprintf("%d", workflow.PipelineNumber),
			},
		}...)
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: ctx.ChannelId,
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	if _, err := config.Mattermost.CreatePost(post); err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Failed to create post for triggered workflow. ChannelID: %s, error: %s", ctx.ChannelId, err.Error()))
		return util.SendEphemeralCommandResponse("Failed to create post for triggered workflow. The build has been triggered though. You can view it in CircleCI.")
	}

	return &model.CommandResponse{}, nil
}

func executeBuild(ctx *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(ctx.UserId + "_auth_token")
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

	vcsAlias, org, repo, headType, head := args[0], args[1], args[2], args[3], args[4]

	vcs, err := service.GetVCS(vcsAlias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to get VCS details. Please try again later. If the problem persists, contact your system administrator.")
	}

	client := util.GetCircleciClient(string(authToken))

	// TODO need to add support for tag here
	body := map[string]string{}

	switch headType {
	case HeadTypeBranch:
		body[HeadTypeBranch] = head
	case HeadTypeTag:
		body[HeadTypeTag] = head
	default:
		return util.SendEphemeralCommandResponse(fmt.Sprintf("Invalid head type. Please specify one of `%s` or `%s`", HeadTypeBranch, HeadTypeTag))
	}

	build, response, err := client.PipelineApi.TriggerPipeline(nil, vcs.Type+"/"+org+"/"+repo, &circleci2.PipelineApiTriggerPipelineOpts{
		Body: optional.NewInterface(body),
	})

	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to trigger build. VCS slug: %s, repo: %s, head type: %s, head: %s, error: %s, response: %v",
			vcsAlias,
			repo,
			headType,
			head,
			err.Error(),
			response,
		))

		var message string

		if response != nil && response.StatusCode == http.StatusNotFound {
			message = "Unable to trigger build. Either the specified workflow doesn't exist or the  auth token is not valid."
		} else {
			message = "Unable to trigger build."
		}

		return util.SendEphemeralCommandResponse(message)
	}
	if response != nil {
		defer response.Body.Close()
	}

	workflows, response, err := client.PipelineApi.ListWorkflowsByPipelineId(context.TODO(), build.Id, nil)
	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to fetch pipeline details. Pipeline ID: %s, VCS slug: %s, repo: %s, head type: %s, head: %s, error: %s, response: %v",
			build.Id,
			vcsAlias,
			repo,
			headType,
			head,
			err.Error(),
			response,
		))

		return util.SendEphemeralCommandResponse("Successfully trigger build but failed to fetch triggered build's details. You  can still view the triggered build in CircleCI.")
	}
	if response != nil {
		defer response.Body.Close()
	}

	attachmentFields := make([]*model.SlackAttachmentField, len(workflows.Items))

	for i, workflow := range workflows.Items {
		attachmentFields[i] = &model.SlackAttachmentField{
			Short: false,
			Title: "Workflow: " + strings.Title(workflow.Name),
			Value: fmt.Sprintf("https://app.circleci.com/pipelines/%s/%d/workflows/%s", vcsAlias+"/"+repo, build.Number, workflow.Id),
		}
	}

	attachmentFields = append(
		attachmentFields,
		&model.SlackAttachmentField{
			Title: "Build Number",
			Value: build.Number,
			Short: true,
		},
		&model.SlackAttachmentField{
			Title: "Build ID",
			Value: build.Id,
			Short: true,
		},
		&model.SlackAttachmentField{
			Title: "State",
			Value: build.State,
			Short: true,
		},
	)

	attachment := util.BaseSlackAttachment()
	attachment.Pretext = fmt.Sprintf("CircleCI build %d initiated successfully.", build.Number)
	attachment.Text = fmt.Sprintf("CircleCI build %d initiated successfully.", build.Number)
	attachment.Fields = attachmentFields

	return &model.CommandResponse{
		Username:    config.BotDisplayName,
		IconURL:     config.BotIconURL,
		Type:        model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Attachments: []*model.SlackAttachment{attachment},
	}, nil
}

func executeAddVCS(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	config.Mattermost.LogInfo(fmt.Sprintf("%v", args))
	if len(args) < 3 {
		return util.SendEphemeralCommandResponse("Invalid number of arguments. Use this command as `/cirecleci add vcs [github | bitbucket] [alias] [base URL]`")
	}

	vcsType, alias, baseURL := args[0], args[1], args[2]

	existingVCS, err := service.GetVCS(alias)
	if err != nil {
		return util.SendEphemeralCommandResponse(
			"Failed to check for existing VCS with same alias. Please try again later. If the problem persists, contact your system administrator.",
		)
	}

	if existingVCS != nil {
		return util.SendEphemeralCommandResponse(fmt.Sprintf("Another VCS existis with the same alias. Please delete existing VCS first if you want to update it. Alias: `%s`, base URL: `%s`", existingVCS.Alias, existingVCS.BaseURL))
	}

	vcs := &serializer.VCS{
		Alias:   alias,
		BaseURL: baseURL,
		Type:    vcsType,
	}

	if err := service.AddVCS(vcs); err != nil {
		return util.SendEphemeralCommandResponse("Failed to save VCS. Please try again later. If the problem persists, contact your system administrator.")
	}

	message := fmt.Sprintf("Successfully added VCS with alias `%s` and base URL `%s`", vcs.Alias, vcs.BaseURL)

	_, _ = config.Mattermost.CreatePost(&model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
		Message:   message,
	})

	return &model.CommandResponse{}, nil
}

func executeDeleteVCS(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) < 1 {
		return util.SendEphemeralCommandResponse("Invalid number of arguments. Use this command as `/cirecleci delete vcs [alias]`")
	}

	alias := args[0]

	existingVCS, err := service.GetVCS(alias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to check VCS. Please try again later. If the problem persists, contact your system administrator.")
	}

	if existingVCS == nil {
		return util.SendEphemeralCommandResponse("No VCS exists with provided alias.")
	}

	if err := service.DeleteVCS(alias); err != nil {
		return util.SendEphemeralCommandResponse("Failed to delete VCS. Please try again later. If the problem persists, contact your system administrator.")
	}

	message := fmt.Sprintf("Successfully deleted VCS with alias `%s`", alias)

	_, _ = config.Mattermost.CreatePost(&model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
		Message:   message,
	})

	return &model.CommandResponse{}, nil
}

func executeListVCS(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	vcsList, err := service.GetVCSList()
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to fetch list of VCS. Please try again later. If the problem persists, contact your system administrator.")
	}

	message := "Available VCS -\n\n| No.  | Type | Alias | Base URL |\n|:------------|:------------|:------------|:------------|\n"
	for i, vcs := range vcsList {
		message += fmt.Sprintf("|%d|%s|%s|%s|\n", i+1, vcs.Type, vcs.Alias, vcs.BaseURL)
	}

	_, _ = config.Mattermost.CreatePost(&model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
		Message:   message,
	})

	return &model.CommandResponse{}, nil
}

// executeProjectSummary - uses insight API
func executeProjectSummary(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) < 1 {
		return util.SendEphemeralCommandResponse("Incorrect syntax. Use this command as `/circleci project summary <VCS alias> <org> <repo>`")
	}

	vcsAlias, org, repo := args[0], args[1], args[2]
	vcs, err := service.GetVCS(vcsAlias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to get VCS details. Please try again later. If the problem persists, contact your system administrator.")
	}

	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}
	client := util.GetCircleciClient(string(authToken))

	insights, response, err := client.InsightsApi.GetProjectWorkflowMetrics(nil, fmt.Sprintf("%s/%s/%s", vcs.Type, org, repo), nil)
	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to fetch project summary. Project slug: %s, error: %s",
			args[1],
			err.Error(),
		))

		return util.SendEphemeralCommandResponse(
			"Failed to fetch project summary from CircleCI. Please try again later. If the problem persists, contact your system administrator.",
		)
	}

	if response != nil {
		defer response.Body.Close()
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
	}

	attachments := make([]*model.SlackAttachment, len(insights.Items))

	for i, insight := range insights.Items {
		attachment := util.BaseSlackAttachment()
		attachment.Title = fmt.Sprintf(
			"Project Summary: %s | %s | %s to %s",
			fmt.Sprintf("%s/%s/%s", vcs.Type, org, repo),
			strings.Title(insight.Name),
			insight.WindowStart.Format(time.UnixDate),
			insight.WindowEnd.Format(time.UnixDate),
		)

		attachment.Fields = []*model.SlackAttachmentField{
			{
				Short: true,
				Title: "Total Runs",
				Value: fmt.Sprintf("%d", insight.Metrics.TotalRuns),
			},
			{
				Short: true,
				Title: "Successful Runs",
				Value: fmt.Sprintf("%d", insight.Metrics.SuccessfulRuns),
			},
			{
				Short: true,
				Title: "Failed Runs",
				Value: fmt.Sprintf("%d", insight.Metrics.FailedRuns),
			},
			{
				Short: true,
				Title: "Success Rate",
				Value: fmt.Sprintf("%f", insight.Metrics.SuccessRate),
			},
			{
				Short: true,
				Title: "Throughput",
				Value: fmt.Sprintf("%f", insight.Metrics.Throughput),
			},
			{
				Short: true,
				Title: "MTTR",
				Value: fmt.Sprintf("%d", insight.Metrics.Mttr),
			},
			{
				Short: true,
				Title: "Total Credits Used",
				Value: fmt.Sprintf("%d", insight.Metrics.TotalCreditsUsed),
			},
			{
				Short: true,
				Title: "Duration: Min",
				Value: fmt.Sprintf("%d", insight.Metrics.DurationMetrics.Min),
			},
			{
				Short: true,
				Title: "Duration: Max",
				Value: fmt.Sprintf("%d", insight.Metrics.DurationMetrics.Max),
			},
			{
				Short: true,
				Title: "Duration: Median",
				Value: fmt.Sprintf("%d", insight.Metrics.DurationMetrics.Median),
			},
			{
				Short: true,
				Title: "Duration: Mean",
				Value: fmt.Sprintf("%d", insight.Metrics.DurationMetrics.Mean),
			},
			{
				Short: true,
				Title: "Duration: P95",
				Value: fmt.Sprintf("%d", insight.Metrics.DurationMetrics.P95),
			},
			{
				Short: true,
				Title: "Duration: Standard Deviation",
				Value: fmt.Sprintf("%f", insight.Metrics.DurationMetrics.StandardDeviation),
			},
		}
		attachments[i] = attachment
	}

	model.ParseSlackAttachment(post, attachments)

	_, appErr = config.Mattermost.CreatePost(post)
	if appErr != nil {
		config.Mattermost.LogError(fmt.Sprintf("Failed to create post for project summary. Project: %s, error: %s", args[1], appErr.Error()))
		return util.SendEphemeralCommandResponse(
			"Failed to create post for project summary. Please try again later. If the problem persists, contact your system administrator.",
		)
	}

	return &model.CommandResponse{}, nil
}

// executeGetPipelineByNumber - uses pipeline API
func executeGetPipelineByNumber(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) < 4 {
		return util.SendEphemeralCommandResponse("Incorrect syntax. Use this command as `/circleci pipeline <vcs alias> <org> <repo> <pipeline number>`")
	}

	vcsAlias, org, repo, pipelineNumber := args[0], args[1], args[2], args[3]

	vcs, err := service.GetVCS(vcsAlias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to get VCS details. Please try again later. If the problem persists, contact your system administrator.")
	}

	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}

	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	client := util.GetCircleciClient(string(authToken))
	pipeline, pipelineResponse, err := client.PipelineApi.GetPipelineByNumber(nil, fmt.Sprintf("%s/%s/%s", vcs.Type, org, repo), pipelineNumber)

	if pipelineResponse != nil {
		defer pipelineResponse.Body.Close()
	}

	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to get pipeline data by number. VCS: %s, org: %s, repo: %s, pipeline number: %s, error: %s",
			vcsAlias,
			org,
			repo,
			pipelineNumber,
			err.Error(),
		))

		return util.SendEphemeralCommandResponse("Failed to get pipeline details. Please try again later. If the problem persists, contact your system administrator.")
	}

	workflows, workflowResponse, err := client.PipelineApi.ListWorkflowsByPipelineId(nil, pipeline.Id, nil)

	if workflowResponse != nil {
		defer workflowResponse.Body.Close()
	}

	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to get pipeline's workflows . VCS: %s, org: %s, repo: %s, pipeline ID: %s, error: %s",
			vcsAlias,
			org,
			repo,
			pipeline.Id,
			err.Error(),
		))

		return util.SendEphemeralCommandResponse("Failed to get pipeline's workflows. Please try again later. If the problem persists, contact your system administrator.")
	}

	jobsByWorkflow := map[circleci2.Workflow1]circleci2.WorkflowJobListResponse{}

	for i, workflow := range workflows.Items {
		jobs, _, err := client.WorkflowApi.ListWorkflowJobs(nil, workflow.Id)

		if err != nil {
			config.Mattermost.LogError(fmt.Sprintf(
				"Failed to get workflow's job. VCS: %s, org: %s, repo: %s, pipeline ID: %s, workflow ID: %s, error: %s",
				vcsAlias,
				org,
				repo,
				pipeline.Id,
				workflow.Id,
				err.Error(),
			))

			return util.SendEphemeralCommandResponse("Failed to get workflow's jobs. Please try again later. If the problem persists, contact your system administrator.")
		}

		jobsByWorkflow[workflows.Items[i]] = jobs
	}

	attachment := util.BaseSlackAttachment()
	attachment.Title = fmt.Sprintf("Pipeline #%s (%s)", pipelineNumber, pipeline.ProjectSlug)
	attachment.Fields = []*model.SlackAttachmentField{
		{
			Short: true,
			Title: "Created On",
			Value: pipeline.CreatedAt.Format(time.UnixDate),
		},
		{
			Short: true,
			Title: "triggered By",
			Value: pipeline.Trigger.Actor.Login,
		},
		{
			Short: false,
			Title: "",
			Value: "***",
		},
	}

	for workflow, jobs := range jobsByWorkflow {
		fields := []*model.SlackAttachmentField{
			{
				Short: true,
				Title: "Workflow",
				Value: workflow.Name,
			},
			{
				Short: true,
				Title: "Status",
				Value: workflow.Status,
			},
		}

		for _, job := range jobs.Items {
			jobFields := []*model.SlackAttachmentField{
				{
					Short: false,
					Title: "",
					Value: "***",
				},
				{
					Short: false,
					Title: "Job ",
					Value: strings.Title(job.Name),
				},
				{
					Short: true,
					Title: "Job Number",
					Value: fmt.Sprintf("%d", job.JobNumber),
				},
				{
					Short: true,
					Title: "Status",
					Value: job.Status,
				},
				{
					Short: true,
					Title: "Type",
					Value: job.Type_,
				},
				{
					Short: true,
					Title: "Started At",
					Value: job.StartedAt.Format(time.UnixDate),
				},
				{
					Short: true,
					Title: "Ended At",
					Value: job.StoppedAt.Format(time.UnixDate),
				},
			}

			fields = append(fields, jobFields...)
		}

		attachment.Fields = append(attachment.Fields, fields...)
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})

	if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to create post for workflow by number. VCS: %s, org: %s, repo: %s, pipeline ID: %s, channelID: %s, error: %s",
			vcsAlias,
			org,
			repo,
			pipeline.Id,
			context.ChannelId,
			appErr.Error(),
		))

		return util.SendEphemeralCommandResponse("Failed to crete post. Please try again later. If the problem persists, contact your system administrator.")
	}

	return &model.CommandResponse{}, nil
}

// executeGetAllEnvironmentVariables - uses project API
func executeGetAllEnvironmentVariables(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) < 3 {
		return util.SendEphemeralCommandResponse("Incorrect syntax. Use this command as `/circleci env vars <vcs alias> <org> <repo>`")
	}

	vcsAlias, org, repo := args[0], args[1], args[2]

	vcs, err := service.GetVCS(vcsAlias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to get VCS details. Please try again later. If the problem persists, contact your system administrator.")
	}

	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}

	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	client := util.GetCircleciClient(string(authToken))
	projectSlug := fmt.Sprintf("%s/%s/%s", vcs.Type, org, repo)

	envVars, response, err := client.ProjectApi.ListEnvVars(nil, projectSlug)
	if response != nil {
		defer response.Body.Close()
	}

	attachment := util.BaseSlackAttachment()
	attachment.Title = "Masked Environment Variables for : " + projectSlug
	attachment.Fields = make([]*model.SlackAttachmentField, len(envVars.Items))

	for i, envVar := range envVars.Items {
		attachment.Fields[i] = &model.SlackAttachmentField{
			Short: true,
			Title: envVar.Name,
			Value: envVar.Value,
		}
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
	}
	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
		config.Mattermost.LogError(fmt.Sprintf("Could not post environment variable post. Channel ID: %s, error: %s", context.ChannelId, appErr.Error()))
		return util.SendEphemeralCommandResponse("Could not create post. Please try again later. If the problem persists, contact your system administrator.")
	}

	return &model.CommandResponse{}, nil
}

func executeRecentWorkflowRuns(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	if len(args) < 4 {
		return util.SendEphemeralCommandResponse("Incorrect syntax. Use this command as `/circleci workflow insights <vcs alias> <org> <repo> <workflow name>`")
	}

	vcsAlias, org, repo, workflowName := args[0], args[1], args[2], args[3]

	vcs, err := service.GetVCS(vcsAlias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to get VCS details. Please try again later. If the problem persists, contact your system administrator.")
	}

	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}

	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	client := util.GetCircleciClient(string(authToken))
	projectSlug := fmt.Sprintf("%s/%s/%s", vcs.Type, org, repo)

	workflowRuns, response, err := client.InsightsApi.GetProjectWorkflowRuns(
		nil,
		projectSlug,
		workflowName,
		utils.Yesterday(),
		utils.Yesterday().Add(24*time.Hour),
		nil,
	)
	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to fetch workflow insight from CircleCI. Project slug: %s, workflow name: %s, err: %s",
			projectSlug,
			workflowName,
			err.Error(),
		))

		return util.SendEphemeralCommandResponse(
			"Failed to fetch workflow insights from CircleCI. Please try again later. If the problem persists, contact your system administrator.",
		)
	}

	runsPerPost := 10

	// Having too many attachments in a post can exceed the safe limit.
	// That's why here we're splitting it into multiple posts.
	for i := 0; i < len(workflowRuns.Items); i += runsPerPost {
		items := workflowRuns.Items[i:funk.MinInt([]int{i + runsPerPost, len(workflowRuns.Items)}).(int)]

		attachments := make([]*model.SlackAttachment, len(items))

		for i, workflow := range items {
			attachment := util.BaseSlackAttachment()
			attachment.Title = "Workflow Insight for " + strings.Title(workflowName)

			attachment.Fields = []*model.SlackAttachmentField{
				{
					Short: true,
					Title: "ID",
					Value: workflow.Id,
				},
				{
					Short: true,
					Title: "Status",
					Value: workflow.Status,
				},
				{
					Short: true,
					Title: "Started At",
					Value: workflow.CreatedAt.Format(time.UnixDate),
				},
				{
					Short: true,
					Title: "Ended At",
					Value: workflow.StoppedAt.Format(time.UnixDate),
				},
				{
					Short: true,
					Title: "Duration",
					Value: humanize.RelTime(workflow.CreatedAt, workflow.StoppedAt, "", ""),
				},
				{
					Short: true,
					Title: "Credits Used",
					Value: fmt.Sprintf("%d", workflow.CreditsUsed),
				},
			}

			attachments[i] = attachment
		}

		post := &model.Post{
			UserId:    config.BotUserID,
			ChannelId: context.ChannelId,
		}

		model.ParseSlackAttachment(post, attachments)
		if _, err := config.Mattermost.CreatePost(post); err != nil {
			config.Mattermost.LogError(fmt.Sprintf("Failed to create post for recent workflows. ChannelID: %s, error: %s", context.ChannelId, err.Error()))
			return util.SendEphemeralCommandResponse("Failed to create post. Please try again later. If the problem persists, contact your system administrator.")
		}
	}

	return &model.CommandResponse{}, nil
}
