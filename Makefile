.PHONY: build
build:
	go build -ldflags "-s -w" -o cf

.PHONY: clean
clean:
	rm -f cf

.PHONY: fmt
fmt:
	gofmt -s -l -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: vet
	go test -v -failfast ./...
