package proto_test

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sdkmocks "github.com/s7techlab/hlf-sdk-go/api/mocks"
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

			Expect(parsedBlock.Data.Envelopes).Should(HaveLen(1))
			Expect(parsedBlock.Data.Envelopes[0].ChannelHeader().ChannelId).Should(Equal(channelName))
			Expect(parsedBlock.Data.Envelopes[0].Signature).ShouldNot(BeEmpty())
			Expect(parsedBlock.Data.Envelopes[0].ChannelHeader().TxId).ShouldNot(BeEmpty())
			Expect(parsedBlock.Data.Envelopes[0].ValidationCode).Should(Equal(peer.TxValidationCode_VALID))
			Expect(parsedBlock.Data.Envelopes[0].SignatureHeader().Creator.IdBytes).ShouldNot(BeEmpty())

			if blockNum < 4 {
				Expect(common.HeaderType(parsedBlock.Data.Envelopes[0].ChannelHeader().Type)).Should(Equal(common.HeaderType_CONFIG))
				Expect(parsedBlock.Data.Envelopes[0].Payload.Transaction.ChannelConfig).ShouldNot(BeZero())

				Expect(parsedBlock.Data.Envelopes[0].SignatureHeader().Creator.Mspid).Should(Equal(ordererMSP))
				Expect(parsedBlock.Data.Envelopes[0].TxActions()).Should(HaveLen(0))
			} else {
				Expect(parsedBlock.Data.Envelopes[0].Payload.Transaction.ChannelConfig).Should(BeZero())
				Expect(common.HeaderType(parsedBlock.Data.Envelopes[0].ChannelHeader().Type)).Should(Equal(common.HeaderType_ENDORSER_TRANSACTION))

				Expect(parsedBlock.Data.Envelopes[0].SignatureHeader().Creator.Mspid).Should(ContainSubstring(org))
				Expect(parsedBlock.Data.Envelopes[0].SignatureHeader().Creator.Mspid).Should(ContainSubstring(msp))

				Expect(parsedBlock.Data.Envelopes[0].TxActions()).Should(HaveLen(1))
				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].Header.Creator.Mspid).Should(ContainSubstring(org))
				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].Header.Creator.Mspid).Should(ContainSubstring(msp))
				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].Header.Creator.IdBytes).ShouldNot(BeEmpty())
				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].Endorsements()).ShouldNot(HaveLen(0))

				for _, endorser := range parsedBlock.Data.Envelopes[0].TxActions()[0].Endorsements() {
					Expect(endorser.Endorser.Mspid).Should(ContainSubstring(org))
					Expect(endorser.Endorser.Mspid).Should(ContainSubstring(msp))
					Expect(endorser.Endorser.IdBytes).ShouldNot(BeEmpty())
				}

				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()).Should(HaveLen(2))
				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[0].Rwset.Reads).ShouldNot(HaveLen(0))
				Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[1].Rwset.Reads).ShouldNot(HaveLen(0))

				for _, read := range parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[0].Rwset.Reads {
					Expect(read.Key).Should(ContainSubstring(namespaces))
					Expect(read.Key).Should(ContainSubstring(chaincodeName))
				}

				if blockNum == 6 {
					Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[0].Rwset.Writes).Should(HaveLen(5))
				}

				for _, read := range parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[1].Rwset.Reads {
					Expect(read.Key).ShouldNot(ContainSubstring(namespaces))
					if blockNum < 7 {
						Expect(read.Key).Should(ContainSubstring(chaincodeName))
					} else {
						Expect(read.Key).Should(ContainSubstring(initialized))
					}
				}

				if blockNum == 7 {
					Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[1].Rwset.Writes).Should(HaveLen(11))
				} else if blockNum > 7 {
					Expect(parsedBlock.Data.Envelopes[0].TxActions()[0].NsReadWriteSet()[1].Rwset.Writes).Should(HaveLen(1))
				}
			}

			if blockNum == 0 {
				Expect(parsedBlock.Metadata.OrdererSignatures).Should(HaveLen(0))
			} else {
				Expect(parsedBlock.Metadata.OrdererSignatures).Should(HaveLen(1))
				Expect(parsedBlock.Metadata.OrdererSignatures[0].Identity.Mspid).Should(Equal(ordererMSP))
				Expect(parsedBlock.Metadata.OrdererSignatures[0].Identity.IdBytes).ShouldNot(BeEmpty())
				Expect(parsedBlock.Metadata.OrdererSignatures[0].Signature).Should(BeNil())
			}

			blockNum++
		}

		err = closer()
		Expect(err).ShouldNot(HaveOccurred())
	})
})
