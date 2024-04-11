package observer_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/s7techlab/hlf-sdk-go/observer"
	testdata "github.com/s7techlab/hlf-sdk-go/testdata/blocks"
)

var _ = Describe("Peer channels", func() {
	var (
		peerChannelsFetcherMock observer.PeerChannelsFetcher
	)
	BeforeEach(func() {
		peerChannelsFetcherMock = observer.NewPeerChannelsFetcherMock(testdata.ChannelsHeights)
	})

	It("default peer channels, no channel matcher", func() {
		peerChannels, err := observer.NewPeerChannels(peerChannelsFetcherMock)
		Expect(err).To(BeNil())

		peerChannels.Observe(ctx)
		time.Sleep(time.Millisecond * 100)

		channelsMap := peerChannels.Channels()

		sampleChannelInfo, exist := channelsMap[testdata.SampleChannel]
		Expect(exist).To(BeTrue())
		Expect(sampleChannelInfo.Channel).To(Equal(testdata.SampleChannel))
		Expect(sampleChannelInfo.Height).To(Equal(testdata.SampleChannelHeight))

		fabcarChannelInfo, exist := channelsMap[testdata.FabcarChannel]
		Expect(exist).To(BeTrue())
		Expect(fabcarChannelInfo.Channel).To(Equal(testdata.FabcarChannel))
		Expect(fabcarChannelInfo.Height).To(Equal(testdata.FabcarChannelHeight))
	})

	It("default peer channels, with channel matcher, exclude Fabcar", func() {
		peerChannels, err := observer.NewPeerChannels(peerChannelsFetcherMock,
			observer.WithChannels([]observer.ChannelToMatch{{MatchPattern: testdata.SampleChannel}}))
		Expect(err).To(BeNil())

		peerChannels.Observe(ctx)
		time.Sleep(time.Millisecond * 100)

		channelsMap := peerChannels.Channels()

		sampleChannelInfo, exist := channelsMap[testdata.SampleChannel]
		Expect(exist).To(BeTrue())
		Expect(sampleChannelInfo.Channel).To(Equal(testdata.SampleChannel))
		Expect(sampleChannelInfo.Height).To(Equal(testdata.SampleChannelHeight))

		fabcarChannelInfo, exist := channelsMap[testdata.FabcarChannel]
		Expect(exist).To(BeFalse())
		Expect(fabcarChannelInfo).To(BeNil())
	})
})
