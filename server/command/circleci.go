package command

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	circleci2 "github.com/TomTucka/go-circleci/circleci"
	"github.com/antihax/optional"
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
	Execute: executeAddVCS,
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
	Execute: executeAddVCS,
	AutocompleteData: &model.AutocompleteData{
		Trigger:  "list vcs",
		HelpText: "List all available VCS.",
	},
}

var CircleCICommandHandler = Handler{
	Command: &model.Command{
		Trigger:          config.CommandPrefix,
		Description:      "Integration with CircleCI.",
		DisplayName:      "CircleCI",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: connect <token>, disconnect, me, projects, build, recent builds.",
		AutoCompleteHint: "[command]",
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
				commandAddVCS.AutocompleteData,
				commandDeleteVCS.AutocompleteData,
				commandListVCS.AutocompleteData,
			},
		},
	},
	handlers: map[string]HandlerFunc{
		"connect":            executeConnect,
		"disconnect":         executeDisconnect,
		"subscribe":          executeSubscribe,
		"unsubscribe":        executeUnsubscribe,
		"list/subscriptions": executeListSubscriptions,
		"build":              executeBuild,
		"recent/builds":      executeListRecentBuilds,
		"add/vcs":            executeAddVCS,
		"delete/vcs":         executeDeleteVCS,
		"list/vcs":           executeListVCS,
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

func executeConnect(ctx *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	// we need the auth token
	if len(args) < 1 {
		return util.SendEphemeralCommandResponse("Please specify the auth token.")
	}

	authToken := args[0]
	conf := circleci2.NewConfiguration()
	conf.AddDefaultHeader("Circle-Token", authToken)

	if err := config.Mattermost.KVSet(ctx.UserId+"_auth_token", []byte(authToken)); err != nil {
		config.Mattermost.LogError("Unable to save auth token to KVStore. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse("Successfully saved auth token.")
}

func executeDisconnect(ctx *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(ctx.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}

	if err := config.Mattermost.KVDelete(ctx.UserId + "_auth_token"); err != nil {
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

	vcs, org, repo, workflow := args[0], args[1], args[2], args[4]
	builds, resp, err := client.InsightsApi.GetProjectWorkflowRuns(context.TODO(), vcs+"/"+org+"/"+repo, workflow, utils.Yesterday(), utils.Yesterday().Add(2*24*time.Hour), nil)

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
				Value: workflow.PipelineNumber,
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

	vcsSlug, org, repo, headType, head := args[0], args[1], args[2], args[3], args[4]
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

	build, response, err := client.PipelineApi.TriggerPipeline(context.TODO(), vcsSlug+"/"+org+"/"+repo, &circleci2.PipelineApiTriggerPipelineOpts{
		Body: optional.NewInterface(body),
	})

	if err != nil {
		config.Mattermost.LogError(fmt.Sprintf(
			"Failed to trigger build. VCS slug: %s, repo: %s, head type: %s, head: %s, error: %s, response: %v",
			vcsSlug,
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
			vcsSlug,
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
			Value: fmt.Sprintf("https://app.circleci.com/pipelines/%s/%d/workflows/%s", vcsSlug+"/"+repo, build.Number, workflow.Id),
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
	if len(args) < 2 {
		return util.SendEphemeralCommandResponse("Invalid number of arguments. Use this command as `/cirecleci add vcs [alias] [base URL]`")
	}

	alias, baseURL := args[0], args[1]

	existingVCS, err := service.GetVCS(alias)
	if err != nil {
		return util.SendEphemeralCommandResponse("Failed to check for existing VCS with same alias. Please try again later. If the problem persists, contact your system administrator.")
	}

	if existingVCS != nil {
		return util.SendEphemeralCommandResponse(fmt.Sprintf("Another VCS existis with the same alias. Please delete existing VCS first if you want to update it. Alias: `%s`, base URL: `%s`", existingVCS.Alias, existingVCS.BaseURL))
	}

	vcs := &serializer.VCS{
		Alias:   alias,
		BaseURL: baseURL,
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

	message := "Available VCS -\n\n| No.  | Alias  | Base URL |\n|:------------|:------------|:------------|\n"
	for i, vcs := range vcsList {
		message += fmt.Sprintf("|%d|%s|%s|\n", i+1, vcs.Alias, vcs.BaseURL)
	}

	_, _ = config.Mattermost.CreatePost(&model.Post{
		UserId:    config.BotUserID,
		ChannelId: context.ChannelId,
		Message:   message,
	})

	return &model.CommandResponse{}, nil
}
