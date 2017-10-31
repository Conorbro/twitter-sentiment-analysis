package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
)

var (
	sc = &SafeConfig{
		C: &Config{},
	}
	configFile = kingpin.Flag("config.file", "Twitter sentiment analysis bot configuration file.").Default("twitter.yml").String()

	// consumerKey       = getenv("TWITTER_CONSUMER_KEY")
	// consumerSecret    = getenv("TWITTER_CONSUMER_SECRET")
	// accessTokenKey    = getenv("TWITTER_ACCESS_TOKEN")
	// accessTokenSecret = getenv("TWITTER_ACCESS_TOKEN_SECRET")
)

// func getenv(name string) string {
// 	v := os.Getenv(name)
// 	if v == "" {
// 		panic("Missing required environment variable " + name)
// 	}
// 	return v
// }

func main() {
	kingpin.Parse()
	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	anaconda.SetConsumerKey(sc.C.TwitterCredentials.TwitterConsumerKey)
	anaconda.SetConsumerSecret(sc.C.TwitterCredentials.TwitterConsumerSecret)
	api := anaconda.NewTwitterApi(sc.C.TwitterCredentials.TwitterAccessToken, sc.C.TwitterCredentials.TwitterAccessTokenSecret)

	log := &logger{logrus.New()}
	api.SetLogger(log)

	stream := api.PublicStreamFilter(url.Values{
		"track": []string{sc.C.TargetHashtag},
	})

	defer stream.Stop()

	for v := range stream.C {
		t, ok := v.(anaconda.Tweet)
		if !ok {
			log.Warningf("Received unexpected value of type %T", v)
			continue
		}

		// Ignore retweets
		if t.RetweetedStatus != nil {
			continue
		}

		fmt.Println(getSentimentAnalysisScore(t.Text))
	}
}

type logger struct {
	*logrus.Logger
}

func (log *logger) Critical(args ...interface{})                 { log.Error(args...) }
func (log *logger) Criticalf(format string, args ...interface{}) { log.Errorf(format, args...) }
func (log *logger) Notice(args ...interface{})                   { log.Info(args...) }
func (log *logger) Noticef(format string, args ...interface{})   { log.Infof(format, args...) }
