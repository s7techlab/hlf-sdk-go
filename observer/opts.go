package observer

// import (
//	"context"
//	"regexp"
//	"time"
//)
//
// const (
//	DefaultCheckNewChannelsPeriod = time.Minute * 5
//)
//
// var (
//	ObserverDefaultOpts = &ObserverOpts{
//		channelsSettings:       nil,
//		blockTransformers:      nil, // no data transforming
//		checkNewChannelsPeriod: 0,   // no checks for new channels
//		blockOffsetFetcher:     nil, // always start from 0
//	}
//)
//
// type (
//	ObserverOpt  func(opts *ObserverOpts)
//	ObserverOpts struct {
//		// we'll wait for upcoming blocks from these channels even if channels not exists now and will be created
//		channelsSettings []ChannelSetting
//		// chain of transformers that will be applied to the response
//		blockTransformers []Transformer
//		// in this period of time we'll check new channels
//		checkNewChannelsPeriod time.Duration
//		// if we got new channel, and we want to read from specific offset(which will be computed on new channel connection time)
//		blockOffsetFetcher OffsetFetcher
//	}
//
//	ChannelSetting struct {
//		NamePattern string
//		FromBlock   uint64
//		// will be set in BlockSubscriber constructor
//		regex *regexp.Regexp
//	}
//)
//
//// WithBlockTransformers - transform response in some ways. like decrypt(could be chains, payload, events)->json->protobuf
// func WithBlockTransformers(transformers ...Transformer) ObserverOpt {
//	return func(opts *ObserverOpts) {
//		opts.blockTransformers = transformers
//	}
//}
//
//// WithChannels - specify channels and from which observer you want to get blocks from channel
//// sets channels 'name'/'seekFromBlock' settings
//// Name could be regex pattern and MUST begin and end with '/'. Example '/channel.*/'
// func WithChannels(subscribedChannels ...ChannelSetting) ObserverOpt {
//	return func(opts *ObserverOpts) {
//		opts.channelsSettings = subscribedChannels
//	}
//}
//
// func WithCheckNewChannels(check bool) ObserverOpt {
//	return func(opts *ObserverOpts) {
//		switch check {
//		case true:
//			if opts.checkNewChannelsPeriod == 0 {
//				opts.checkNewChannelsPeriod = DefaultCheckNewChannelsPeriod
//			}
//
//		case false:
//			opts.checkNewChannelsPeriod = 0
//		}
//	}
//}
//
//// WithCheckNewChannelsPeriod - in this period of time we'll check new channels
//// and subscribe to it if they satisfy settings
// func WithCheckNewChannelsPeriod(period time.Duration) ObserverOpt {
//	return func(opts *ObserverOpts) {
//		opts.checkNewChannelsPeriod = period
//	}
//}
//
//// WithBlockOffsetFetcher - sets function that will be called on new channel connection
//// offset from this function will be used as 'from_block' option instead of reading from 0
// func WithBlockOffsetFetcher(blockOffsetFetcher func(ctx context.Context, channel string) (uint64, error)) ObserverOpt {
//	return func(opts *ObserverOpts) {
//		opts.blockOffsetFetcher = blockOffsetFetcher
//	}
//}
