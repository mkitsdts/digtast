# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go gRPC service for a digital labor container. Runtime entry is `main.go`, which currently starts the server on `:10086`. The gRPC server wiring is in `internal/server`, and RPC implementations live in `internal/server/api`.

Core domain code is under `internal/`: `internal/agent` owns LLM-backed agent creation, chat execution, session run cancellation, and prompt assembly; `internal/center` is the singleton registry for active agents plus shared FTP/gateway/display managers; `internal/memory` persists sessions as JSONL under the workspace path; `internal/vdisplay`, `internal/ftp`, and `internal/gateway` contain display, FTP, and external channel integrations. Reusable packages are under `pkg/`, including config, context management, queue, registry, workspace, model DTOs, and errors. Built-in tool integrations are in `tools/`. API contracts and generated Go bindings are in `proto/`; treat `*.pb.go` and `*_grpc.pb.go` as generated files. Deployment assets are under `deploy/`.

## Build, Test, and Development Commands
Use standard Go commands from the repository root:

- `go build ./...` builds all packages and catches compile errors.
- `go test ./...` runs the full suite, but the current baseline is not green; see "Current Gaps" before relying on it.
- `go test ./internal/memory ./pkg/queue ./pkg/registry ./pkg/ctxmanager` is a smaller useful loop for low-level packages.
- `go fmt ./...` formats Go source with the standard formatter.
- `go vet ./...` checks for common Go mistakes.
- `buf generate` regenerates gRPC code after editing [proto/container.proto](/Users/mkitsdts/code/digtast/proto/container.proto). This repo currently has `buf.gen.yaml` but no `buf.yaml`, so add or restore `buf.yaml` before expecting `buf generate` to work from a clean checkout.

The module currently declares Go `1.26.1` in `go.mod`; use that toolchain unless intentionally downgrading and testing compatibility.

## Coding Style & Naming Conventions
Follow idiomatic Go style: tabs for indentation, exported names in `CamelCase`, package-local helpers in `camelCase`, and short package names. Keep package boundaries clear: service and business logic belongs in `internal/`, reusable helpers in `pkg/`, and external tool implementations in `tools/`.

Do not hand-edit generated protobuf files. Update `proto/container.proto`, run generation, and review generated changes separately. Avoid hard-coded credentials, model keys, absolute user paths, or provider-specific test secrets in committed tests. Prefer explicit validation errors at API and manager boundaries instead of nil pointer panics or placeholder success responses.

## Testing Guidelines
Place tests next to the code they exercise using Go's `*_test.go` convention. Prefer table-driven tests for RPC handlers, manager validation, and package utilities. New features and bug fixes should include at least one focused test or a short PR note explaining why automated coverage is not practical yet.

Current test risks are part of the working baseline:

- `internal/agent/TestChat` calls a live Doubao endpoint with a hard-coded API key and passes an empty `sessionID`; convert this to a fake chat model or skip behind an explicit integration flag.
- `internal/center/TestCreateAgent_MissingKey` expects nil config validation, but `CreateAgent(nil)` currently panics.
- `internal/memory/TestGetMessages_ReturnsSnapshot` expects deep-copy semantics; `GetMessages` only copies the slice, not the pointed messages.
- `internal/vdisplay/TestVisualDisplayManagerDisplays` can panic because no VNC display implementation is registered before use.
- `go test ./...` may continue running after these failures, so isolate failing packages while stabilizing the suite.

## Current Gaps
Many RPCs are still stubs. `StopService`, `RestartService`, `BackupService`, Task RPCs, Skill RPCs, Tool RPCs, and MCP RPCs currently return success, fixed IDs, or fixed backup URLs without durable behavior. `RemoveSession` and `CompressSession` are not fully wired to the memory store. Treat these as unimplemented API surface, not production behavior.

Session and context lifecycle need tightening. `SendMessageToSession` creates a session ID only for empty requests, but stream responses do not return the session ID, making client-side discovery difficult. `ctxmanager` stores background contexts without cancel functions, while `DigitalAgent` separately tracks cancel funcs by session. Align this design before adding long-running task control.

Memory persistence needs correctness hardening. `workspace.GetWorkspacePath()` currently resolves to the user's home directory, so sessions are written under `$HOME/sessions`, not a repo-local or container-scoped data directory. `internal/memory.Store.AppendMessage` can panic for missing cache entries or empty message slices, and message merging is incomplete.

Provider configuration is incomplete. Default base URLs and model names in `internal/agent/chatmodel.go` are empty, `config.yaml` is empty, and `Center.GetCenter()` still has a TODO for config parsing. Startup should load explicit config and fail clearly when provider settings are missing.

## Recommended Next Steps
Stabilize correctness before expanding features:

1. Make `go test ./...` deterministic and green without external network calls or real model credentials.
2. Replace stub RPC success paths with real state changes or explicit `codes.Unimplemented` errors so clients cannot mistake placeholders for working behavior.
3. Fix session lifecycle: return or surface generated session IDs for streamed messages, implement `RemoveSession`, and define where session data is stored.
4. Add config loading for model providers, workspace paths, FTP, VNC, and gateway channels.
5. Wire VNC/FTP/gateway startup and shutdown through `StartService`, `StopService`, `RestartService`, and `RemoveService`.

## Commit & Pull Request Guidelines
Use short, imperative commit subjects such as `validate center agent config` or `wire session removal`. Keep commits scoped to one behavior change. PRs should include a concise summary, affected paths, test results, and API examples when gRPC behavior changes. If `go test ./...` is not green, list the exact failing packages and why they are unrelated or intentionally deferred.
