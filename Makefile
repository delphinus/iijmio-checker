# ref. http://postd.cc/auto-documented-makefile/
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build `iijmio-checker` binary
	rm -f tmpl.go
	go get github.com/gin-gonic/gin
	go get github.com/gin-contrib/sessions
	go get github.com/google/uuid
	go get github.com/jessevdk/go-assets
	go get gopkg.in/urfave/cli.v2
	go-assets-builder tmpl > tmpl.go
	go build -o iijmio-checker
