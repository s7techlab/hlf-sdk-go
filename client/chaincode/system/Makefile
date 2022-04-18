generate:
	@protoc -I=. \
	-I=../../../third_party  \
	-I=../../../third_party/hyperledger/fabric-protos  \
	--go_out=paths=source_relative,plugins=grpc:.  \
	--grpc-gateway_out=paths=source_relative:. \
	--swagger_out=logtostderr=true:. \
	*.proto
