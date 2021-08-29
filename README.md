# Kard-bot
A discord bot destined for greatness.

# Features
- [ ] DnD dice roller
- [ ] Copy-pastas when requested and/or when triggered
- [ ] [Uwu-ifier](https://lingojam.com/uwu-ify)
- [ ] [Creepy asterisks](https://www.reddit.com/r/creepyasterisks/)
- [ ] Mock certain questions or phrases
- [ ] "Quack" any time a user types an expletive
- [ ] Subscribe to social media accounts (maybe a webhook would be more appropriate?)
- [ ] Play music via youtube à la [rythm bot](https://rythm.fm/)
- [ ] Configurably replace words with other words

# Discord Installation
This bot is not currently hosted anywhere. If you want to use it, you can always try [hosting it yourself](#hosting-installation)! :)

# Hosting Installation
Hosting this bot requires a Discord Bot Token. You can generate one by visiting the [Discord Developer Portal](https://discord.com/developers/),
and then creating a new application with an accompanying bot. Be sure to give the bot its needed permissions in the OAuth2 section, and then invite it
to your server(s) using the link that is generated for you.

For now, you will have to build the Kard-bot binary yourself. Free-time permitting, I may provide a released version or docker image. 

Assuming you have the [Go](https://golang.org/) runtime installed, you can install Kard-bot with a simple shell command.

```shell
go install github.com/TannerKvarfordt/Kard-bot
```

In order to authenticate with Discord, Kard-bot looks for the `DISCORD_BOT_TOKEN` environment variable. 
It is recommended to place that variable in a `.env` file at the root of the project. Note that existing
environment variables take precedence over anything in the `.env` file.

```shell
DISCORD_BOT_TOKEN="Your bot token here"
```

With your token in place, you can simply run the Kard-bot binary to bring it to life!
For a more robust running solution, consider creating a [systemd service](https://docs.fedoraproject.org/en-US/quick-docs/understanding-and-administering-systemd/#creating-new-systemd-services) or something similar to suit your needs.


# References
Useful resources for writing a Discord bot.
### Discord API Wrappers
- [discordpy](https://github.com/Rapptz/discord.py)
- [discordgo](https://github.com/bwmarrin/discordgo)

### Documentation
- [Discord Developer Portal](https://discord.com/developers/docs/intro)

### Tutorials
- [Real Python Tutorial](https://realpython.com/how-to-make-a-discord-bot-python/)

