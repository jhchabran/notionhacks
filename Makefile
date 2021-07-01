NAME := nn

.PHONY: all $(NAME) test build

all: $(NAME)

build: $(NAME)

$(NAME):
	cd ./cmd/$@ && go install

test:
	go test -cover -timeout=1m ./...
	cd cmd/nn && go test -cover -timeout=1m ./...

