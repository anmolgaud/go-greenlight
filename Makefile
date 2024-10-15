.DEFAULT_GOAL := run
build:
	@go build -o ./bin/app ./cmd/api/*

run: build
	@./bin/app