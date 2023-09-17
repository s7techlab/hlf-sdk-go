//go:build unit

package observer_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	sdkmocks "github.com/s7techlab/hlf-sdk-go/client/testing"

	"gitlab.n-t.io/atm-ru/microservices/back/explorer/observer"
	testdata "gitlab.n-t.io/atm-ru/microservices/back/explorer/testdata/blocks"
)

var (
	ctx = context.Background()

	channelPeerMock *observer.ChannelPeerMock
	blockPeer       *observer.BlockPeer

	channelPeerMockConcurrently *observer.ChannelPeerMock
	blockPeerConcurrently       *observer.BlockPeer
	blocksByChannels            *observer.BlocksByChannels
)

var _ = BeforeSuite(func() {
	const closeChannelWhenAllRead = true
	blockDelivererMock, err := sdkmocks.NewBlocksDelivererMock(fmt.Sprintf("../%s", testdata.Path), closeChannelWhenAllRead)
	Expect(err).ShouldNot(HaveOccurred())

	channelPeerMock = observer.NewChannelPeerMock()
	for channel := range testdata.TestChannels {
		channelPeerMock.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	blockPeer = observer.NewBlockPeer(channelPeerMock, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithBlockPeerObservePeriod(time.Nanosecond))

	_, err = blockPeer.Observe(ctx)
	Expect(err).ShouldNot(HaveOccurred())

	channelPeerMockConcurrently = observer.NewChannelPeerMock()
	for channel := range testdata.TestChannels {
		channelPeerMockConcurrently.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	blockPeerConcurrently = observer.NewBlockPeer(channelPeerMockConcurrently, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithBlockPeerObservePeriod(time.Nanosecond))

	blocksByChannels, err = blockPeerConcurrently.ObserveByChannels(ctx)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Block Peer", func() {
	Context("Channels number check", func() {
		Context("Block peer", func() {
			It("should return current number of channels", func() {
				channelObservers := blockPeer.ChannelObservers()
				Expect(channelObservers).To(HaveLen(len(testdata.TestChannels)))
			})

			It("should add channels to channelPeerMock", func() {
				newChannels := []string{"channel1", "channel2", "channel3"}
				for _, channel := range newChannels {
					channelPeerMock.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
				}

				// wait to blockPeer observer
				time.Sleep(time.Millisecond * 10)

				channelObservers := blockPeer.ChannelObservers()
				Expect(channelObservers).To(HaveLen(len(testdata.TestChannels) + len(newChannels)))
			})
		})

		Context("Block peer concurrently", func() {
			It("should return current number of channels", func() {
				channelObservers := blockPeerConcurrently.ChannelObservers()
				Expect(channelObservers).To(HaveLen(len(testdata.TestChannels)))

				channelsWithBlocks := blocksByChannels.Observe()

				for i := 0; i < len(testdata.TestChannels); i++ {
					sampleOrFabcarChannelBlocks := <-channelsWithBlocks
					if sampleOrFabcarChannelBlocks.Name == testdata.SampleChannel {
						Expect(sampleOrFabcarChannelBlocks.Name).To(Equal(testdata.SampleChannel))
					} else {
						Expect(sampleOrFabcarChannelBlocks.Name).To(Equal(testdata.FabcarChannel))
					}

					Expect(sampleOrFabcarChannelBlocks.Blocks).NotTo(BeNil())
				}
			})

			It("should add channels to channelPeerMock", func() {
				channel4, channel5, channel6 := "channel4", "channel5", "channel6"
				newChannels := []string{channel4, channel5, channel6}
				for _, channel := range newChannels {
					channelPeerMockConcurrently.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
				}

				// wait to blockPeer observer
				time.Sleep(time.Millisecond * 200)

				channelObservers := blockPeerConcurrently.ChannelObservers()
				Expect(channelObservers).To(HaveLen(len(testdata.TestChannels) + len(newChannels)))

				channelsWithBlocks := blocksByChannels.Observe()

				for i := 0; i < len(newChannels); i++ {
					channel4Or5Or6Blocks := <-channelsWithBlocks

					if channel4Or5Or6Blocks.Name == channel4 {
						Expect(channel4Or5Or6Blocks.Name).To(Equal(channel4))
						Expect(channel4Or5Or6Blocks.Blocks).NotTo(BeNil())
					} else if channel4Or5Or6Blocks.Name == channel5 {
						Expect(channel4Or5Or6Blocks.Name).To(Equal(channel5))
						Expect(channel4Or5Or6Blocks.Blocks).NotTo(BeNil())
					} else {
						Expect(channel4Or5Or6Blocks.Name).To(Equal(channel6))
						Expect(channel4Or5Or6Blocks.Blocks).NotTo(BeNil())
					}
				}
			})
		})
	})
})
