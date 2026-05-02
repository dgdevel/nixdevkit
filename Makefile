.DEFAULT_GOAL := compile

compile:
	go build -o nixdevkit-mcp .
	go build -o nixdevkit-config ./cmd/nixdevkit-config/
	go build -o nixdevkit-setup-indexer ./cmd/nixdevkit-setup-indexer/
	CGO_ENABLED=1 go build -o nixdevkit-indexer ./cmd/nixdevkit-indexer/

clean:
	rm -f nixdevkit-mcp nixdevkit-config nixdevkit-setup-indexer nixdevkit-indexer

test:
	go vet ./...
	go test ./... || go test -v ./...

run: compile
	./nixdevkit-mcp --stdio

.PHONY: compile clean test run
