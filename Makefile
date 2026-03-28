.PHONY: build test vet clean run

build:
	go build -o chiliz ./cmd/chiliz

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f chiliz

run:
	go run ./cmd/chiliz
