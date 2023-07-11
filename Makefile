GOARCH = amd64

UNAME = $(shell uname -s)

VERSION = $(shell cat VERSION)

BIN_NAME = server_start

BUILD_DIRECTORY = build

ifndef OS
	ifeq ($(UNAME), Linux)
		OS = linux
	else ifeq ($(UNAME), Darwin)
		OS = darwin
	endif
endif

.DEFAULT_GOAL := all

all: fmt $(BUILD_DIRECTORY)/$(BIN_NAME)

$(BUILD_DIRECTORY)/$(BIN_NAME): api.go ./**/*.go
	mkdir -p build
	GOOS=$(OS) GOARCH="$(GOARCH)" go build -o $(BUILD_DIRECTORY)/$(BIN_NAME) api.go

clean:
	rm -f ./build/*

fmt:
	go fmt $$(go list ./...)

.PHONY: install clean fmt
