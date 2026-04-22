BINARY := pr_approved

.PHONY: run build test test-v lint tidy clean

run:
	go run main.go

build:
	go build -o $(BINARY) .

test:
	go test ./...

test-v:
	go test ./... -v

lint:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
