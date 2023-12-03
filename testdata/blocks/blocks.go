package blocks

const (
	Path = "testdata/blocks/fixtures"

	SampleChannel              = "sample-channel"
	SampleChannelHeight uint64 = 10
	FabcarChannel              = "fabcar-channel"
	FabcarChannelHeight uint64 = 12

	SampleChaincode = "sample"
	FabcarChaincode = "fabcar"
)

var (
	Channels        = []string{SampleChannel, FabcarChannel}
	ChannelsHeights = map[string]uint64{SampleChannel: SampleChannelHeight, FabcarChannel: FabcarChannelHeight}

	Chaincodes = []string{SampleChaincode, FabcarChaincode}
)
