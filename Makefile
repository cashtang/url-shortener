GO=go
TARGET=url_shortener
OUTPUT=./build


.PHONEY: clean

build:
	$(GO) build -v -o $(OUTPUT)/$(TARGET)


win:
	$(call build_execuable,windows,.exe)

linux:
	$(call build_execuable,linux,)

osx:
	$(call build_execuable,darwin,)

all: win linux osx

clean:
	rm -rf $(OUTPUT)


docker: build
	$(DOCKER) build -t harbor.supwisdom.com/epay/urlshortener:latest	


define build_execuable
	GOOS=$(1) GOARCH=amd64 $(GO) build -v -o $(OUTPUT)/$(TARGET)_$(1)$(2)
endef
