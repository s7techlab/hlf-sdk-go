GOFLAGS ?= -mod=vendor

PROTO_PACKAGES_GO := proto

test:
	@echo "go test -mod vendor ./..."
	@go test ./...

proto: clean
	@for pkg in $(PROTO_PACKAGES_GO) ;do echo $$pkg && buf generate --template buf.gen.go.yaml $$pkg -o ./$$(echo $$pkg | cut -d "/" -f1); done

clean:
	@for pkg in $(PROTO_PACKAGES_GO); do find $$pkg \( -name '*.pb.go' -or -name '*.pb.md' \) -delete;done
