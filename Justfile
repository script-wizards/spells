build:
	go build ./cmd/spells

test:
	go test ./...

lint:
	go vet ./...