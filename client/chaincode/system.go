package chaincode

const (
	CSCC = `cscc`

	CSCCJoinChain        string = "JoinChain"
	CSCCGetConfigBlock   string = "GetConfigBlock"
	CSCCGetChannels      string = "GetChannels"
	CSCCGetConfigTree    string = `GetConfigTree`    // HLF Peer V1.x
	CSCCGetChannelConfig string = "GetChannelConfig" // HLF Peer V2 +

	QSCC      = `qscc`
	LSCC      = `lscc`
	Lifecycle = `_lifecycle`
)
