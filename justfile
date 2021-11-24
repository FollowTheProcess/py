# Project directories and files

ROOT := justfile_directory()
PROJECT_BIN := "./bin"
GORELEASER_DIST := "./dist"
COVERAGE_DATA := "./coverage.out"
COVERAGE_HTML := "./coverage.html"

# Go build stuff

PROJECT_NAME := "py"
PROJECT_PATH := "github.com/FollowTheProcess/py"
PROJECT_ENTRY_POINT := "./cmd/py"
COMMIT_SHA := `git rev-parse HEAD`
VERSION_LDFLAG := PROJECT_PATH + "/cli.version"
COMMIT_LDFLAG := PROJECT_PATH + "/cli.commit"

# Docs

DOT_FILE := join(ROOT, "docs", "control_flow", "control_flow.dot")
DOT_FILE_NO_STEM := without_extension(DOT_FILE)
DOT_SVG := DOT_FILE_NO_STEM + ".svg"
DOT_PNG := DOT_FILE_NO_STEM + ".png"
MAN_MD := join(ROOT, "docs", "man_page", "py.1.md")
MAN_FILE := join(ROOT, "docs", "man_page", "py.1")

# By default print the list of recipes
_default:
    @just --list

# Show justfile variables
show:
    @just --evaluate

# Tidy up dependencies in go.mod and go.sum
tidy:
    go mod tidy

# Compile the project binary
build: tidy fmt
    go build -ldflags="-X {{ VERSION_LDFLAG }}=dev -X {{ COMMIT_LDFLAG }}={{ COMMIT_SHA }}" -o {{ PROJECT_BIN }}/{{ PROJECT_NAME }} {{ PROJECT_ENTRY_POINT }}

# Compile the project and run with debugging on
debug *args: build
    PY_PYTHON="" PYLAUNCH_DEBUG=1 {{ PROJECT_BIN }}/{{ PROJECT_NAME }} {{ args }}

# Run go fmt on all project files
fmt:
    gofumpt -extra -w .

# Run all project unit tests
test *flags: fmt
    go test -race ./... {{ flags }}

# Run all project benchmarks
bench: fmt
    go test ./... -bench=. -benchmem

# View a CPU/Memory profile (type = {cpu|mem})
pprof type:
    go tool pprof -http=:8000 {{ type }}.pprof

# Lint the project and auto-fix errors if possible
lint: fmt
    golangci-lint run --fix

# Calculate test coverage and render the html
cover:
    go test -race -cover -coverprofile={{ COVERAGE_DATA }} ./...
    go tool cover -html={{ COVERAGE_DATA }} -o {{ COVERAGE_HTML }}
    open {{ COVERAGE_HTML }}

# Remove build artifacts and other project clutter
clean:
    go clean ./...
    rm -rf {{ PROJECT_NAME }} {{ PROJECT_BIN }} {{ COVERAGE_DATA }} {{ COVERAGE_HTML }} {{ GORELEASER_DIST }} *.pprof

# Run unit tests and linting in one go
check: test lint

# Run all recipes (other than clean) in a sensible order
all: build test bench lint dot man

# Print lines of code (for fun)
sloc:
    find . -name "*.go" | xargs wc -l | sort -nr

# Build the control flow diagram as a SVG
_dot_svg:
    dot -T "svg" -o {{ DOT_SVG }} {{ DOT_FILE }}

# Build the control flow diagram as a PNG
_dot_png:
    dot -T "png" -o {{ DOT_PNG }} {{ DOT_FILE }}

# Convert the markdown-formatted man page to the man file format
_man-md:
    pandoc {{ MAN_MD }} --standalone -t man -o {{ MAN_FILE }}

# Build the man page
man: _man-md
    #!/usr/bin/env python3

    import datetime
    import pathlib

    with open("{{ MAN_FILE }}", "r", encoding="utf-8") as file:
        man_text = file.read()

    new_man_text = man_text.replace(
        "CURRENT_DATE", datetime.date.today().isoformat()
    )

    with open("{{ MAN_FILE }}", "w", encoding="utf-8") as file:
        file.write(new_man_text)

# Build the control flow diagram
dot: _dot_svg _dot_png

# Install the project on your machine
install: uninstall build
    cp {{ PROJECT_BIN }}/{{ PROJECT_NAME }} $GOBIN/

# Uninstall the project from your machine
uninstall:
    rm -rf $GOBIN/{{ PROJECT_NAME }}
