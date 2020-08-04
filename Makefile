GO=go
TARGET=url_shortener
OUTPUT=./build
VERSION=$(shell git describe --all --tags --abbrev=4 --dirty)
GOVERSION=$(shell go version)
OS=$(shell uname -s)
TARGET_SUFFIX=

ifeq ($(OS),Darwin)
    BUILDTYPE=darwin
else ifeq ($(OS),Linux)
    BUILDTYPE=linux
else
    BUILDTYPE=windows
    TARGET_SUFFIX=.exe
endif

.PHONEY: clean build

build:
	$(call build_execuable,$(BUILDTYPE),$(TARGET_SUFFIX))

win:
	$(call build_execuable,windows,.exe)

linux:
	$(call build_execuable,linux,)

osx:
	$(call build_execuable,darwin,)

debug:
	$(GO) run .

all: win linux osx

clean:
	rm -rf $(OUTPUT)

docker: 
	$(DOCKER) build -t harbor.supwisdom.com/epay/urlshortener:latest	


define build_execuable
	GOOS=$(1) GOARCH=amd64 $(GO) build -v \
    -ldflags '-X "main.GOVERSION=$(GOVERSION)" -X "main.VERSION=$(VERSION)"' -o $(OUTPUT)/$(TARGET)_$(1)$(2)
endef
