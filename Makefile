# Go parameters
WORK_DIR=$(shell pwd)
GOOS=linux
GOARCH=amd64
GOPRIVATE=gitlab-smartgaia.sercomm.com
GOCMD=go
GODEBUGOPTIONS=
GOBUILD=GOOS=${GOOS} GOARCH=${GOARCH} GOPRIVATE=$(GOPRIVATE) $(GOCMD) build $(GODEBUGOPTIONS)
GOCLEAN=$(GOCMD) clean
GOGET=GOPRIVATE=$(GOPRIVATE) $(GOCMD) get
GOTEST=GOPRIVATE=$(GOPRIVATE) $(GOCMD) test

.PHONY: all test

all: clean test

clean:

test: 
	$(GOTEST) -v ./...