GO := $(shell command -v go version 2>/dev/null)
GOPATH ?= $(shell pwd)

check:
ifndef GO
	$(error "GO is not installed. Follow the link to install GO: https://golang.org/doc/install")
endif

	GOPATH=${GOPATH} go get -v github.com/satori-com/satori-rtm-sdk-go/rtm

	@echo "Use 'make run' to run the application"