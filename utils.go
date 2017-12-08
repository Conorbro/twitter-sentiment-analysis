package main

import (
	"regexp"
	"strings"
)

// cleanseTweet removes hashtags and @s from tweet, preserves content after # character.
func cleanseTweet(tweet string) string {
	tweet = strings.Replace(tweet, "#", "", -1)
	atUserRegex := regexp.MustCompile("@[A-Za-z]*")
	tweet = atUserRegex.ReplaceAllString(tweet, "")
	return tweet
}

// Round rounds
func Round(x, unit float32) float32 {
	if x > 0 {
		return float32(int32(x/(unit+0.5))) * unit
	}
	return float32(int32(x/unit-0.5)) * unit
}
