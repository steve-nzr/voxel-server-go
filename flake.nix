{
  description = "Description for the project";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    dagger.url = "github:dagger/nix";
    dagger.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = inputs@{ flake-parts, dagger, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        # To import an internal flake module: ./other.nix
        # To import an external flake module:
        #   1. Add foo to inputs
        #   2. Add foo as a parameter to the outputs function
        #   3. Add here: foo.flakeModule

      ];
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ];
      perSystem = { config, self', inputs', pkgs, system, ... }: {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = with pkgs; [
            go
            gopls
            delve
            golangci-lint
            nil
            wget
            git
            go
            dagger.packages.${system}.dagger
          ];
        };

        packages.default = pkgs.buildGoModule {
          pname = "gotcp";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-WrtuYUl6f8nbg6GIc697shOCoSyoYHJHJjASJmSDVTc=";
          ldflags = [ "-s" "-w" ];

          env = {
            CGO_ENABLED = "0";
          };
        };

        apps.default = {
          type = "app";
          program = "${self'.packages.default}/bin/voxel-server-go";
          meta = {
            description = "Voxel server application";
          };
        };

        checks.default = pkgs.runCommand "gotcp-checks" {
          nativeBuildInputs = [ pkgs.go pkgs.golangci-lint ];
        } ''
          export HOME=$TMPDIR
          cd ${./.}
          golangci-lint fmt --diff
          go fmt ./...
          go vet ./...
          golangci-lint run ./...
          go test ./...
          touch $out
        '';

        formatter = pkgs.nixpkgs-fmt;
      };
    };
}
