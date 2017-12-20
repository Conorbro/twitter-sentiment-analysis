package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
)

var (
	sc = &SafeConfig{
		C: &Config{},
	}
	configFile    = kingpin.Flag("config.file", "Twitter sentiment analysis bot configuration file.").Default("twitter.yml").String()
	listenAddress = kingpin.Flag("listen.addr", "Address to listen on for graph view server.").Default(":3000").String()
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

// TwitterClient represents the Twitter client for making Twitter API calls.
type TwitterClient struct {
	api *anaconda.TwitterApi
}

// CurrentConsensus represents the current consensus Twitter has towards a certain #topic.
type CurrentConsensus struct {
	SentimentScoreRollingAvg float32
	DataPoints               int
	Total                    float32
}

// Sample represents an individual tweet's sentiment analysis score and its corresponding timestamp.
type Sample struct {
	SentimentScore float32 `json:"sentiment_analysis_score"`
	Timestamp      string  `json:"time_stamp"`
	Location       string  `json:"location"`
}

var (
	cc            CurrentConsensus
	twitterClient *TwitterClient
)

// NewTwitterClient authenticates with Twitter api and returns Twitter Client obj.
func NewTwitterClient() *TwitterClient {
	anaconda.SetConsumerKey(sc.C.TwitterCredentials.TwitterConsumerKey)
	anaconda.SetConsumerSecret(sc.C.TwitterCredentials.TwitterConsumerSecret)
	api := anaconda.NewTwitterApi(sc.C.TwitterCredentials.TwitterAccessToken, sc.C.TwitterCredentials.TwitterAccessTokenSecret)
	return &TwitterClient{api: api}
}

func (t *TwitterClient) consumeStream() {
	log := &logger{logrus.New()}
	t.api.SetLogger(log)

	stream := t.api.PublicStreamFilter(url.Values{
		"track": []string{sc.C.TargetHashtag},
	})

	f, err := os.Create("output.json")
	if err != nil {
		log.Errorf("Error creating output file: %v", err)
	}

	defer stream.Stop()
	log.Info("Starting stream...")
	for v := range stream.C {
		log.Info("Getting tweet...")
		t, ok := v.(anaconda.Tweet)
		if t.Lang != "en" {
			log.Debug("Tweet not English language, skipping...")
			continue
		}
		if !ok {
			log.Warningf("Received unexpected value of type %T, skipping...", v)
			continue
		}
		// Ignore retweets
		if t.RetweetedStatus != nil {
			continue
		}
		tweetText := cleanseTweet(t.Text)
		sentimentAnalysisScore := getSentimentAnalysisScore(tweetText)
		cc.Total += sentimentAnalysisScore
		cc.DataPoints++
		log.Infof("cc.Total = %v", cc.Total)
		log.Infof("cc.DataPoints = %v", float32(cc.DataPoints))
		cc.SentimentScoreRollingAvg = Round(cc.Total/float32(cc.DataPoints), 0.00005)
		log.Infof("Rolling average = %v", cc.SentimentScoreRollingAvg)
		smpl := &Sample{SentimentScore: sentimentAnalysisScore, Location: t.Place.FullName, Timestamp: t.CreatedAt}
		writeSmpl, err := json.Marshal(smpl)
		if err != nil {
			log.Errorf("Error converting sample to JSON: %v", err)
		}
		f.WriteString(string(writeSmpl) + "\n")
	}
}

func postToTwitter() {
	ticker := time.NewTicker(2 * time.Hour)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			status := fmt.Sprintf("Current Average Sentiment Analysis for %v: %v for %v samples", sc.C.TargetHashtag, cc.SentimentScoreRollingAvg, cc.DataPoints)
			_, err := twitterClient.api.PostTweet(status, nil)
			if err != nil {
				log.Fatalf("Error posting Tweet: %v", err)
			}
			log.Println("Current Average Sentiment Analysis:", cc.SentimentScoreRollingAvg)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", sc.C.GoogleApplicationCredentialsFile)

	twitterClient = NewTwitterClient()

	// Kick off goroutine to consume tweets from Twitter.
	go twitterClient.consumeStream()
	// Post the current average sentiment analysis to Twitter.
	go postToTwitter()

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index", cc)
	})
	mux.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))

	if err := http.ListenAndServe(*listenAddress, mux); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
		os.Exit(1)
	}

}

type logger struct {
	*logrus.Logger
}

func (log *logger) Critical(args ...interface{})                 { log.Error(args...) }
func (log *logger) Criticalf(format string, args ...interface{}) { log.Errorf(format, args...) }
func (log *logger) Notice(args ...interface{})                   { log.Info(args...) }
func (log *logger) Noticef(format string, args ...interface{})   { log.Infof(format, args...) }
