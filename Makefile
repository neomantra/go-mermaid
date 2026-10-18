.PHONY: build test cover lint fmt golden tidy clean examples-test wasm wasm-site check-site serve-site clean-site notices notices-check

build: ## Build the CLI
	go build -o bin/mermaid ./cmd/mermaid

test: ## Run tests with race detector
	go test -race ./...

cover: ## Run tests and open an HTML coverage report
	go test -coverprofile=coverage.out -coverpkg=./... ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "wrote coverage.html"

lint: ## Run golangci-lint (must be installed)
	golangci-lint run

fmt: ## Format code
	gofmt -s -w .
	go run golang.org/x/tools/cmd/goimports@latest -w .

golden: ## Regenerate golden SVG files
	go test -run TestGolden -update .

tidy: ## Tidy modules
	go mod tidy
	GOWORK=off go -C examples mod tidy

clean: clean-site ## Remove build artifacts
	rm -rf bin coverage.out coverage.html

# Demos: examples/ is a separate module that owns go-booba and the UI
# dependencies, so the library's go.mod stays lean. Native demo commands set
# GOWORK=off; browser builds use wasm.work, which joins both modules and
# selects the Bubble Tea and clipboard forks that js/wasm needs.

WASM_ENV := GOWORK=$(CURDIR)/wasm.work GOOS=js GOARCH=wasm CGO_ENABLED=0

examples-test: ## Test and vet the demos natively
	GOWORK=off go -C examples test -race ./...
	GOWORK=off go -C examples vet ./...

wasm: ## Build the browser demos into web/demos/
	mkdir -p web/demos
	$(WASM_ENV) go -C examples build -mod=readonly -o ../web/demos/editor.wasm ./editor
	$(WASM_ENV) go -C examples build -mod=readonly -ldflags='-s -w' -o ../web/demos/render.wasm ./render

wasm-site: wasm ## Build the demo site in web/ with the pinned booba runtime
	GOWORK=off go -C examples tool booba-assets --no-index ../web/
	cp LICENSE web/LICENSE
	cp examples/THIRD_PARTY_NOTICES.md web/THIRD_PARTY_NOTICES.md

check-site: ## Check the built site and exercise both demos under Node
	node --check web/site.js
	node --check web/playground.js
	node examples/scripts/check-site.mjs
	node examples/scripts/smoke-render.mjs
	node examples/scripts/smoke-editor.mjs

serve-site: wasm-site ## Serve web/ at http://localhost:8000/
	python3 -m http.server 8000 --bind 127.0.0.1 --directory web

clean-site: ## Remove generated demo site files
	rm -rf web/booba web/ghostty-web web/demos
	rm -f web/wasm_exec.js web/LICENSE web/THIRD_PARTY_NOTICES.md

notices: ## Regenerate the demos' dependency licenses
	GOWORK=off go -C examples run ./internal/notices > examples/THIRD_PARTY_NOTICES.md

notices-check: ## Verify the demos' dependency licenses are current
	GOWORK=off go -C examples run ./internal/notices | cmp -s - examples/THIRD_PARTY_NOTICES.md

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
