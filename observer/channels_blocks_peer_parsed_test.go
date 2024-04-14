package observer_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
	sdkmocks "github.com/s7techlab/hlf-sdk-go/client/deliver/testing"
	"github.com/s7techlab/hlf-sdk-go/observer"
	testdata "github.com/s7techlab/hlf-sdk-go/testdata/blocks"
)

var (
	peerChannelsMockForParsed *observer.PeerChannelsMock
	channelsBlocksPeerParsed  *observer.ChannelsBlocksPeerParsed
	parsedBlocks              <-chan *observer.Block[*hlfproto.Block]

	peerChannelsMockConcurrentlyForParsed *observer.PeerChannelsMock
	channelsBlocksPeerConcurrentlyParsed  *observer.ChannelsBlocksPeerParsed
	channelWithChannelsParsed             *observer.ChannelWithChannels[*hlfproto.Block]
)

func channelsBlocksPeerParsedTestBeforeSuit() {
	const closeChannelWhenAllRead = true
	blockDelivererMock, err := sdkmocks.NewBlocksDelivererMock(fmt.Sprintf("../%s", testdata.Path), closeChannelWhenAllRead)
	Expect(err).ShouldNot(HaveOccurred())

	peerChannelsMockForParsed = observer.NewPeerChannelsMock()
	for _, channel := range testdata.Channels {
		peerChannelsMockForParsed.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	channelsBlocksPeerParsed = observer.NewChannelsBlocksPeerParsed(peerChannelsMockForParsed, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithChannelsBlocksPeerRefreshPeriod(time.Nanosecond))

	parsedBlocks = channelsBlocksPeerParsed.Observe(ctx)

	peerChannelsMockConcurrentlyForParsed = observer.NewPeerChannelsMock()
	for _, channel := range testdata.Channels {
		peerChannelsMockConcurrentlyForParsed.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	channelsBlocksPeerConcurrentlyParsed = observer.NewChannelsBlocksPeerParsed(peerChannelsMockConcurrentlyForParsed, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithChannelsBlocksPeerRefreshPeriod(time.Nanosecond))

	channelWithChannelsParsed = channelsBlocksPeerConcurrentlyParsed.ObserveByChannels(ctx)
}

var _ = Describe("All channels blocks parsed", func() {
	Context("Sequentially", func() {
		It("should return current number of channels", func() {
			channels := channelsBlocksPeerParsed.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels)))
		})

		It("should add channels to peerChannelsMock", func() {
			newChannels := map[string]struct{}{"channel1": {}, "channel2": {}, "channel3": {}}

			for channel := range newChannels {
				peerChannelsMockForParsed.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
			}

			// wait to channelsBlocksPeerParsed observer
			time.Sleep(time.Second + time.Millisecond*10)

			channels := channelsBlocksPeerParsed.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels) + len(newChannels)))
		})

		It("should return correct channels heights", func() {
			channelsBlocksHeights := map[string]uint64{testdata.SampleChannel: 0, testdata.FabcarChannel: 0}
			for b := range parsedBlocks {
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

	Context("Concurrently", func() {
		It("should return current number of channels", func() {
			channels := channelsBlocksPeerConcurrentlyParsed.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels)))

			channelsWithBlocks := channelWithChannelsParsed.Observe()

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

		It("should add channels to peerChannelsMock", func() {
			channel4, channel5, channel6 := "channel4", "channel5", "channel6"
			newChannels := []string{channel4, channel5, channel6}
			for _, channel := range newChannels {
				peerChannelsMockConcurrentlyForParsed.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
			}

			// wait to channelsBlocksPeerParsed observer
			time.Sleep(time.Millisecond * 200)

			channels := channelsBlocksPeerConcurrentlyParsed.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels) + len(newChannels)))

			channelsWithBlocks := channelWithChannelsParsed.Observe()

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
