<div align="center">
<img src="Robo_cat.png" alt="drawing" width="250"/>
</div>

# Kard-bot
A discord bot destined for greatness.

# Features
- [x] Respond to user greetings and goodbyes
- [x] DnD dice roller
- [x] Copy-pastas when requested and/or when triggered
- [x] Reddit Roulette
- [ ] [Uwu-ifier](https://lingojam.com/uwu-ify)
- [ ] [Creepy asterisks](https://www.reddit.com/r/creepyasterisks/)
- [ ] Mock certain questions or phrases
- [ ] "Quack" any time a user types an expletive
- [ ] Subscribe to social media accounts (maybe a webhook would be more appropriate?)
- [ ] Play music via youtube à la [rythm bot](https://rythm.fm/)
- [ ] Configurably replace words with other words

# Installation
This bot is not currently hosted anywhere. If you want to use it, you can always try [hosting it yourself](#hosting-installation)! :)

# Hosting Installation
Hosting this bot requires a Discord Bot Token. You can generate one by visiting the [Discord Developer Portal](https://discord.com/developers/),
and then creating a new application with an accompanying bot. Give the bot its needed permissions in the OAuth2 section **(Be sure to tick the "applications.commands" box!)**, and then invite it
to your server(s) using the link that is generated for you.

For now, you will have to build the Kard-bot binary yourself. Free-time permitting, I may provide a released version or docker image. 

Assuming you have the [Go](https://golang.org/) runtime installed, you can install Kard-bot with a simple shell command.

```shell
go install github.com/TannerKvarfordt/Kard-bot
```

In order to authenticate with Discord, Kard-bot looks for the `KARDBOT_TOKEN` environment variable. 
It is recommended to place that variable in a `.env` file at the root of the project. Note that existing
environment variables take precedence over anything in the `.env` file.

```shell
KARDBOT_TOKEN="Your bot token here"
```

With your token in place, you can simply run the Kard-bot binary to bring it to life!
For a more robust running solution, consider creating a [systemd service](https://docs.fedoraproject.org/en-US/quick-docs/understanding-and-administering-systemd/#creating-new-systemd-services) or something similar to suit your needs.

Some commands are restricted so that only the bot owner can run them. The bot owner is specified by the `KARDBOT_OWNER_ID` environment variable.
It can be set in the same manner as the `KARDBOT_TOKEN` variable. Its value should be the user ID of the bot owner. Note that this is not the same
as the owner's username. The user ID is a unique ID assigned by Discord. You can retrieve it by enabling developer mode in your Discord client, right
clicking a user, and selecting "Copy ID".

# References
Useful resources for writing a Discord bot.
### Discord API Wrappers
- [discordpy](https://github.com/Rapptz/discord.py)
- [discordgo](https://github.com/bwmarrin/discordgo)
  - ~~[dgc](https://github.com/lus/dgc)~~ will be deprecated April 2022 due to Discord API updates :(
- [others](https://discordapi.com/unofficial/comparison.html)

### Documentation
- [Discord Developer Portal](https://discord.com/developers/docs/intro)

### Tutorials
- [Real Python Tutorial](https://realpython.com/how-to-make-a-discord-bot-python/)

