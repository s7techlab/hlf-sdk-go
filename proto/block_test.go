package proto_test

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sdkmocks "github.com/s7techlab/hlf-sdk-go/client/testing"
	"github.com/s7techlab/hlf-sdk-go/proto"
)

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Block")
}

const (
	channelName   = "sample-channel"
	chaincodeName = "sample"
	namespaces    = "namespaces/"
	initialized   = "initialized"
	ordererMSP    = "OrdererMSP"
	org           = "Org"
	msp           = "MSP"
)

var (
	blockDelivererMock *sdkmocks.BlocksDelivererMock
	err                error
)

var _ = BeforeSuite(func(done Done) {
	const closeChannelWhenAllRead = true
	blockDelivererMock, err = sdkmocks.NewBlocksDelivererMock("./fixtures", closeChannelWhenAllRead)
	Expect(err).ShouldNot(HaveOccurred())

	close(done)
}, 1000)

var _ = Describe("Block parse test", func() {
	It("", func() {
		blocks, closer, err := blockDelivererMock.Blocks(context.Background(), channelName, nil)
		Expect(err).ShouldNot(HaveOccurred())

		blockNum := 0
		for {
			block, ok := <-blocks
			if !ok {
				break
			}

			parsedBlock, err := proto.ParseBlock(block)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(parsedBlock.Header.Number).Should(BeNumerically("==", blockNum))
			if blockNum != 0 {
				Expect(parsedBlock.Header.PreviousHash).ShouldNot(BeEmpty())
			}
			Expect(parsedBlock.Header.DataHash).ShouldNot(BeEmpty())

			Expect(parsedBlock.Envelopes).Should(HaveLen(1))
			Expect(parsedBlock.Envelopes[0].ChannelHeader.ChannelId).Should(Equal(channelName))
			Expect(parsedBlock.Envelopes[0].Signature).ShouldNot(BeEmpty())
			Expect(parsedBlock.Envelopes[0].ChannelHeader.TxId).ShouldNot(BeEmpty())
			Expect(parsedBlock.Envelopes[0].ValidationCode).Should(Equal(peer.TxValidationCode_VALID))
			Expect(parsedBlock.Envelopes[0].Transaction.CreatorIdentity.IdBytes).ShouldNot(BeEmpty())

			if blockNum < 4 {
				Expect(common.HeaderType(parsedBlock.Envelopes[0].ChannelHeader.Type)).Should(Equal(common.HeaderType_CONFIG))
				Expect(parsedBlock.Envelopes[0].ChannelConfig).ShouldNot(BeZero())

				Expect(parsedBlock.Envelopes[0].Transaction.CreatorIdentity.Mspid).Should(Equal(ordererMSP))
				Expect(parsedBlock.Envelopes[0].Transaction.Actions).Should(HaveLen(0))
			} else {
				Expect(parsedBlock.Envelopes[0].ChannelConfig).Should(BeZero())
				Expect(common.HeaderType(parsedBlock.Envelopes[0].ChannelHeader.Type)).Should(Equal(common.HeaderType_ENDORSER_TRANSACTION))

				Expect(parsedBlock.Envelopes[0].Transaction.CreatorIdentity.Mspid).Should(ContainSubstring(org))
				Expect(parsedBlock.Envelopes[0].Transaction.CreatorIdentity.Mspid).Should(ContainSubstring(msp))

				Expect(parsedBlock.Envelopes[0].Transaction.Actions).Should(HaveLen(1))
				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].CreatorIdentity.Mspid).Should(ContainSubstring(org))
				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].CreatorIdentity.Mspid).Should(ContainSubstring(msp))
				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].CreatorIdentity.IdBytes).ShouldNot(BeEmpty())
				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].Endorsers).ShouldNot(HaveLen(0))

				for _, endorser := range parsedBlock.Envelopes[0].Transaction.Actions[0].Endorsers {
					Expect(endorser.Mspid).Should(ContainSubstring(org))
					Expect(endorser.Mspid).Should(ContainSubstring(msp))
					Expect(endorser.IdBytes).ShouldNot(BeEmpty())
				}

				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets).Should(HaveLen(2))
				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[0].Reads).ShouldNot(HaveLen(0))
				Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[1].Reads).ShouldNot(HaveLen(0))

				for _, read := range parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[0].Reads {
					Expect(read.Key).Should(ContainSubstring(namespaces))
					Expect(read.Key).Should(ContainSubstring(chaincodeName))
				}

				if blockNum == 6 {
					Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[0].Writes).Should(HaveLen(5))
				}

				for _, read := range parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[1].Reads {
					Expect(read.Key).ShouldNot(ContainSubstring(namespaces))
					if blockNum < 7 {
						Expect(read.Key).Should(ContainSubstring(chaincodeName))
					} else {
						Expect(read.Key).Should(ContainSubstring(initialized))
					}
				}

				if blockNum == 7 {
					Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[1].Writes).Should(HaveLen(11))
				} else if blockNum > 7 {
					Expect(parsedBlock.Envelopes[0].Transaction.Actions[0].ReadWriteSets[1].Writes).Should(HaveLen(1))
				}
			}

			if blockNum == 0 {
				Expect(parsedBlock.OrdererSignatures).Should(HaveLen(0))
			} else {
				Expect(parsedBlock.OrdererSignatures).Should(HaveLen(1))
				Expect(parsedBlock.OrdererSignatures[0].Identity.Mspid).Should(Equal(ordererMSP))
				Expect(parsedBlock.OrdererSignatures[0].Identity.IdBytes).ShouldNot(BeEmpty())
				Expect(parsedBlock.OrdererSignatures[0].Signature).Should(BeNil())
			}

			blockNum++
		}

		err = closer()
		Expect(err).ShouldNot(HaveOccurred())
	})
})
