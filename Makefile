.PHONY: build run clean

build:
	go build -o go-auth-api ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -f go-auth-api





