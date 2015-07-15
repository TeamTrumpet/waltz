# GOOS=darwin
GOOS=linux
GOARCH=amd64

all:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o waltz-bin ./cmd/waltz/main.go
