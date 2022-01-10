PKG=github.com/CGamesPlay/pilikino
GOFLAGS = -ldflags '-s -w -extldflags "-static"'

.PHONY: install
install:
	go install $(PKG)/...

.PHONY: test
test:
	go test $(PKG)/...
