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
	Status      string `json:"status"`
	BuildNum    string `json:"build_num"`
	RepoName    string `json:"repo_name"`
	Tag         string `json:"tag"`
	Commit      string `json:"commit"`
	BuildURL    string `json:"build_url"`
	CompareURL  string `json:"compare_url"`
	VCSType     string `json:"vcs_type"`
	VCSBaseURL  string `json:"vcs_base_url"`
	OrgName     string `json:"org_name"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	PullRequest string `json:"pull_request"` // TODO: multiple PRs
	Job         string `json:"job"`
	WorkflowID  string `json:"workflow_id"`
}

func (r *CircleCIWebhookRequest) GetSubscription() Subscription {
	s := Subscription{
		VCSType:  r.VCSType,
		BaseURL:  r.VCSBaseURL,
		OrgName:  r.OrgName,
		RepoName: r.RepoName,
	}
	return s
}

func (r *CircleCIWebhookRequest) getSlackAttachmentFields() []*model.SlackAttachmentField {
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

	return slackAttachmentFields
}

func (r *CircleCIWebhookRequest) GenerateFailurePost() *model.Post {
	if r == nil {
		return nil
	}

	attachment := &model.SlackAttachment{
		Color:    "#d10c20",
		Title:    fmt.Sprintf("Oops. Build [%s](%s) failed.", r.BuildNum, r.BuildURL),
		Fields:   r.getSlackAttachmentFields(),
		ThumbURL: config.WorkflowFailedIconURL,
	}

	post := &model.Post{
		UserId: config.BotUserID,
	}

	post.AddProp("override_icon_url", config.BotIconURLFailed)
	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	return post
}

func (r *CircleCIWebhookRequest) GenerateSuccessPost() *model.Post {
	if r == nil {
		return nil
	}

	attachment := &model.SlackAttachment{
		Color:    "#41aa58",
		Title:    fmt.Sprintf("Build [%s](%s) passed.", r.BuildNum, r.BuildURL),
		Fields:   r.getSlackAttachmentFields(),
		ThumbURL: config.WorkflowSuccessIconURL,
	}

	post := &model.Post{
		UserId: config.BotUserID,
	}

	post.AddProp("override_icon_url", config.BotIconURLSuccess)
	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	return post
}
