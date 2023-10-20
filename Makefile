GOARCH = amd64

UNAME = $(shell uname -s)

VERSION = $(shell cat VERSION)

BIN_NAME = server_start

BUILD_DIRECTORY = build

ifndef OS
	ifeq ($(UNAME), Linux)
		OS = linux
	else ifeq ($(UNAME), Darwin)
		OS = darwin
	endif
endif

.DEFAULT_GOAL := all

all: fmt $(BUILD_DIRECTORY)/$(BIN_NAME)

$(BUILD_DIRECTORY)/$(BIN_NAME): api.go ./**/*.go
	mkdir -p build
	GOOS=$(OS) GOARCH="$(GOARCH)" go build -o $(BUILD_DIRECTORY)/$(BIN_NAME) api.go

app.yaml: app.dev.yaml check-env
	@cat app.dev.yaml \
	| sed "s#{{SUPABASE_URL}}#${SUPABASE_URL}#" \
	| sed "s/{{SUPABASE_KEY}}/${SUPABASE_KEY}/" \
	| sed "s/{{OPENAI_API_KEY}}/${OPENAI_API_KEY}/" \
	| sed "s/{{COHERE_API_KEY}}/${COHERE_API_KEY}/" \
	| sed "s/{{OPENAI_ORGANIZATION}}/${OPENAI_ORGANIZATION}/" \
	| sed "s/{{POSTHOG_API_KEY}}/${POSTHOG_API_KEY}/" \
	| sed "s/{{REPLICATE_API_KEY}}/${REPLICATE_API_KEY}/" \
	| sed "s#{{LLAMA_URL}}#${LLAMA_URL}#" \
	| sed "s#{{POSTGRES_URI}}#${POSTGRES_URI}#" \
	| sed "s#{{API_URL}}#${API_URL}#" \
	| sed 's/{{ELEVENLABS_API_KEY}}/${ELEVENLABS_API_KEY}/' \
	| sed "s/{{JWT_SECRET}}/${JWT_SECRET}/" > app.yaml

check-env:
ifndef SUPABASE_URL
	$(error SUPABASE_URL is undefined)
endif
ifndef SUPABASE_KEY
	$(error SUPABASE_KEY is undefined)
endif
ifndef OPENAI_API_KEY
	$(error OPENAI_API_KEY is undefined)
endif
ifndef COHERE_API_KEY
	$(error COHERE_API_KEY is undefined)
endif
ifndef OPENAI_ORGANIZATION
	$(error OPENAI_ORGANIZATION is undefined)
endif
ifndef POSTHOG_API_KEY
	$(error POSTHOG_API_KEY is undefined)
endif
ifndef REPLICATE_API_KEY
	$(error REPLICATE_API_KEY is undefined)
endif
ifndef JWT_SECRET
	$(error JWT_SECRET is undefined)
endif
ifndef POSTGRES_URI
	$(error POSTGRES_URI is undefined)
endif
ifndef LLAMA_URL
	$(error LLAMA_URL is undefined)
endif
ifndef API_URL
	$(error API_URL is undefined)
endif
ifndef ELEVENLABS_API_KEY
	$(error ELEVENLABS_API_KEY is undefined)
endif

deploy: app.yaml
	gcloud app deploy --quiet --version v0

clean:
	rm -f ./build/* app.yaml

fmt:
	go fmt $$(go list ./...)

.PHONY: clean fmt check-env deploy
