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

	embed, _ := buildRedditPostEmbed(post)
	if embed == nil {
		log.Error("Embed is nil")
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func buildRedditPostEmbed(post *reddit.Post) (*discordgo.MessageEmbed, error) {
	if post == nil {
		return nil, fmt.Errorf("post is nil")
	}

	var imageURL = ""
	// TODO: make valid image file extensions configurable
	if matched, err := regexp.MatchString(`^https://([^\s]+(\.(?i)(jpg|png|gif|bmp|jpeg))$)`, post.URL); err != nil {
		log.Error(err)
	} else if matched {
		imageURL = post.URL
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
		SetDescription(post.Body).
		AddField("Author and Subreddit:", fmt.Sprintf("u/%s on %s", post.Author, post.SubredditNamePrefixed)).
		AddField("Score:", fmt.Sprintf("%s %d (%d%% upvoted)", voteEmoji, post.Score, int(post.UpvoteRatio*100))).
		AddField("Comments:", fmt.Sprintf("🗨️ %d", post.NumberOfComments)).
		SetColor(int(hexColor)).
		SetURL(fmt.Sprintf("https://www.reddit.com%v", post.Permalink)).
		SetImage(imageURL).
		SetAuthor().Truncate().MessageEmbed

	return embed, nil
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
	if randomBoolean() {
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
	var sub *reddit.Subreddit
	var err error
	if nsfw == nil {
		sub, err = getRandomSubreddit()
	} else if *nsfw {
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

	if nsfw == nil {
		return posts[0], nil
	}
	for _, post := range posts {
		if post.NSFW == *nsfw {
			return post, nil
		}
	}
	return nil, fmt.Errorf("no posts found matching nsfw=%t", *nsfw)
}
