package chaincode

func argsToBytes(args ...string) [][]byte {
	retArgs := make([][]byte, 0)
	for _, arg := range args {
		retArgs = append(retArgs, []byte(arg))
	}
	return retArgs
}
