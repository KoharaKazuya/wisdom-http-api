.PHONY: build deploy local-server

build:
	sam build

deploy: build
	sam deploy

local-server: build
	sam local start-api
