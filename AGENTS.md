# Repository Guidelines

## Project Structure
This repository is a Go gRPC service for a digital labor container. The runtime entry is [main.go](/Users/mkitsdts/code/digtast/main.go), and the gRPC server setup is under [internal/server](/Users/mkitsdts/code/digtast/internal/server). RPC implementations live in [internal/server/api](/Users/mkitsdts/code/digtast/internal/server/api).

Core domain code is under `internal/`: `internal/agent` handles LLM-backed agent creation, chat execution, session cancellation, and prompt assembly; `internal/center` owns the active-agent registry plus shared FTP, gateway, and display managers; `internal/memory` and `pkg/workspace` cover session and workspace persistence; `internal/vdisplay`, `internal/ftp`, and `internal/gateway` contain external service integrations.

Reusable packages are under `pkg/`, built-in tool integrations are under `tools/`, prompt implementations are under `prompt/`, and deployment assets are under `deploy/`. The source proto currently lives at [labor/v1/container.proto](/Users/mkitsdts/code/digtast/labor/v1/container.proto). Generated Go bindings are currently under `proto/`; treat `*.pb.go` and `*_grpc.pb.go` as generated files and do not hand-edit them.

## Build and Test Commands
Use standard Go commands from the repository root:

- `go build ./...` builds all packages and catches compile errors.
- `go test ./...` runs the full test suite. The baseline may still fail because several tests depend on external services or incomplete behavior.
- `go test ./internal/memory ./pkg/queue ./pkg/registry ./pkg/ctxmanager` is a smaller loop for low-level packages.
- `go fmt ./...` formats Go source.
- `go vet ./...` checks for common Go mistakes.
- `buf generate` should regenerate gRPC code after proto edits, but this repo currently has `buf.gen.yaml` without `buf.yaml`; restore or add `buf.yaml` before relying on generation from a clean checkout.

The module declares Go `1.26.1` in `go.mod`; use that toolchain unless intentionally changing compatibility.

## Code Style
Write idiomatic, simple Go. Prefer small functions, explicit data flow, and clear validation over broad abstractions. Keep package boundaries clear: service orchestration belongs in `internal/server/api`, domain behavior in `internal/`, reusable helpers in `pkg/`, and external tool adapters in `tools/`.

Use Go naming conventions: exported names in `CamelCase`, package-local names in `camelCase`, concise package names, and constants that follow normal Go style rather than snake_case. Keep comments useful and sparse; comment behavior or constraints that are not obvious from the code.

Use `log/slog` for logging. Do not introduce `fmt.Println`, the standard `log` package, `logrus`, or other logging APIs in production code unless there is a specific compatibility reason. Prefer structured attributes, for example `slog.Error("failed to start ftp server", "port", port, "error", err)`.

Return explicit errors at API and manager boundaries. Avoid nil pointer panics, placeholder success responses, hard-coded credentials, provider-specific test secrets, and absolute user paths. Keep new code readable first, then extensible where there is a concrete extension point.

## Protobuf and API Changes
Update [labor/v1/container.proto](/Users/mkitsdts/code/digtast/labor/v1/container.proto) for contract changes, then regenerate bindings. Review generated files separately from hand-written code. Because the current proto source path, `go_package`, and generated output location are not fully aligned, verify imports and generated package names before committing API changes.

For unimplemented RPC behavior, return an explicit error such as `codes.Unimplemented` instead of a fake success response. Clients should not be able to mistake a stub for durable behavior.

## Testing Guidelines
Place tests next to the code they exercise using Go's `*_test.go` convention. Prefer table-driven tests for RPC handlers, manager validation, memory behavior, and utility packages. New features and bug fixes should include focused tests; if automated coverage is not practical, document the reason and the manual verification performed.

Tests must be deterministic by default. Live model calls, real network dependencies, VNC/FTP environment requirements, and provider credentials should be hidden behind explicit integration flags or replaced with fakes.

## Current Problems
The current codebase has several known issues that should be considered before expanding features:

- Test stability is not guaranteed. `internal/agent/TestChat` performs a live Doubao call and uses a package-level API key; it should use a fake model or be gated as an integration test.
- Some tests describe behavior that the implementation does not yet satisfy, including nil config validation in `internal/center`, memory snapshot/deep-copy semantics, and VNC manager registration/setup.
- `RemoveSession` and `CompressSession` still return success without real memory-store behavior.
- `BackupService` returns a fixed-looking backup URL and does not perform a real backup.
- `SendMessageToSession` creates a session ID for empty requests, but stream responses do not expose that generated ID, which makes client-side discovery difficult.
- Session cancellation and context ownership are split between `ctxmanager` and `DigitalAgent`; long-running task lifecycle should be consolidated.
- Workspace and memory persistence need hardening. Session data may be written outside a container-scoped data directory, and some memory append/merge paths need stronger validation.
- Provider configuration is incomplete. Startup should load explicit model provider, workspace, FTP, VNC, and gateway config and fail clearly when required settings are missing.
- Gateway integrations such as Feishu and Telegram still contain TODO-only placeholders.
- VNC start/stop logic is rigid and needs better environment detection, lifecycle handling, and error reporting.
- Proto generation is inconsistent: the proto source is under `labor/v1`, generated bindings are under `proto`, and `buf.yaml` is missing.

## Recommended Priorities
Stabilize correctness before adding new surface area:

1. Make `go test ./...` deterministic and green without external credentials or live network calls.
2. Replace placeholder RPC success paths with real behavior or explicit unimplemented errors.
3. Fix session lifecycle: generated session IDs, cancellation, removal, compression, and persistence should have one coherent ownership model.
4. Normalize proto generation and generated package layout.
5. Add explicit config loading and validation for providers, workspace paths, FTP, VNC, and gateway channels.

## Commit and PR Guidelines
Use short, imperative commit subjects such as `validate center agent config` or `wire session removal`. Keep commits scoped to one behavior change. PRs should include a concise summary, affected paths, test results, and API examples when gRPC behavior changes. If `go test ./...` is not green, list the exact failing packages and explain whether they are related to the change.
