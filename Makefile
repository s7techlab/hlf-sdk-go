GOFLAGS ?= -mod=vendor

PROTO_PACKAGES_GO := block
PROTO_PACKAGES_SVC := service

test:
	@echo "go test -mod vendor ./..."
	@go test ./...

proto: clean
	@for pkg in $(PROTO_PACKAGES_GO) ;do echo $$pkg && buf generate --template buf.gen.go.yaml $$pkg -o ./$$(echo $$pkg | cut -d "/" -f1); done
	@for pkg in $(PROTO_PACKAGES_SVC) ;do echo $$pkg && buf generate --template buf.gen.svc.yaml $$pkg -o ./$$(echo $$pkg | cut -d "/" -f1); done

clean:
	@for pkg in $(PROTO_PACKAGES_GO); do find $$pkg \( -name '*.pb.go' -or -name '*.pb.md' \) -delete;done
	@for pkg in $(PROTO_PACKAGES_SVC); do find $$pkg \( -name '*.pb.go' -or -name '*.pb.cc.go' -or -name '*.pb.gw.go' -or -name '*.swagger.json' -or -name '*.pb.md' \) -delete;done