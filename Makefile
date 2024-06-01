build:
	go build -ldflags "-s -w" -o cf

clean:
	rm -f cf

fmt:
	gofmt -s -l -w .

vet:
	go vet ./...

test: vet
	go test -v -failfast ./...

release:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o cf-linux-amd64
	GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o cf-linux-386
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o cf-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o cf-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o cf-windows-amd64.exe
	GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o cf-windows-386.exe
