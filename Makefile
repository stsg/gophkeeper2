B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
DATE=$(shell date +%Y%m%d-%H:%M:%S)
REV=$(GITREV)-$(BRANCH)-$(DATE)

info:
	- @echo "revision $(REV)"

genproto:
	rm pkg/pb/*.go || true
	protoc --proto_path=api/proto  --go_out=pkg/pb --go_opt=paths=source_relative --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative api/proto/*.proto

buildserver: info
	@ echo
	@ echo "Compiling Server Binary"
	@ echo
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(REV) -X main.buildDate=$(DATE) -X main.defaultConfigPath=./cfg/config.json" -o bin/gophkeeper cmd/server/main.go

buildclient: info
	@ echo
	@ echo "Compiling Client Binary"
	@ echo
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(REV) -X main.buildDate=$(DATE) -X main.defaultConfigPath=./cfg/gpk_config.json" -o bin/gpk-client-linux cmd/client/main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(REV) -X main.buildDate=$(DATE) -X main.defaultConfigPath=./cfg/gpk_config.json" -o bin/gpk-client-mac cmd/client/main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=$(REV) -X main.buildDate=$(DATE) -X main.defaultConfigPath=./cfg/gpk_config.json" -o bin/gpk-client-win cmd/client/main.go

docker:
	docker build -t starky/gophkeeper:master .

clean:
	@ echo
	@ echo "Cleaning"
	@ echo
	rm bin/gophkeeper || true
	rm bin/gpk-client-*

tidy:
	@ echo
	@ echo "Tidying"
	@ echo
	go mod tidy

run:
	go run cmd/server/main.go

runclient:
	go run cmd/client/main.go

lint:
	@ echo
	@ echo "Linting"
	@ echo
	golangci-lint run

test:
	@ echo
	@ echo "Testing"
	@ echo
	go test ./...

.PHONY: *
