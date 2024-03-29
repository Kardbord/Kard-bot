
<div align="center">
<img src="Robo_cat.png" alt="drawing" width="250"/>
</div>

# Kard-bot

[![Build](https://github.com/Kardbord/Kard-bot/actions/workflows/go.yml/badge.svg)](https://github.com/Kardbord/Kard-bot/actions/workflows/go.yml)
[![CodeQL](https://github.com/Kardbord/Kard-bot/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/Kardbord/Kard-bot/actions/workflows/codeql-analysis.yml)
[![Docker Image](https://github.com/Kardbord/Kard-bot/actions/workflows/release.yml/badge.svg)](https://github.com/Kardbord/Kard-bot/actions/workflows/release.yml)
[![Image Efficiency](https://github.com/Kardbord/Kard-bot/actions/workflows/image-dive.yml/badge.svg)](https://github.com/Kardbord/Kard-bot/actions/workflows/image-dive.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Kardbord/Kard-bot)](https://goreportcard.com/report/github.com/Kardbord/Kard-bot)

A discord bot destined for greatness.

## Table of Contents

- [Kard-bot](#kard-bot)
- [Features](#features)
- [User Guide](#user-guide)
- [Hosting Installation](#hosting-installation)
  - [Host using Docker](#host-using-docker)
  - [Precompiled binaries](#precompiled-binaries)
  - [Building from source](#building-from-source)
- [General Notes](#general-notes)
- [References](#references)
  - [Discord API Wrappers](#discord-api-wrappers)
  - [Documentation](#documentation)
  - [Tutorials](#tutorials)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

## Features

- [x] Respond to user greetings and goodbyes
- [x] DnD dice roller
- [x] Copy-pastas
- [x] Reddit Roulette
- [x] [Uwu-ifier](https://lingojam.com/uwu-ify)
- [x] Print out a help message
- [x] Let users know when it is Wednesday
- [x] Daily compliments DM'd to users who opt in
- [x] Creepy DMs sent to users who opt in
- [x] Provide "odds" that a user specified event will occur
- [x] Build memes from provided templates and user provided text
- [x] Generate a story from a user's prompt
- [x] Allow server admins to generate and edit a role selection menu  
- [x] Allow users to create embeds
- [x] Madlibs
- [x] Server Clock
- [x] User polls
- [x] AI text-to-image generation using [DALL·E 2](https://openai.com/dall-e-2/)
- [ ] Inform users when Kard-bot is updated
- [ ] Mock certain questions or phrases
- [ ] "Quack" any time a user types an expletive
- [ ] Subscribe to social media accounts (maybe a webhook would be more appropriate?)
- [ ] Play music via youtube à la [rythm bot](https://rythm.fm/)
- [ ] Configurably replace words with other words
- [ ] Search and link to DnD wiki articles
- [ ] Allow users to query Google; provide direct links to top results and a link to all results
- [ ] Youtube Roulette

## User Guide

This bot is not publicly hosted anywhere. If you want to use it, you can always try [hosting it yourself](#hosting-installation)! :)

## Hosting Installation

Hosting this bot requires a Discord Bot Token. You can generate one by visiting the [Discord Developer Portal](https://discord.com/developers/applications),
and then creating a new application with an accompanying bot. Give the bot its needed permissions in the OAuth2 section **(Be sure to tick the "applications.commands" box!)**, and then invite it to your server(s) using the link that is generated for you.

Now that the bot is invited, you should see it as an offline user in your server. Now you only need to start the bot backend to bring it online! You have three options:

1. [Using the provided Docker images](#host-using-docker)
2. [Using a precompiled binary](#precompiled-binaries)
3. [Building from source](#building-from-source)

Whichever you choose, you'll want to edit the included `.env` file to include the bot token you generated earlier. You'll also want to
add any additional API tokens for functionality you plan on using, and set the time zone by setting the `TZ` variable.

```shell
KARDBOT_TOKEN="Your bot token here"
TZ="Your time zone here ex: America/Boise"
```

### Host using Docker

**Prerequisites**

- [Docker](https://www.docker.com/get-started)
- [docker-compose](https://docs.docker.com/compose/install/)

**Instructions**

Head over to the [Releases](https://github.com/Kardbord/Kard-bot/releases) page and download the `kardbot-<TAG>.tar.gz` tarball for the desired release.
These tarballs contain everything needed to get an instance of the bot up and running, provided that the host machine has internet access.
Untar it on the host machine.

With your token in place and your config updated, you can simply run `docker-compose up -d` from the untarred directory to get your bot started!
The Docker daemon will automatically download the needed docker image from [Docker Hub](https://hub.docker.com/r/tkvarfordt/kardbot/tags) or the
[GitHub Container Registry](https://github.com/Kardbord/Kard-bot/pkgs/container/kard-bot).
To check the status of the docker container, you can use `docker ps -a` or `docker logs <CONTAINER-NAME>`.

### Precompiled Binaries

**Prerequisites**

- Know your operating system and architecture.

**Instructions**

Head over to the [Releases](https://github.com/Kardbord/Kard-bot/releases) page and download the appropriate tarball for your operating system and architecture.
Untar it on the host machine.

With your token in place and your config updated, you can simply run the Kard-bot binary to bring it to life!
For a more robust running solution, consider creating a [systemd service](https://docs.fedoraproject.org/en-US/quick-docs/understanding-and-administering-systemd/#creating-new-systemd-services) or [using the provided Docker image](#host-using-docker).

### Building from source

**Prerequisites**

- The [Go](https://golang.org/) 1.18 runtime or later.

**Instructions**

Assuming you have the [Go](https://golang.org/) runtime installed, you can install Kard-bot with a simple set of shell commands.

```shell
go get github.com/Kardbord/Kard-bot
cd $GOPATH/src/github.com/Kardbord/Kard-bot
go build
```

With your token in place and your config updated, you can simply run the Kard-bot binary to bring it to life!
For a more robust running solution, consider creating a [systemd service](https://docs.fedoraproject.org/en-US/quick-docs/understanding-and-administering-systemd/#creating-new-systemd-services) or [using the provided Docker image](#host-using-docker).

## General Notes

Some commands are restricted so that only the bot owner can run them. The bot owner is specified by the `KARDBOT_OWNER_ID` environment variable.
It can be set in the same manner as the `KARDBOT_TOKEN` variable. Its value should be the user ID of the bot owner. Note that this is not the same
as the owner's username. The user ID is a unique ID assigned by Discord. You can retrieve it by enabling developer mode in your Discord client, right
clicking a user, and selecting "Copy ID".

## References

Useful resources for writing a Discord bot.

### Discord API Wrappers

- [discordpy](https://github.com/Rapptz/discord.py)
- [discordgo](https://github.com/bwmarrin/discordgo)
- [others](https://discordapi.com/unofficial/comparison.html)

### Documentation

- [Discord Developer Portal](https://discord.com/developers/docs/intro)

### Tutorials

- [Real Python Tutorial](https://realpython.com/how-to-make-a-discord-bot-python/)
