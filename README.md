# mattermost-plugin-circleci

[![Build Status](https://img.shields.io/circleci/project/github/chetanyakan/mattermost-plugin-circleci/master)](https://circleci.com/gh/chetanyakan/mattermost-plugin-circleci)
[![Release](https://img.shields.io/github/v/release/chetanyakan/mattermost-plugin-circleci)](https://github.com/chetanyakan/mattermost-plugin-circleci/releases/latest)

### Platform & tools

- Make sure you have following components installed:

  - Go - v1.15 - [Getting Started](https://golang.org/doc/install)
    > **Note:** If you have installed Go to a custom location, make sure the `$GOROOT` variable is set properly. Refer [Installing to a custom location](https://golang.org/doc/install#install).
  - Make

## Setting up CircleCI

Set up CircleCI to run the build job for each branch and build-and-release for each tag.

1. Go to [CircleCI Dashboard](https://circleci.com/dashboard).
2. In the top left, you will find the Org switcher. Select your Organisation.
3. If this is your first project on CircleCI, go to the Projects page, click the **Add Projects** button, then click the **Set Up Project** button next to your project. You may also click **Start Building** to manually trigger your first build.
4. To manage GitHub releases using CircleCI, you need to add your github personal access token to your project's environment variables.
   - Follow the instructions [here](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line) to create a personal access token. For CircleCI releases, you would need the `repo` scope.
   - Add the environment variable to your project as `GITHUB_TOKEN` by following the instructions [here](https://circleci.com/docs/2.0/env-vars/#setting-an-environment-variable-in-a-project).
   
#### Usage Instructions

Bot User: CircleCI

Slash Commands:

- /connect
- /disconnect
- /me

Webhooks:

- For all finished builds (successful or failed)


# Mattermost CircleCI Plugin

A CircleCI plugin for Mattermost. Supports CircleCI SaaS.

## Table of Contents

 - [Audience](#audience)
 - [License](#license)
 - [About the GitHub Plugin](#about-the-github-plugin)
 - [Before You Start](#before-you-start)
 - [Configuration](#configuration)
 - [Using the Plugin](#using-the-plugin)
 - [Onboarding Your Users](#onboarding-your-users)
 - [Slash Commands](#slash-commands)
 - [Frequently Asked Questions](#frequently-asked-questions)
 - [Development](#development)

![GitHub plugin screenshot](images/circleci_mattermost.png)

## Audience

This guide is intended for Mattermost System Admins setting up the CircleCI plugin and Mattermost users who want information about the plugin functionality. For more information about contributing to this plugin, visit the [Development section](#development).

## Licensed

This repository is licensed under the [Apache 2.0 License](https://github.com/mattermost/mattermost-plugin-github/blob/master/LICENSE).

## About the CircleCI Plugin

The Mattermost CircleCI plugin uses a webhook to connect your CircleCI account to Mattermost to listen for incoming CircleCI events. Event notifications are published in the channels where they are subscribed. 

After your System Admin has [configured the CircleCI plugin](#configuration), run `/CircleCI connect` in a Mattermost channel to connect your Mattermost and CircleCI accounts.

Once connected, you'll have access to the following features:

* __Event Subscriptions__ - Ability to subscribe to build notifications for specified repositories.
* __Build__ - Ability to trigger build in CircleCI for a project. The build can be triggered for either a branch or a tag.
* __Recent Builds__ - View recent builds for a repository's workflow. For example, view recentr builds for `release` workflow.
* __Pipeline by Number__ - Get details of a pipeline by it's number. Each pipeline execution in CircleCI has a user-readable number which can be used for identifying a pipeline execution.  
* __Environment__ - Get a list of *masked* context variables available to in pipeline.
* __Project Insights__ - Get project insights on demand such as success rate, throughput, mean duration etc.
* __Workflow Insights__ - Get insights of a workflow's runs. This includs details such as execution time, credits used and status.

## Before You Start

This guide assumes:

- You have a CircleCI account.
- You're a Mattermost System Admin.
- You're running Mattermost v5.20 or higher.

## Configuration

Configuration is started in GitHub and completed in Mattermost. 

**Note:** If you're using GitHub Enterprise, replace all GitHub links below with your GitHub Enterprise URL.

### Step 1: Register an OAuth Application in GitHub

1. Go to https://github.com/settings/applications/new to register an OAuth app.
2. Set the following values:
   - **Application Name:** `Mattermost GitHub Plugin - <your company name>`
   - **Homepage URL:** `https://github.com/mattermost/mattermost-plugin-github`
   - **Authorization callback URL:** `https://your-mattermost-url.com/plugins/github/oauth/complete`, replacing `https://your-mattermost-url.com` with your Mattermost URL.
3. Submit.
4. Copy the **Client ID** and **Client Secret** in the resulting screen.
5. Go to **System Console > Plugins > GitHub** and enter the **GitHub OAuth Client ID** and **GitHub OAuth Client Secret** you copied in a previous step.
6. Hit **Save**.

### Step 2: Create a Webhook in GitHub

You must create a webhook for each organization you want to receive notifications for or subscribe to.

1. In **System Console > Plugins > GitHub**, generate a new value for **Webhook Secret**. Copy it, as you will use it in a later step.
2. Hit **Save** to save the secret.
3. Go to the **Settings** page of your GitHub organization you want to send notifications from, then select **Webhooks** in the sidebar.
4. Click **Add Webhook**.
5. Set the following values:
   - **Payload URL:** `https://your-mattermost-url.com/plugins/github/webhook`, replacing `https://your-mattermost-url.com` with your Mattermost URL.
   - **Content Type:** `application/json`
   - **Secret:** the webhook secret you copied previously.
6. Select **Let me select individual events** for "Which events would you like to trigger this webhook?".
7. Select the following events: `Branch or Tag creation`, `Branch or Tag deletion`, `Issue comments`, `Issues`, `Pull requests`, `Pull request review`, `Pull request review comments`, `Pushes`.
7. Hit **Add Webhook** to save it.

If you have multiple organizations, repeat the process starting from step 3 to create a webhook for each organization.

### Step 3: Configure the Plugin in Mattermost

If you have an existing Mattermost user account with the name `github`, the plugin will post using the `github` account but without a `BOT` tag.

To prevent this, either:

- Convert the `github` user to a bot account by running `mattermost user convert github --bot` in the CLI.

or

- If the user is an existing user account you want to preserve, change its username and restart the Mattermost server. Once restarted, the plugin will create a bot 
account with the name `github`.

**Note:** For `v0.9.0` and earlier of the GitHub plugin, instead of using bot accounts, set the username the plugin is attached to in **System Console > Plugins > GitHub**.

#### Generate a Key
  
Open **System Console > Plugins > GitHub** and do the following:

1. Generate a new value for **At Rest Encryption Key**.
2. (Optional) **GitHub Organization:** Lock the plugin to a single GitHub organization by setting this field to the name of your GitHub organization.
3. (Optional) **Enable Private Repositories:** Allow the plugin to receive notifications from private repositories by setting this value to `true`.
4. (**Enterprise Only**) **Enterprise Base URL** and **Enterprise Upload URL**: Set these values to your GitHub Enterprise URLs, e.g. `https://github.example.com`. The Base and Upload URLs are often the same. When enabled, existing users must reconnect their accounts to gain access to private repositories. Affected users will be notified by the plugin once private repositories are enabled.
5. Hit **Save**.
6. Go to **System Console > Plugins > Management** and click **Enable** to enable the GitHub plugin.

You're all set!

## Using the Plugin

Once configuration is complete, run the `/github connect` slash command from any channel within Mattermost to connect your Mattermost account with GitHub.

## Onboarding Your Users

When you’ve tested the plugin and confirmed it’s working, notify your team so they can connect their GitHub account to Mattermost and get started. Copy and paste the text below, edit it to suit your requirements, and send it out.

> Hi team, 

> We've set up the Mattermost GitHub plugin, so you can get notifications from GitHub in Mattermost. To get started, run the `/github connect` slash command from any channel within Mattermost to connect your Mattermost account with GitHub. Then, take a look at the [slash commands](#slash-commands) section for details about how to use the plugin.

## Slash Commands

* __Autocomplete slash commands__ - Explore all the available slash commands by typing `/` in the text input box - the autocomplete suggestions help by providing a format example in black text and a short description of the slash command in grey text. Visit the [executing commands](https://docs.mattermost.com/help/messaging/executing-commands.html) documentation for more details.
* __Subscribe to a respository__ - Use `/github subscribe` to subscribe a Mattermost channel to receive notifications for new pull requests, issues, branch creation, and more in a GitHub repository.

   - For instance, to post notifications for issues, issue comments, and pull requests matching the label `Help Wanted` from `mattermost/mattermost-server`, use:
   ```
   /github subscribe mattermost/mattermost-server issues,pulls,issue_comments,label:"Help Wanted"
   ```
  - The following flags are supported:
     - `--exclude-org-member`: events triggered by organization members will not be delivered. It will be locked to the organization provided in the plugin configuration and it will only work for users whose membership is public. Note that organization members and collaborators are not the same.
   
* __Get to do items__ - Use `/github todo` to get an ephemeral message with items to do in GitHub, including a list of unread messages and pull requests awaiting your review.
* __Update settings__ - Use `/github settings` to update your settings for notifications and daily reminders.
* __And more!__ - Run `/github help` to see what else the slash command can do.

## Frequently Asked Questions

### How do I connect a repository instead of an organization?

Set up your GitHub webhook from the repository instead of the organization. Notifications and subscriptions will then be sent only for repositories you create webhooks for. The reminder and `/github todo` will still search the whole organization, but only list items assigned to you.

### How do I send notifications when a certain label is applied?

Suppose you want to send notifications to a Mattermost channel when `Severity/Critical` label is applied to any issue in the `mattermost/mattermost-plugin-github` repository. Then, use this command to subscribe to these notifications:

```
/github subscribe mattermost/mattermost-plugin-github issues,label:"Severity/Critical"
```

### How do I share feedback on this plugin?

Feel free to create a GitHub issue or [join the GitHub Plugin channel on our community Mattermost instance](https://community-release.mattermost.com/core/channels/github-plugin) to discuss.

### How does the plugin save user data for each connected GitHub user?

GitHub user tokens are AES encrypted with an At Rest Encryption Key configured in the plugin's settings page. Once encrypted, the tokens are saved in the `PluginKeyValueStore` table in your Mattermost database.

## Development

This plugin contains both a server and web app portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/extend/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/extend/plugins/developer-setup/) for more information about developing and extending plugins.
