GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
BINARY_NAME = transmission-go
INSTALL_DIR ?= /usr/local/bin/

all: build
build:
	CGO_LDFLAGS_ALLOW='-Wl,-O1,--sort-common,--as-needed,-z,relro,-z,now' $(GOBUILD) -o $(BINARY_NAME) -v
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
install:
	cp ./$(BINARY_NAME) $(INSTALL_DIR)
