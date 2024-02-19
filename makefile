BINARY_NAME=myapp

ENV_FILE=.env

MAIN_FILE=cmd/*.go

default: build

build:
	@echo "Building..."
	@go build -o ${BINARY_NAME} ${MAIN_FILE}

run:
	@echo "Running..."
	@./loadenv.sh go run ${MAIN_FILE}

clean:
	@echo "Cleaning..."
	@go clean
	@rm -f ${BINARY_NAME}

help:
	@echo "Makefile for Go application"
	@echo "Usage:"
	@echo "  make        - build the binary"
	@echo "  make run    - run the application"
	@echo "  make clean  - remove binary files"
	@echo "  make help   - display this help"

.PHONY: default build run clean help