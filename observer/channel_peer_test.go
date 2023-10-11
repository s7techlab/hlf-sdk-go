package observer_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/s7techlab/hlf-sdk-go/observer"
	testdata "github.com/s7techlab/hlf-sdk-go/testdata/blocks"
)

var _ = Describe("Channel peer", func() {
	var (
		channelPeerFetcherMock observer.PeerChannelsFetcher
	)

	BeforeEach(func() {
		channelPeerFetcherMock = observer.NewChannelPeerFetcherMock(testdata.ChannelsHeights)
	})

	Context("", func() {
		It("default channel peer, no channel matcher", func() {
			channelPeer, err := observer.NewChannelPeer(channelPeerFetcherMock)
			Expect(err).To(BeNil())

			channelPeer.Observe(ctx)
			time.Sleep(time.Millisecond * 100)

			channelsMap := channelPeer.Channels()

			sampleChannelInfo, exist := channelsMap[testdata.SampleChannel]
			Expect(exist).To(BeTrue())
			Expect(sampleChannelInfo.Channel).To(Equal(testdata.SampleChannel))
			Expect(sampleChannelInfo.Height).To(Equal(testdata.SampleChannelHeight))

			fabcarChannelInfo, exist := channelsMap[testdata.FabcarChannel]
			Expect(exist).To(BeTrue())
			Expect(fabcarChannelInfo.Channel).To(Equal(testdata.FabcarChannel))
			Expect(fabcarChannelInfo.Height).To(Equal(testdata.FabcarChannelHeight))
		})

		It("default channel peer, with channel matcher, exclude Fabcar", func() {
			channelPeer, err := observer.NewChannelPeer(channelPeerFetcherMock,
				observer.WithChannels([]observer.ChannelToMatch{{MatchPattern: testdata.SampleChannel}}))
			Expect(err).To(BeNil())

			channelPeer.Observe(ctx)
			time.Sleep(time.Millisecond * 100)

			channelsMap := channelPeer.Channels()

			sampleChannelInfo, exist := channelsMap[testdata.SampleChannel]
			Expect(exist).To(BeTrue())
			Expect(sampleChannelInfo.Channel).To(Equal(testdata.SampleChannel))
			Expect(sampleChannelInfo.Height).To(Equal(testdata.SampleChannelHeight))

			fabcarChannelInfo, exist := channelsMap[testdata.FabcarChannel]
			Expect(exist).To(BeFalse())
			Expect(fabcarChannelInfo).To(BeNil())
		})
	})
})
