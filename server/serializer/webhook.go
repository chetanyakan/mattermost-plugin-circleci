package serializer

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

var colors = map[string]string{
	"test":    "#AAAAAA",
	"success": "#41aa58",
	"failure": "#d10c20",
}

var texts = map[string]string{
	"test":    "Hello from CircleCI.",
	"success": "CircleCI build success.",
	"failure": "CircleCI build failure.",
}

type CircleCIWebhookRequest struct {
	ChannelID   string `json:"channel_id"`
	Status      string `json:"status"`
	BuildNum    string `json:"build_num"`
	RepoName    string `json:"repo_name"`
	Tag         string `json:"tag"`
	Commit      string `json:"commit"`
	BuildURL    string `json:"build_url"`
	CompareURL  string `json:"compare_url"`
	OrgName     string `json:"org_name"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	PullRequest string `json:"pull_request"` // TODO: multiple PRs
	Job         string `json:"job"`
	WorkflowID  string `json:"workflow_id"`
}

func (r *CircleCIWebhookRequest) GenerateFailurePost() *model.Post {
	if r == nil {
		return nil
	}

	slackAttachmentFields := []*model.SlackAttachmentField{
		{
			Title: "Tag",
			Value: r.Tag,
			Short: true,
		},
		{
			Title: "Branch",
			Value: r.Branch,
			Short: true,
		},
		{
			Title: "Triggered By",
			Value: r.Username,
			Short: true,
		},
		{
			Title: "Commit",
			Value: r.Commit,
			Short: true,
		},
		{
			Title: "Organisation/User",
			Value: r.OrgName,
			Short: true,
		},
		{
			Title: "Repo",
			Value: r.RepoName,
			Short: true,
		},
		{
			Title: "Job",
			Value: r.Job,
			Short: true,
		},
		{
			Title: "Workflow ID",
			Value: r.WorkflowID,
			Short: true,
		},
	}

	attachment := &model.SlackAttachment{
		Color:    "#d10c20",
		Title:    fmt.Sprintf("Oops. Build [%s](%s) failed.", r.BuildNum, r.BuildURL),
		Fields:   slackAttachmentFields,
		ThumbURL: "/plugins/" + config.PluginName + "/static/icon_failed.png",
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: r.ChannelID,
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	return post
}

func (r *CircleCIWebhookRequest) GenerateSuccessPost() *model.Post {
	if r == nil {
		return nil
	}

	slackAttachmentFields := []*model.SlackAttachmentField{
		{
			Title: "Tag",
			Value: r.Tag,
			Short: true,
		},
		{
			Title: "Branch",
			Value: r.Branch,
			Short: true,
		},
		{
			Title: "Triggered By",
			Value: r.Username,
			Short: true,
		},
		{
			Title: "Commit",
			Value: r.Commit,
			Short: true,
		},
		{
			Title: "Organisation/User",
			Value: r.OrgName,
			Short: true,
		},
		{
			Title: "Repo",
			Value: r.RepoName,
			Short: true,
		},
		{
			Title: "Job",
			Value: r.Job,
			Short: true,
		},
		{
			Title: "Workflow ID",
			Value: r.WorkflowID,
			Short: true,
		},
	}

	attachment := &model.SlackAttachment{
		Color:    "#41aa58",
		Title:    fmt.Sprintf("Build [%s](%s) passed.", r.BuildNum, r.BuildURL),
		Fields:   slackAttachmentFields,
		ThumbURL: "/plugins/" + config.PluginName + "/static/icon_success.png",
	}

	post := &model.Post{
		UserId:    config.BotUserID,
		ChannelId: r.ChannelID,
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	return post
}
