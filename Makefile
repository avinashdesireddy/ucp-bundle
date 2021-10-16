PATH := $(PATH):$(PWD)/bin
build:
	GOOS=linux GOARCH=arm64 go build -o ucp-bundle-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o ucp-bundle-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o ucp-bundle.exe main.go

kbuild:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/kubectl-mke main.go
