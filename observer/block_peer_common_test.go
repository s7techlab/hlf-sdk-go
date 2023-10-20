package observer_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sdkmocks "github.com/s7techlab/hlf-sdk-go/api/mock"
	"github.com/s7techlab/hlf-sdk-go/observer"
	testdata "github.com/s7techlab/hlf-sdk-go/testdata/blocks"
)

var (
	ctx = context.Background()

	channelPeerMockForCommon *observer.ChannelPeerMock
	commonBlockPeer          *observer.BlockPeer
	commonBlocks             <-chan *observer.Block

	channelPeerMockConcurrentlyForCommon *observer.ChannelPeerMock
	commonBlockPeerConcurrently          *observer.BlockPeer
	commonBlocksByChannels               *observer.BlocksByChannels
)

func blockPeerCommonTestBeforeSuit() {
	const closeChannelWhenAllRead = true
	blockDelivererMock, err := sdkmocks.NewBlocksDelivererMock(fmt.Sprintf("../%s", testdata.Path), closeChannelWhenAllRead)
	Expect(err).ShouldNot(HaveOccurred())

	channelPeerMockForCommon = observer.NewChannelPeerMock()
	for _, channel := range testdata.Channels {
		channelPeerMockForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	commonBlockPeer = observer.NewBlockPeer(channelPeerMockForCommon, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithBlockPeerObservePeriod(time.Nanosecond))

	commonBlocks = commonBlockPeer.Observe(ctx)

	channelPeerMockConcurrentlyForCommon = observer.NewChannelPeerMock()
	for _, channel := range testdata.Channels {
		channelPeerMockConcurrentlyForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	commonBlockPeerConcurrently = observer.NewBlockPeer(channelPeerMockConcurrentlyForCommon, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithBlockPeerObservePeriod(time.Nanosecond))

	commonBlocksByChannels = commonBlockPeerConcurrently.ObserveByChannels(ctx)
}

var _ = Describe("Block Peer", func() {
	Context("Block peer", func() {
		It("should return current number of channels", func() {
			channelObservers := commonBlockPeer.ChannelObservers()
			Expect(channelObservers).To(HaveLen(len(testdata.Channels)))
		})

		It("should add channels to channelPeerMock", func() {
			newChannels := map[string]struct{}{"channel1": {}, "channel2": {}, "channel3": {}}

			for channel := range newChannels {
				channelPeerMockForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
			}

			// wait to commonBlockPeer observer
			time.Sleep(time.Millisecond * 10)

			channelObservers := commonBlockPeer.ChannelObservers()
			Expect(channelObservers).To(HaveLen(len(testdata.Channels) + len(newChannels)))
		})

		It("should return correct channels heights", func() {
			channelsBlocksHeights := map[string]uint64{testdata.SampleChannel: 0, testdata.FabcarChannel: 0}
			for b := range commonBlocks {
				curBlockChannel := ""
				// it must only these channels, new ones do not have any blocks
				if b.Channel == testdata.SampleChannel {
					curBlockChannel = testdata.SampleChannel
				} else if b.Channel == testdata.FabcarChannel {
					curBlockChannel = testdata.FabcarChannel
				}

				Expect(b.Channel).To(Equal(curBlockChannel))

				blockNum := channelsBlocksHeights[curBlockChannel]
				Expect(b.Block.Header.Number).To(Equal(blockNum))

				channelsBlocksHeights[curBlockChannel]++

				if channelsBlocksHeights[testdata.SampleChannel] == testdata.SampleChannelHeight && channelsBlocksHeights[testdata.FabcarChannel] == testdata.FabcarChannelHeight {
					break
				}
			}

			Expect(channelsBlocksHeights[testdata.SampleChannel]).To(Equal(testdata.SampleChannelHeight))
			Expect(channelsBlocksHeights[testdata.FabcarChannel]).To(Equal(testdata.FabcarChannelHeight))
		})
	})

	Context("Block peer concurrently", func() {
		It("should return current number of channels", func() {
			channelObservers := commonBlockPeerConcurrently.ChannelObservers()
			Expect(channelObservers).To(HaveLen(len(testdata.Channels)))

			channelsWithBlocks := commonBlocksByChannels.Observe()

			for i := 0; i < len(testdata.Channels); i++ {
				sampleOrFabcarChannelBlocks := <-channelsWithBlocks

				curBlockChannel := ""
				curChannelHeight := uint64(0)
				// it must only these channels, new ones do not have any blocks
				if sampleOrFabcarChannelBlocks.Name == testdata.SampleChannel {
					curBlockChannel = testdata.SampleChannel
					curChannelHeight = testdata.SampleChannelHeight
				} else if sampleOrFabcarChannelBlocks.Name == testdata.FabcarChannel {
					curBlockChannel = testdata.FabcarChannel
					curChannelHeight = testdata.FabcarChannelHeight
				}

				Expect(sampleOrFabcarChannelBlocks.Name).To(Equal(curBlockChannel))
				Expect(sampleOrFabcarChannelBlocks.Blocks).NotTo(BeNil())

				channelBlocksHeight := uint64(0)
				for block := range sampleOrFabcarChannelBlocks.Blocks {
					Expect(block.Channel).To(Equal(curBlockChannel))
					Expect(block.Block.Header.Number).To(Equal(channelBlocksHeight))

					channelBlocksHeight++

					if channelBlocksHeight == curChannelHeight {
						break
					}
				}

				Expect(channelBlocksHeight).To(Equal(curChannelHeight))
			}
		})

		It("should add channels to channelPeerMock", func() {
			channel4, channel5, channel6 := "channel4", "channel5", "channel6"
			newChannels := []string{channel4, channel5, channel6}
			for _, channel := range newChannels {
				channelPeerMockConcurrentlyForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
			}

			// wait to commonBlockPeer observer
			time.Sleep(time.Millisecond * 200)

			channelObservers := commonBlockPeerConcurrently.ChannelObservers()
			Expect(channelObservers).To(HaveLen(len(testdata.Channels) + len(newChannels)))

			channelsWithBlocks := commonBlocksByChannels.Observe()

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
