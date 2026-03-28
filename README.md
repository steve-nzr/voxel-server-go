# voxel-server-go

Experimental game server project to explore the **Actor Model** in Go while building a server compatible with the Minecraft protocol.

## Purpose

- Learn and validate actor-based architecture in a real networked system.
- Model Minecraft server behavior with clear message passing and actor isolation.
- Keep development reproducible and consistent across machines using Nix.

## Requirements

- Nix with flakes enabled.

## Quick Start (Nix)

```bash
# Enter the reproducible dev shell
nix develop

# Build the project package
nix build

# Run the app defined by the flake
nix run

# Run formatting/lint/test checks defined in flake outputs
nix flake check
```

Output
```bash
user@host$ nix run
INFO[0000] Waiting for actor system to start...         
INFO[0000] Server started with PID: goakt://voxel-server-go@127.0.0.1:0/server 
```

## Development Workflow

Inside the Nix shell (`nix develop`), you can use normal Go commands:

```bash
go test ./...
go vet ./...
go fmt ./...
```

Nix pins toolchain and dependencies so everyone gets the same versions.

## Updating Go Dependencies

When `go.mod`/`go.sum` changes, `vendorHash` in `flake.nix` must be updated:

1. Temporarily set `vendorHash = pkgs.lib.fakeHash;`
2. Run `nix build`
3. Copy the `got: sha256-...` value from the error
4. Paste it back into `vendorHash`
5. Run `nix build` again to confirm

## Notes

- This project is currently focused on architecture exploration, not production readiness.
- Protocol coverage and actor topology will evolve as experiments continue.
