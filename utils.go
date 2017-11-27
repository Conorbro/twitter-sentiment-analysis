package main

import (
	"regexp"
	"strings"
)

// cleanseTweet removes hashtags and @s from tweet, preserves content after # character.
func cleanseTweet(tweet string) string {
	tweet = strings.Replace(tweet, "#", "", -1)
	atUserRegex := regexp.MustCompile("@[A-Za-z]*")
	// urlRegex := regexp.MustCompile("(\b(https?|ftp|file)://)?[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]")
	// tweet = urlRegex.ReplaceAllString(tweet, "")
	tweet = atUserRegex.ReplaceAllString(tweet, "")
	return tweet
}
