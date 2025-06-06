GO_TEST = go tool gotest.tools/gotestsum --format pkgname

#   🔨 TOOLS       #
##@ Tools

prep: prep/tools

prep/tools:

	@if ! command -v lefthook >/dev/null 2>&1; then \
		echo "lefthook is not installed. Installing lefthook..."; \
		curl -1sLf https://gobinaries.com/evilmartians/lefthook/install | sh; \
	fi

	@if [ ! -f .lefthook.yml ]; then \
		echo "Installing lefthook configuration..."; \
		lefthook install; \
	fi

prep/tools/mockgen:
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "mockgen is not installed. Installing mockgen..."; \
		go install go.uber.org/mock/mockgen@latest; \
	fi

#   🧹 Formatting   #
##@ Formatting

fmt:
	go tool mvdan.cc/gofumpt -w .

fmt/check:
	go tool mvdan.cc/gofumpt -d .

#   🔍 Linting     #
##@ Linting

lint:
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run ./...

lint/fix:
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --fix ./...

#   ⛹🏽‍ License   #
##@ License

license: license/headers/check

license/headers/check:
	go tool github.com/hashicorp/copywrite headers --plan

license/headers/apply:
	go tool github.com/hashicorp/copywrite headers

#   🧪 Testing     #
##@ Testing

test/ci: test/unit

test/unit:
	mkdir -p build/reports
	$(GO_TEST) --junitfile build/reports/test-unit.xml -- -race ./... -count=1 -short -cover -coverprofile build/reports/unit-test-coverage.out

test/coverage: test/unit
	go tool cover -func=build/reports/unit-test-coverage.out

test/coverage/html: test/unit
	go tool cover -html=build/reports/unit-test-coverage.out -o build/reports/coverage.html

#   🏗️ Building    #
##@ Building

build: license/headers/check
	go build ./...

clean:
	rm -rf build/
	go clean

#   🚀 All-in-One  #
##@ All

all: fmt lint test/unit

#   🔄 Git Hooks   #
##@ Git Hooks

hooks/install:
	lefthook install

hooks/uninstall:
	lefthook uninstall