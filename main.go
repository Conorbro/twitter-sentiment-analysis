package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

// CurrentConsensus represents the current consensus Twitter has towards a certain #topic.
type CurrentConsensus struct {
	SentimentScoreRollingAvg float32
	DataPoints               int
	Total                    float32
}

var cc CurrentConsensus

func consumeStream() {
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
		cc.Total += getSentimentAnalysisScore(t.Text)
		cc.DataPoints++
		cc.SentimentScoreRollingAvg = cc.Total / float32(cc.DataPoints)
		fmt.Println("Rolling average =", cc.SentimentScoreRollingAvg)
		fmt.Println(getSentimentAnalysisScore(t.Text))
	}
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	if err := sc.ReloadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
		os.Exit(1)
	}

	go consumeStream()

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index", cc)
	})
	mux.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	page := fmt.Sprintf(`<html>
	// 				<head>
	// 				<title>Twitter Sentiment Analysis</title>
	// 				</head>
	// 				<body>
	// 				<h1>%s</h1>
	// 				</body>
	// 				</html>`, sc.C.TargetHashtag)
	// 	w.Write([]byte(page))
	// })

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
