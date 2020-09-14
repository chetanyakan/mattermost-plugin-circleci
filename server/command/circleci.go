package command

import (
	"fmt"
	circleci2 "github.com/TomTucka/go-circleci/circleci"
	"github.com/antihax/optional"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
	"net/http"
	"strings"
	"time"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

const (
	invalidCommand = "Invalid command parameters. Please use `/circleci help` for more information."
	HeadTypeBranch = "branch"
	HeadTypeTag    = "tag"
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

	if err := config.Mattermost.KVSet(ctx.UserId+"_auth_token", []byte(authToken)); err != nil {
		config.Mattermost.LogError("Unable to save auth token to KVStore. Error: " + err.Error())
		return nil, err
	}

	return util.SendEphemeralCommandResponse(fmt.Sprintf("Successfully saved auth token."))
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

func executeListRecentBuilds(context *model.CommandArgs, args ...string) (*model.CommandResponse, *model.AppError) {
	authToken, appErr := config.Mattermost.KVGet(context.UserId + "_auth_token")
	if appErr != nil {
		return nil, appErr
	}
	if string(authToken) == "" {
		return util.SendEphemeralCommandResponse("Not connected. Please connect and try again later.")
	}
	client := util.GetCircleciClient(string(authToken))

	vcs, repo, workflow := args[0], args[1], args[2]
	builds, _, err := client.InsightsApi.GetProjectWorkflowRuns(nil, vcs+"/"+repo, workflow, utils.Yesterday(), utils.Yesterday().Add(2*24*time.Hour), nil)
	if err != nil {
		return util.SendEphemeralCommandResponse("Unable to connect to circleci. Make sure the auth token is still valid. Error: " + err.Error())
	}

	workflowCache := map[string]circleci2.Workflow{}
	pipelineCache := map[string]circleci2.Pipeline{}
	userCache := map[string]circleci2.User{}
	//attachments := []*model.SlackAttachment{}
	attachment := util.BaseSlackAttachment()

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
		ChannelId: context.ChannelId,
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	if _, err := config.Mattermost.CreatePost(post); err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Failed to create post for triggered workflow. ChannelID: %s, error: %s", context.ChannelId, err.Error()))
		return util.SendEphemeralCommandResponse("Failed to create post for triggered workflow. The build has been triggered though. You can view it in CircleCI.")
	}

	return &model.CommandResponse{}, nil
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

	vcsSlug, repo, headType, head := args[0], args[1], args[2], args[3]
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

	build, response, err := client.PipelineApi.TriggerPipeline(nil, vcsSlug+"/"+repo, &circleci2.PipelineApiTriggerPipelineOpts{
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
			message = fmt.Sprintf("Unable to trigger build. Either the specified workflow doesn't exist or the  auth token is not valid.")
		} else {
			message = fmt.Sprintf("Unable to trigger build.")
		}

		return util.SendEphemeralCommandResponse(message)
	}

	workflows, response, err := client.PipelineApi.ListWorkflowsByPipelineId(nil, build.Id, nil)
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
