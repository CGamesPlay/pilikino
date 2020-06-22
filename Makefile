PKG=github.com/CGamesPlay/pilikino
SOURCES = $(wildcard *.go) go.mod go.sum
GOFLAGS = -ldflags '-s -w -extldflags "-static"'

bin/darwin_amd64/pilikino: $(SOURCES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(GOFLAGS) $(PKG)/cmd/pilikino

bin/linux_amd64/pilikino: $(SOURCES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(GOFLAGS) $(PKG)/cmd/pilikino

bin/%.tar.gz: bin/%/pilikino
	tar -czf $@ -C $(dir $^) $(notdir $^)

.PHONY: release
release: bin/darwin_amd64.tar.gz bin/linux_amd64.tar.gz

.PHONY: generate
generate:
	go generate $(PKG)/...

.PHONY: install
install:
	go install $(PKG)/...

.PHONY: test
test: install
	go test $(PKG)/...
