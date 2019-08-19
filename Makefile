#
# Testing
#

COVEROUT  ?= cover.out
COVERMODE ?= count
TESTTAGS  ?=

.PHONY: test
ifdef VERBOSE
test: TESTFLAGS += -v
endif
ifdef RACE
test: TESTFLAGS += -race
endif
test: TESTFLAGS += -tags="$(TESTTAGS)"
test: TESTFLAGS += -coverprofile $(COVEROUT)
test: TESTFLAGS += -covermode $(COVERMODE)
test: ## Run go test
	go test ./... $(TESTFLAGS)

#
# Docker Compose
#

compose-up: ## Start docker-compose services, e.g for integration tests against AWS services
	docker-compose -f ./deployments/docker-compose.yml up -d

compose-down: ## Stop and remove containers
	docker-compose -f ./deployments/docker-compose.yml down
