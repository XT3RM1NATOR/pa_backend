-include .env
export

CURRENT_DIR=$(shell pwd)
APP=backend
CMD_DIR=./cmd

.DEFAULT_GOAL = build

# generate swagger
.PHONY: swagger-gen
swagger-gen:
	swag init --dir ./internal -g ./app/server/server.go -o ./docs -ot yaml