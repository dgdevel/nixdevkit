.DEFAULT_GOAL := compile

compile:
	go build -o nixdevkit .
	go build -o nixdevkit-config ./cmd/nixdevkit-config/

clean:
	rm -f nixdevkit nixdevkit-config

test:
	go test -v ./...

run: compile
	./nixdevkit --stdio

.PHONY: compile clean test run
