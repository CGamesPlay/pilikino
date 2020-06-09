PKG=github.com/CGamesPlay/pilikino

.PHONY: install
install:
	go install $(PKG)/...

.PHONY: test
test: install
	go test $(PKG)/...
