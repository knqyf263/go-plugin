{
  description = "bitmagnet dev shell";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-25.05";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = {
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs {
      system = system;
    };
    in {
      formatter = pkgs.alejandra;
      devShells = {
        default = pkgs.mkShell {
          hardeningDisable = [ "fortify" ]; 
          packages = with pkgs; [
            go
            golangci-lint
            protobuf
            protoc-gen-go
          ];
        };
      };
    });
}
