package kardbot

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/TannerKvarfordt/Kard-bot/kardbot/dg_helpers"
	"github.com/bwmarrin/discordgo"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/vartanbeno/go-reddit/v2/reddit"

	log "github.com/sirupsen/logrus"
)

var (
	redditClient = func() *reddit.Client { return nil }
	redditCtx    = func() context.Context { return context.Background() }
)

const (
	redditRouletteSubCmdSFW  string = "sfw"
	redditRouletteSubCmdNSFW string = "nsfw"
	redditRouletteSubCmdAny  string = "any"

	// reddit API returns a max of 100 posts at a time
	redditMaxPostsPerRequest int = 100
)

func init() {
	client, err := reddit.NewReadonlyClient()
	if err != nil {
		log.Fatal("Could not initialize reddit client")
	}
	redditClient = func() *reddit.Client { return client }
}

func redditRoulette(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: move this check into a helper function
	if authorID, err := getInteractionCreateAuthorID(i); err == nil {
		if authorID == s.State.User.ID {
			log.Trace("Ignoring message from self")
			return
		}
	} else {
		log.Error(err)
		return
	}

	var post *reddit.Post
	var err error
	switch i.ApplicationCommandData().Options[0].Name {
	case redditRouletteSubCmdAny:
		post, err = getRandomRedditPost(nil)
	case redditRouletteSubCmdSFW:
		nsfw := false
		post, err = getRandomRedditPost(&nsfw)
	case redditRouletteSubCmdNSFW:
		nsfw := true
		post, err = getRandomRedditPost(&nsfw)
	default:
		log.Error("Reached unreachable case...")
	}
	if err != nil {
		log.Error(err)
		return
	}
	if post == nil {
		log.Error("post is nil")
		return
	}

	embed, err := buildRedditPostEmbed(post)
	if err != nil {
		log.Error(err)
		return
	}
	if embed == nil {
		log.Error("embed is nil")
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Error(err)
	}
}

func buildRedditPostEmbed(post *reddit.Post) (*discordgo.MessageEmbed, error) {
	if post == nil {
		return nil, fmt.Errorf("post is nil")
	}
	hexColor, _ := strconv.ParseInt(strings.Replace(colorful.FastHappyColor().Hex(), "#", "", -1), 16, 32)
	var voteEmoji = ""
	if post.Score > 0 {
		voteEmoji = "👍"
	} else {
		voteEmoji = "👎"
	}

	embed := dg_helpers.NewEmbed().
		SetTitle(post.Title).
		SetDescription(fmt.Sprintf("%s by u/%s", post.SubredditNamePrefixed, post.Author)).
		SetFooter(fmt.Sprintf("%s %d (%d%% upvoted) 💬 %d", voteEmoji, post.Score, int(post.UpvoteRatio*100), post.NumberOfComments)).
		SetColor(int(hexColor)).
		SetURL(fmt.Sprintf("https://www.reddit.com%s", post.Permalink)).
		SetType(discordgo.EmbedTypeRich).
		Truncate()

	if post.Body != "" {
		embed.AddField("-", post.Body)
	}

	embedRedditMedia(post, embed)
	return embed.MessageEmbed, nil
}

// Compile these regexps at init time so we don't have to
// do it every time.
var (
	isImageRegex = func() *regexp.Regexp { return nil }
	isGifvRegex  = func() *regexp.Regexp { return nil }
	isVideoRegex = func() *regexp.Regexp { return nil }
)

func init() {
	imgRegex := regexp.MustCompile(`^([^\s]+(\.(?i)(jpg|jpeg|png|gif))$)`)
	isImageRegex = func() *regexp.Regexp { return imgRegex }

	gifvRegex := regexp.MustCompile(`^([^\s]+(\.(?i)(gifv))$)`)
	isGifvRegex = func() *regexp.Regexp { return gifvRegex }

	videoRegex := regexp.MustCompile(`^([^\s]+(\.(?i)(webm|mp4|wav))$)`)
	isVideoRegex = func() *regexp.Regexp { return videoRegex }
}

func embedRedditMedia(post *reddit.Post, embed *dg_helpers.Embed) {
	if post == nil {
		log.Error("post is nil")
		return
	}
	if embed == nil {
		log.Error("embed is nil")
		return
	}
	if !isHTTPS(post.URL) {
		log.Warn("link is not https, won't embed", post.URL)
		return
	}

	if isImageRegex().MatchString(post.URL) {
		log.Debug("Embedding image", post.URL)
		embed.SetImage(post.URL)
	} else if isGifvRegex().MatchString(post.URL) {
		// TODO: should gifv's be embedded as videos?
		log.Debug("Embedding GIFV", post.URL)
		embed.SetImage(post.URL)
	} else if isVideoRegex().MatchString(post.URL) {
		log.Debug("Embedding video", post.URL)
		embed.SetVideo(post.URL)
	}
}

func getRandomSubredditSFW() (*reddit.Subreddit, error) {
	sub, _, err := redditClient().Subreddit.Random(redditCtx())
	return sub, err
}

func getRandomSubredditNSFW() (*reddit.Subreddit, error) {
	sub, _, err := redditClient().Subreddit.RandomNSFW(redditCtx())
	return sub, err
}

func getRandomSubreddit() (*reddit.Subreddit, error) {
	if RandomBoolean() {
		return getRandomSubredditNSFW()
	}
	return getRandomSubredditSFW()
}

func getTopPosts(count int, subreddit string) ([]*reddit.Post, error) {
	if count > redditMaxPostsPerRequest {
		log.Warnf("Request for %d exceeds max posts per request, retrieving only %d", count, redditMaxPostsPerRequest)
		count = redditMaxPostsPerRequest
	}
	posts, _, err := redditClient().Subreddit.TopPosts(redditCtx(), subreddit, &reddit.ListPostOptions{
		ListOptions: reddit.ListOptions{
			Limit: count,
		},
	})
	return posts, err
}

// Attempts to retrieve a random reddit post.
// Ensures post is marked sfw or nsfw based on
// the 'nsfw' parameter. If the 'nsfw' parameter
// is nil, no check is done.
func getRandomRedditPost(nsfw *bool) (*reddit.Post, error) {
	if nsfw == nil {
		post, _, err := redditClient().Post.Random(redditCtx())
		return post.Post, err
	}

	// No way as far as I can tell to get a random post that is
	// guaranteed SFW or NSFW, but you can do so for subreddits.
	// Get a random SFW or NSFW sub, retrieve the top posts, then
	// shuffle them and iterate until a post is found matching the
	// sfw/nsfw criteria.
	var sub *reddit.Subreddit
	var err error
	if *nsfw {
		sub, err = getRandomSubredditNSFW()
	} else {
		sub, err = getRandomSubredditSFW()
	}
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, fmt.Errorf("nil subreddit retrieved")
	}

	posts, err := getTopPosts(redditMaxPostsPerRequest, sub.Name)
	if err != nil {
		return nil, err
	}
	if len(posts) < 1 {
		return nil, fmt.Errorf("no posts retrieved")
	}

	rand.Shuffle(len(posts), func(i, j int) {
		posts[i], posts[j] = posts[j], posts[i]
	})

	for _, post := range posts {
		if post.NSFW == *nsfw {
			return post, nil
		}
	}
	return nil, fmt.Errorf("no posts found matching nsfw=%t", *nsfw)
}
