package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net/http"
	"time"

	"gopkg.in/matryer/try.v1"
)

const NumberOfRetries = 10

// CLI needs to be exported
type CLI struct {
	Download DownloadCommand `cmd help:"downloads all tweet by a given user"`
}

type DownloadCommand struct {
	User string `help:"the name user you want to download all tweets of"`
}

type EnvConfig struct {
	ConsumerKey    string `envconfig:"CONSUMER_KEY" required:"true"`
	ConsumerSecret string `envconfig:"CONSUMER_SECRET" required:"true"`
	AccessToken    string `envconfig:"ACCESS_TOKEN" required:"true"`
	AccessSecret   string `envconfig:"ACCESS_SECRET" required:"true"`
}

func (dc *DownloadCommand) Run(cli CLI, env EnvConfig) error {
	out := []twitter.Tweet{}

	config := oauth1.NewConfig(env.ConsumerKey, env.ConsumerSecret)
	token := oauth1.NewToken(env.AccessToken, env.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	next := true
	lastId := int64(0)

	for next {
		err := try.Do(func(attempt int) (bool, error) {

			log.Info(len(out))

			tweets, res, inErr := client.Search.Tweets(&twitter.SearchTweetParams{
				Query: fmt.Sprintf("@" + cli.Download.User),
				MaxID: lastId,
			})
			if inErr != nil || res.StatusCode != http.StatusOK {
				log.WithError(inErr).Errorf("Could not get tweets for account %s", cli.Download.User)
			}

			if tweets.Statuses[len(tweets.Statuses)-1].ID == lastId {
				next = false
			}

			lastId = tweets.Statuses[len(tweets.Statuses)-1].ID

			out = append(out, tweets.Statuses...)

			jitter := rand.Float64() * 5 // between 0 and 5 seconds jitter
			time.Sleep(time.Duration(jitter + math.Min(30, math.Pow(2, float64(attempt)/2))))
			return attempt < NumberOfRetries, inErr
		})
		if err != nil {
			log.WithError(err).Errorf("could not download all tweets of user %s", cli.Download.User)

			return err
		}
	}

	log.Infof("found %d tweets", len(out))

	for _, tweet := range out {
		jsonTweet, err := json.Marshal(tweet)
		if err != nil {
			log.WithError(err).Errorf("could not marshal tweet to json %v", tweet)
		}
		fmt.Println(jsonTweet)
	}

	return nil
}

func main() {

	var env EnvConfig

	err := envconfig.Process("", &env)
	if err != nil {
		log.WithError(err).Panic("error parsing environment config")
	}

	var cli CLI

	ctx := kong.Parse(&cli)
	err = ctx.Run(cli, env)
	if err != nil {
		log.WithError(err).Panic("unexpected CLI error")
	}

}
