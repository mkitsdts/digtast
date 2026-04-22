# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go service for a digital labor container. Entry code starts at `main.go`. Core application logic lives under `internal/`: `internal/server` for gRPC service handlers, `internal/agent` for session and message flow, `internal/memory` for memory management, and `internal/registry` for registration concerns. Reusable packages live under `pkg/`, including config, context management, errors, and tool abstractions. API contracts and generated gRPC bindings live under `proto/`; treat `*.pb.go` and `*_grpc.pb.go` as generated files. Tool integrations are under `tools/`, and deployment assets are under `deploy/`.

## Build, Test, and Development Commands
Use standard Go commands from the repository root:

- `go build ./...` builds all packages and catches compile errors.
- `go test ./...` runs the full test suite; add tests before relying on this in CI.
- `go fmt ./...` formats Go source with the standard formatter.
- `go vet ./...` checks for common Go mistakes.
- `buf generate` regenerates gRPC code after editing [proto/container.proto](/Users/mkitsdts/code/digtast/proto/container.proto).

Run these before opening a PR when you change Go code or protobuf definitions.

## Coding Style & Naming Conventions
Follow idiomatic Go style: tabs for indentation, exported names in `CamelCase`, package-local helpers in `camelCase`, and concise package names such as `errs` or `conf`. Keep package boundaries clear: business logic belongs in `internal/`, reusable helpers in `pkg/`. Do not hand-edit generated protobuf files; update the `.proto` file and rerun `buf generate`.

## Testing Guidelines
Place tests next to the code they exercise using Go’s `*_test.go` convention, for example `internal/server/api/service_test.go`. Prefer table-driven tests for request/response handlers and package-level unit tests for `pkg/` utilities. New features and bug fixes should include at least one focused test or a brief note explaining why automated coverage is not practical yet.

## Commit & Pull Request Guidelines
Current history is minimal and inconsistent (`fix last error commit`, `temp commit`). Use short, imperative commit subjects instead, such as `add session removal validation`. Keep commits scoped to one change. PRs should include a clear summary, affected paths, test results from `go test ./...`, and API examples or screenshots when behavior changes are user-visible.
