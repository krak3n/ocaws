COVEROUT  ?= cover.out
COVERMODE ?= atomic
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
