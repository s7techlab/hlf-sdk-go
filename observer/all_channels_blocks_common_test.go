package observer_test

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sdkmocks "github.com/s7techlab/hlf-sdk-go/client/deliver/testing"
	"github.com/s7techlab/hlf-sdk-go/observer"
	testdata "github.com/s7techlab/hlf-sdk-go/testdata/blocks"
)

var (
	ctx = context.Background()

	peerChannelsMockForCommon *observer.PeerChannelsMock
	allChannelBlocksCommon    *observer.AllChannelBlocksCommon
	commonBlocks              <-chan *observer.Block[*common.Block]

	peerChannelsMockConcurrentlyForCommon *observer.PeerChannelsMock
	allChannelBlocksConcurrentlyCommon    *observer.AllChannelBlocksCommon
	channelWithChannelsCommon             *observer.ChannelWithChannels[*common.Block]
)

func allChannelsBlocksCommonTestBeforeSuit() {
	const closeChannelWhenAllRead = true
	blockDelivererMock, err := sdkmocks.NewBlocksDelivererMock(fmt.Sprintf("../%s", testdata.Path), closeChannelWhenAllRead)
	Expect(err).ShouldNot(HaveOccurred())

	peerChannelsMockForCommon = observer.NewPeerChannelsMock()
	for _, channel := range testdata.Channels {
		peerChannelsMockForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	allChannelBlocksCommon = observer.NewAllChannelBlocksCommon(peerChannelsMockForCommon, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithAllChannelsBlocksObservePeriod(time.Nanosecond))

	commonBlocks = allChannelBlocksCommon.Observe(ctx)

	peerChannelsMockConcurrentlyForCommon = observer.NewPeerChannelsMock()
	for _, channel := range testdata.Channels {
		peerChannelsMockConcurrentlyForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
	}

	allChannelBlocksConcurrentlyCommon = observer.NewAllChannelBlocksCommon(peerChannelsMockConcurrentlyForCommon, blockDelivererMock,
		observer.WithBlockStopRecreateStream(true), observer.WithAllChannelsBlocksObservePeriod(time.Nanosecond))

	channelWithChannelsCommon = allChannelBlocksConcurrentlyCommon.ObserveByChannels(ctx)
}

var _ = Describe("All channels blocks common", func() {
	Context("Sequentially", func() {
		It("should return current number of channels", func() {
			channels := allChannelBlocksCommon.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels)))
		})

		It("should add channels to peerChannelsMock", func() {
			newChannels := map[string]struct{}{"channel1": {}, "channel2": {}, "channel3": {}}

			for channel := range newChannels {
				peerChannelsMockForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
			}

			// wait to allChannelsBlocksCommon observer
			time.Sleep(time.Millisecond * 10)

			channels := allChannelBlocksCommon.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels) + len(newChannels)))
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

	Context("Concurrently", func() {
		It("should return current number of channels", func() {
			channels := allChannelBlocksConcurrentlyCommon.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels)))

			channelsWithBlocks := channelWithChannelsCommon.Observe()

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
				peerChannelsMockConcurrentlyForCommon.UpdateChannelInfo(&observer.ChannelInfo{Channel: channel})
			}

			// wait to allChannelsBlocksCommon observer
			time.Sleep(time.Millisecond * 200)

			channels := allChannelBlocksConcurrentlyCommon.Channels()
			Expect(channels).To(HaveLen(len(testdata.Channels) + len(newChannels)))

			channelsWithBlocks := channelWithChannelsCommon.Observe()

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
