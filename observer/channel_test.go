//go:build unit

package observer_test

// import (
//	"context"
//	"time"
//
//	"github.com/s7techlab/hlf-sdk-go/api"
//	sdkMocks "github.com/s7techlab/hlf-sdk-go/client/testing"
//	"go.uber.org/zap"
//
//	blocksubscriber "go.b2bchain.tech/explorer/observer"
//	"go.b2bchain.tech/explorer/testdata"
//)
//
//var _ = Describe("BlockSubscriber", func() {
//	Context("Block parsing", func() {
//		var (
//			getChannelsMock api.ChannelsFetcher
//			bdMock          api.BlocksDeliverer
//
//			mspID, mspHost string
//			err            error
//		)
//		BeforeEach(func() {
//			const closeChannelWhenAllRead = true
//
//			bdMock, err = sdkMocks.NewBlocksDelivererMock("../testdata/blocks", closeChannelWhenAllRead)
//			Expect(err).To(BeNil())
//
//			var channelNames []string
//			for k := range testdata.TestChannels {
//				channelNames = append(channelNames, k)
//			}
//
//			getChannelsMock = sdkMocks.NewChannelsFetcherMock(channelNames)
//		})
//
//		Context("should subscribe to channels and parse blocks", func() {
//			It("default configuration. no channel provided, no auto subscribe", func() {
//				bs, err := blocksubscriber.NewPeerBlockSubscriber(bdMock, getChannelsMock, mspID, mspHost, zap.NewExample())
//				Expect(err).To(BeNil())
//
//				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
//				defer cancel()
//
//				parsedBlocks, err := bs.Start(ctx)
//				Expect(err).To(BeNil())
//
//				totalBlocks := 0
//				for range parsedBlocks {
//					totalBlocks++
//				}
//				Expect(totalBlocks).To(Equal(0))
//			})
//
//			It("default configuration, with auto subscribe to all channels", func() {
//				bs, err := blocksubscriber.NewPeerBlockSubscriber(bdMock, getChannelsMock, mspID, mspHost, zap.NewExample(),
//					blocksubscriber.SetObserveNewChannels(true))
//				Expect(err).To(BeNil())
//
//				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
//				defer cancel()
//
//				parsedBlocks, err := bs.Start(ctx)
//				Expect(err).To(BeNil())
//
//				totalBlocks := 0
//				for range parsedBlocks {
//					totalBlocks++
//				}
//				Expect(totalBlocks).To(Equal(13 + 9 + 20 + 16)) // 58 - total blocks in folder
//			})
//
//			Context("test channel settings", func() {
//				It("with one channel and disabled auto subscribe", func() {
//					seekFromBlock := 10
//					channelName := "asset-transfer-basic"
//					// basic channel have 13 blocks, we want to read from observer with index 10
//					bs, err := blocksubscriber.NewPeerBlockSubscriber(
//						bdMock,
//						getChannelsMock,
//						mspID, mspHost,
//						zap.NewExample(),
//						blocksubscriber.WithChannelSetting(blocksubscriber.ChannelSetting{
//							NamePattern: channelName,
//							FromBlock:   uint64(seekFromBlock),
//						}),
//					)
//					Expect(err).To(BeNil())
//
//					ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
//					defer cancel()
//
//					parsedBlocks, err := bs.Start(ctx)
//					Expect(err).To(BeNil())
//
//					totalBlocks := 0
//					for v := range parsedBlocks {
//						Expect(v.Block.Header.Number).To(Equal(uint64(seekFromBlock + totalBlocks)))
//						Expect(v.ChannelName).To(Equal(channelName))
//						totalBlocks++
//					}
//					Expect(totalBlocks).To(Equal(3))
//				})
//
//				It("with two configured channel(one with regex) and disabled auto subscribe", func() {
//					bs, err := blocksubscriber.NewPeerBlockSubscriber(
//						bdMock,
//						getChannelsMock,
//						mspID, mspHost,
//						zap.NewExample(),
//						blocksubscriber.WithChannelSetting(
//							// expect 3 blocks from here
//							blocksubscriber.ChannelSetting{
//								NamePattern: "asset-transfer-basic",
//								// basic channel have 13 blocks, we want to read from observer with index 10
//								FromBlock: 10,
//							},
//							// and in sbe(20) + secured-agreement(16)
//							blocksubscriber.ChannelSetting{
//								NamePattern: "/^asset-transfer-s*/",
//								FromBlock:   0,
//							}),
//						// 3+20+16=39 total blocks
//					)
//					Expect(err).To(BeNil())
//
//					ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
//					defer cancel()
//
//					parsedBlocks, err := bs.Start(ctx)
//					Expect(err).To(BeNil())
//
//					totalBlocks := 0
//					for range parsedBlocks {
//						totalBlocks++
//					}
//					Expect(totalBlocks).To(Equal(39))
//				})
//			})
//		})
//	})
//})
