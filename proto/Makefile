generate:
	@protoc -I=. \
		-I=../third_party/hyperledger/fabric-protos  \
	    --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		*.proto
