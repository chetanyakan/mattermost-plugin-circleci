package serializer

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

type CircleCIWebhookRequest struct {
	Status         string `json:"status"`
	BuildNum       string `json:"build_num"`
	RepoName       string `json:"repo_name"`
	Tag            string `json:"tag"`
	Commit         string `json:"commit"`
	BuildURL       string `json:"build_url"`
	CompareURL     string `json:"compare_url"`
	RepoURL        string `json:"repo_url"`
	OrgName        string `json:"org_name"`
	Branch         string `json:"branch"`
	Username       string `json:"username"`
	PullRequest    string `json:"pull_request"`
	PipelineNumber string `json:"pipeline_number"`
	JobName        string `json:"job_name"`
	WorkflowID     string `json:"workflow_id"`
}

func (r *CircleCIWebhookRequest) GetSubscription() Subscription {
	var vcs *VCS
	// TODO: Add support for Github Enterprise
	if strings.HasPrefix(r.RepoURL, "git@github.com") {
		vcs = DefaultVCSList[VCSTypeGithub]
	} else {
		vcs = DefaultVCSList[VCSTypeBitbucket]
	}

	s := Subscription{
		VCSType:  vcs.Type,
		BaseURL:  vcs.BaseURL,
		OrgName:  r.OrgName,
		RepoName: r.RepoName,
	}
	return s
}

func (r *CircleCIWebhookRequest) getSlackAttachmentFields() []*model.SlackAttachmentField {
	workflowText := r.PipelineNumber
	if strings.TrimSpace(workflowText) == "" {
		workflowText = "Visit Workflow"
	}

	slackAttachmentFields := []*model.SlackAttachmentField{
		{
			Title: "Project",
			Value: r.OrgName + "/" + r.RepoName,
			Short: false,
		},
		{
			Title: "Job Number",
			Value: fmt.Sprintf("[%s](%s)", r.BuildNum, r.BuildURL),
			Short: true,
		},
		{
			Title: "Triggered By",
			Value: "@" + r.Username,
			Short: true,
		},
		{
			Title: "Workflow",
			Value: fmt.Sprintf("[%s](%s)", workflowText, "https://circleci.com/workflow-run/"+r.WorkflowID),
			Short: true,
		},
	}

	if r.Branch != "" {
		slackAttachmentFields = append(slackAttachmentFields, &model.SlackAttachmentField{
			Title: "Branch",
			Value: r.Branch,
			Short: true,
		})
	}

	if r.Tag != "" {
		slackAttachmentFields = append(slackAttachmentFields, &model.SlackAttachmentField{
			Title: "Tag",
			Value: r.Tag,
			Short: true,
		})
	}

	return slackAttachmentFields
}

func (r *CircleCIWebhookRequest) GenerateFailurePost() *model.Post {
	if r == nil {
		return nil
	}

	attachment := &model.SlackAttachment{
		Color:    "#d10c20",
		Title:    fmt.Sprintf(":red_circle: A **%s** job has failed!", r.JobName),
		Fields:   r.getSlackAttachmentFields(),
		ThumbURL: config.BotIconURLFailed,
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
		Title:    fmt.Sprintf(":tada: A **%s** job has succeeded!", r.JobName),
		Fields:   r.getSlackAttachmentFields(),
		ThumbURL: config.BotIconURLSuccess,
	}

	post := &model.Post{
		UserId: config.BotUserID,
	}

	post.AddProp("override_icon_url", config.BotIconURLSuccess)
	model.ParseSlackAttachment(post, []*model.SlackAttachment{attachment})
	return post
}
