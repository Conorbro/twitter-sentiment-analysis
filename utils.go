package main

import (
	"regexp"
	"strings"
)

// cleanseTweet removes hashtags and @s from tweets.
func cleanseTweet(tweet string) string {
	tweet = strings.Replace(tweet, "#", "", -1)
	atUserRegex := regexp.MustCompile("@[A-Za-z]*")
	tweet = atUserRegex.ReplaceAllString(tweet, "")
	return tweet
}
