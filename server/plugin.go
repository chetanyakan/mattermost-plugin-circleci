package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/pkg/errors"

	"github.com/chetanyakan/mattermost-plugin-circleci/server/command"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/config"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/controller"
	"github.com/chetanyakan/mattermost-plugin-circleci/server/util"
)

const (
	botUserName    = "circleci"
	botDisplayName = "CircleCI"
	botDescription = "Created by the CircleCI Plugin."
)

type Plugin struct {
	plugin.MattermostPlugin

	handler http.Handler
}

func (p *Plugin) OnActivate() error {
	config.Mattermost = p.API

	if err := p.setupStaticFileServer(); err != nil {
		p.API.LogError(err.Error())
		return err
	}

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	if err := p.registerCommands(); err != nil {
		config.Mattermost.LogError(err.Error())
		return err
	}

	return nil
}

func (p *Plugin) initBotUser() (string, error) {
	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    botUserName,
		DisplayName: botDisplayName,
		Description: botDescription,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to ensure bot")
	}

	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		return "", errors.Wrap(err, "failed to get bundle path")
	}

	fmt.Println("BundlePath: " + bundlePath)
	// profileImage, err := ioutil.ReadFile(filepath.Join(bundlePath, "assets", "circleci.png"))
	// if err != nil {
	// 	return "", errors.Wrap(err, "failed to read profile image")
	// }
	//
	// if appErr := p.API.SetProfileImage(botUserID, profileImage); appErr != nil {
	// 	return "", errors.Wrap(appErr, "failed to set profile image")
	// }

	return botUserID, nil
}

func (p *Plugin) setupStaticFileServer() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	p.handler = http.FileServer(http.Dir(filepath.Dir(exe) + config.ServerExeToWebappRootPath))
	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	if config.Mattermost != nil {
		var configuration config.Configuration

		if err := config.Mattermost.LoadPluginConfiguration(&configuration); err != nil {
			config.Mattermost.LogError("Error in LoadPluginConfiguration: " + err.Error())
			return err
		}

		if err := configuration.ProcessConfiguration(); err != nil {
			config.Mattermost.LogError("Error in ProcessConfiguration: " + err.Error())
			return err
		}

		if err := configuration.IsValid(); err != nil {
			config.Mattermost.LogError("Error in Validating Configuration: " + err.Error())
			return err
		}

		botUserID, err := p.initBotUser()
		if err != nil {
			config.Mattermost.LogError(err.Error())
			return err
		}

		configuration.BotUserID = botUserID
		config.SetConfig(&configuration)
	}
	return nil
}

func (p *Plugin) registerCommands() error {
	if err := config.Mattermost.RegisterCommand(command.Master().Command); err != nil {
		config.Mattermost.LogError("Cound't register command", err, map[string]interface{}{"command": command.Master().Command.Trigger})
		return err
	}

	return nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split, argErr := util.SplitArgs(args.Command)
	if argErr != nil {
		return util.SendEphemeralText(argErr.Error())
	}

	cmdName := split[0]
	var params []string

	if len(split) > 1 {
		params = split[1:]
	}

	if cmdName != "/"+command.Master().Command.Trigger {
		return nil, &model.AppError{Message: "Unknown command: [" + cmdName + "] encountered"}
	}

	context := p.prepareContext(args)
	if response, err := command.Master().Validate(params, context); response != nil {
		return response, err
	}

	config.Mattermost.LogInfo("Executing command: " + cmdName + " with params: [" + strings.Join(params, ", ") + "]")
	return command.Master().Execute(params, context)
}

func (p *Plugin) prepareContext(args *model.CommandArgs) command.Context {
	return command.Context{
		CommandArgs: args,
		Props:       make(map[string]interface{}),
	}
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	conf := config.GetConfig()

	if err := conf.IsValid(); err != nil {
		config.Mattermost.LogError("This plugin is not configured: " + err.Error())
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

	endpoint := controller.GetEndpoint(r)
	if endpoint == nil {
		p.handler.ServeHTTP(w, r)
		return
	}

	if endpoint.RequiresAuth && !controller.Authenticated(w, r) {
		config.Mattermost.LogError(fmt.Sprintf("Endpoint: %s '%s' requires Authentication.", endpoint.Method, endpoint.Path))
		http.Error(w, "This endpoint requires authentication.", http.StatusForbidden)
		return
	}

	if err := endpoint.Execute(w, r); err != nil {
		config.Mattermost.LogError(fmt.Sprintf("Processing: %s '%s'.", r.Method, r.URL.String()), "Error", err.Error())
	}
}
func main() {
	plugin.ClientMain(&Plugin{})
}
