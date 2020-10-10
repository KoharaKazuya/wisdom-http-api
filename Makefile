.PHONY: build deploy local-server

build:
	sam build

deploy: build
	sam deploy

local-server: build
	sam local start-api --port=3001 --env-vars=local-server-envs.json
