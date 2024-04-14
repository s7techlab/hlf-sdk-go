package observer

import (
	"fmt"
	"regexp"

	"github.com/hyperledger/fabric-protos-go/peer"
)

const MatchAnyPattern = `*`

type (
	ChannelToMatch struct {
		Name            string `json:"name" yaml:"name"`
		MatchPattern    string `json:"match_pattern" yaml:"matchPattern"`
		NotMatchPattern string `json:"not_match_pattern" yaml:"notMatchPattern"`
	}

	ChannelMatched struct {
		Name string
		// name from settings that lead to this subscription
		MatchPattern    string
		NotMatchPattern string
	}

	ChannelsMatcher struct {
		matchers []*channelMatcher
	}

	channelMatcher struct {
		name            string
		matchPattern    string
		notMatchPattern string
		matchAny        bool
		regexp          *regexp.Regexp
		regexpNotMatch  *regexp.Regexp
	}
)

var MatchAllChannels = []ChannelToMatch{{
	MatchPattern: MatchAnyPattern,
}}

func newChannelMatcher(toMatch ChannelToMatch) (*channelMatcher, error) {
	matcher := &channelMatcher{
		name:            toMatch.Name,
		matchPattern:    toMatch.MatchPattern,
		notMatchPattern: toMatch.NotMatchPattern,
	}

	if toMatch.MatchPattern == MatchAnyPattern {
		matcher.matchAny = true
		return matcher, nil
	}

	if toMatch.NotMatchPattern != `` {
		var err error
		matcher.regexpNotMatch, err = regexp.Compile(toMatch.NotMatchPattern)
		if err != nil {
			return nil, err
		}
	}

	if toMatch.MatchPattern != `` {
		var err error
		matcher.regexp, err = regexp.Compile(toMatch.MatchPattern)
		if err != nil {
			return nil, err
		}
	}

	return matcher, nil
}

func (cm *channelMatcher) Match(channel string) *ChannelMatched {
	switch {
	case cm.matchAny:
		return &ChannelMatched{
			Name:         channel,
			MatchPattern: MatchAnyPattern,
		}

	case cm.name == channel:
		return &ChannelMatched{
			Name: channel,
		}

	default:
		chMatched := &ChannelMatched{
			Name:            channel,
			MatchPattern:    cm.matchPattern,
			NotMatchPattern: cm.notMatchPattern,
		}
		if cm.regexpNotMatch != nil && cm.regexpNotMatch.MatchString(channel) {
			return nil
		}

		if cm.regexp != nil && !cm.regexp.MatchString(channel) {
			return nil
		}

		return chMatched
	}
}

func NewChannelsMatcher(channelsToMatch []ChannelToMatch) (*ChannelsMatcher, error) {
	if len(channelsToMatch) == 0 {
		channelsToMatch = MatchAllChannels
	}
	channelsMatcher := &ChannelsMatcher{}
	for _, toMatch := range channelsToMatch {
		matcher, err := newChannelMatcher(toMatch)
		if err != nil {
			pattern := toMatch.MatchPattern
			if toMatch.NotMatchPattern != `` {
				pattern = toMatch.NotMatchPattern
			}

			return nil, fmt.Errorf(`channel match name=%s, pattern=%s: %w`, toMatch.Name, pattern, err)
		}

		channelsMatcher.matchers = append(channelsMatcher.matchers, matcher)
	}

	return channelsMatcher, nil
}

func ChannelsInfoToStrings(channelsInfo []*peer.ChannelInfo) []string {
	channels := make([]string, 0)
	for _, channelInfo := range channelsInfo {
		channels = append(channels, channelInfo.ChannelId)
	}

	return channels
}

func (cm *ChannelsMatcher) Match(channels []string) ([]*ChannelMatched, error) {
	var matched []*ChannelMatched

	for _, channel := range channels {
		for _, matcher := range cm.matchers {
			if match := matcher.Match(channel); match != nil {
				matched = append(matched, match)
				break
			}
		}
	}
	return matched, nil
}
