PKG=github.com/CGamesPlay/pilikino
SOURCES = $(wildcard *.go) go.mod go.sum
GOFLAGS = -ldflags '-s -w -extldflags "-static"'

.PHONY: install
install: generate
	go install $(PKG)/...

.PHONY: test
test: install
	go test $(PKG)/...

.PHONY: release
release: bin/darwin_amd64.tar.gz bin/linux_amd64.tar.gz

.PHONY: generate
generate:
	go generate $(PKG)/...

.PHONY: docker-vroom
docker-vroom:
	docker build bin/linux_amd64 -f vim/vroom/Dockerfile -t vroom

.PHONY: vim-test
vim-test: bin/linux_amd64/pilikino
	docker run --rm -it -v `pwd`:/vroom vroom

bin/darwin_amd64/pilikino: $(SOURCES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(GOFLAGS) $(PKG)/cmd/pilikino

bin/linux_amd64/pilikino: $(SOURCES)
	mkdir -p $(dir $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(GOFLAGS) $(PKG)/cmd/pilikino

bin/%.tar.gz: bin/%/pilikino
	tar -czf $@ -C $(dir $^) $(notdir $^)
