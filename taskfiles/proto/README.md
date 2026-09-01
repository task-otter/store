# Proto Taskfile

A [TaskOtter](https://github.com/task-otter/store) module for generating Go and gRPC source files from [Protocol Buffer](https://protobuf.dev/) definitions using [protoc](https://github.com/protocolbuffers/protobuf).

## What is this Taskfile?

This module generates Go and gRPC source files from `.proto` files. The `gen`
task auto-installs `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` via
`nix:install:profile`.

## Usage

### Standalone

```sh
task -t taskfiles/proto/Taskfile.yml gen
```

Install only, without generating:

```sh
task nix:install:profile NIX_INSTALLABLE="nixpkgs#protobuf nixpkgs#protoc-gen-go nixpkgs#protoc-gen-go-grpc"
```

Generate from a specific proto directory and keep generated files relative to
the Go module root:

```sh
task -t taskfiles/proto/Taskfile.yml gen \
  PROTO_PATH=api \
  PROTO_PATTERN="v1/*.proto" \
  GO_MODULE=github.com/example/project
```

`GO_MODULE` must match the module-path prefix used by the proto files'
`go_package` options. Leave it empty when module-relative output is not needed.

Remove generated files before regenerating:

```sh
task -t taskfiles/proto/Taskfile.yml ungen
task -t taskfiles/proto/Taskfile.yml gen
```

### Included in your Taskfile

```yaml
includes:
  proto:
    taskfile: taskfiles/proto/Taskfile.yml
```

Then run:

```sh
task proto:gen
task proto:ungen
```

## Public Tasks

| Task | Description |
|---|---|
| `gen` | Generate Go files from proto definitions |
| `ungen` | Remove generated protobuf (.pb.go) files from the working tree |

## Variables

| Variable | Default | Description |
|---|---|---|
| `PROTO_NIX_INSTALLABLE` | `nixpkgs#protobuf nixpkgs#protoc-gen-go nixpkgs#protoc-gen-go-grpc` | Flake installables passed to `nix:install:profile` |
| `GO_MODULE` | `""` | Module path stripped from generated output paths |
| `PROTO_PATH` | `"."` | Search root and value passed to protoc `--proto_path` |
| `PROTO_PATTERN` | `"*.proto"` | `find -name` pattern for discovering .proto source files |

Pin a revision by overriding the installable, for example
`PROTO_NIX_INSTALLABLE=github:NixOS/nixpkgs/<rev>#protobuf`.

## Notes

- Install goes through `nix:install:profile` (Nix is installed first if missing). Native Windows is not supported; use WSL2.
- `GO_MODULE`, `PROTO_PATH`, and `PROTO_PATTERN` are declared as top-level vars
  here, which outrank vars supplied by an inclusion: set them on the command
  line or from the environment.
