APP_NAME=dp-analyzer
DOCKER_REPO=letitbeat
VERSION=`cat version`
.PHONY: test build 
.DEFAULT: help

help: ## Show Help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: test build ## Launch test and build targets

build: ## Build the application
	go build

test:  ## Launch tests
	go test -v ./...

fmt:  ## Run go fmt against the code
	go fmt ./...

vet:  ## Run go vet against the code	
	go vet ./...

lint: ## Run go lint against the code
	golint ./...

dbuild: ## Build the docker image
	@docker build --force-rm -t $(APP_NAME) .

release: tag login push
push:  ## push the image to docker hub
	@docker push $(DOCKER_REPO)/$(APP_NAME):$(VERSION)
login:
	@echo '$(DOCKER_PASSWORD)' | docker login -u '$(DOCKER_USERNAME)' --password-stdin
tag:  ## Tag the image
	@echo '>> Tagging image'
	@docker tag $(APP_NAME) $(DOCKER_REPO)/$(APP_NAME):$(VERSION)
vers: ## Output the current version
	@echo $(VERSION)
