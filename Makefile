PKG=github.com/CGamesPlay/pilikino

.PHONY: generate
generate:
	go generate $(PKG)/...

.PHONY: install
install:
	go install $(PKG)/...

.PHONY: test
test: install
	go test $(PKG)/...
