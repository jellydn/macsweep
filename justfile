default:
    @just --list

check:
    ./scripts/check.sh

test:
    go test -race ./...

format:
    gofmt -w cmd internal

dev:
    go tool mygo dev

build:
    CGO_ENABLED=0 go tool mygo build -platform darwin/universal -skip-notarize

doctor:
    go tool mygo doctor
