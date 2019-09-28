package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-server/model"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
)

func CircleCIConnect(authToken string) {
	resp, err := http.Get("https://circleci.com/api/v1.1/me?circle-token=" + authToken)
	if err != nil {
		config.Mattermost.LogError("Unable to get user for API token")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		config.Mattermost.LogError("Unable to read response body")
	}
	config.Mattermost.LogInfo("", "Body", string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		config.Mattermost.LogError("Unable to decode response body", "Error", err.Error())
		return
	}
	fmt.Println(result)

	attachment := &model.SlackAttachment{
		Fallback: result["name"].(string),
		Color:    "#7FC1EE",
		Title:    result["name"].(string),
		ThumbURL: result["avatar_url"].(string),
	}
	attachment.Pretext = fmt.Sprintf("Initiated by CircleCI user: %s", result["name"].(string))
	post := &model.Post{
		UserId:    config.GetConfig().BotUserID,
		ChannelId: "", // TODO:  Set  Channel ID from command request
		Type:      model.POST_DEFAULT,
		Props: map[string]interface{}{
			"attachments": []*model.SlackAttachment{attachment},
		},
	}

	if _, appErr := config.Mattermost.CreatePost(post); appErr != nil {
		config.Mattermost.LogError("Could not create a post", "err", appErr.Error())
		return
	}
}
