.PHONY: build test lint run install clean

build:
	go build ./...

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

install:
	go install .

clean:
	go clean ./...
	rm -f coverage.out

# Run the CLI locally without installing
run:
	go run . $(ARGS)
