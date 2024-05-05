.PHONY: build
build:
	go build -ldflags "-s -w" -o cf

.PHONY: clean
clean:
	rm -f cf

.PHONY: install
install: build
	go install

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: vet
	go test -v -failfast ./...
