{
  description = "ACLI development environment";

  nixConfig = {
    commit-lock-file-summary = "chore: update flake.lock";
  };

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/triplet";
  };

  outputs =
    {
      nixpkgs,
      systems,
      ...
    }:
    let
      inherit (nixpkgs) lib;
      forEachSystem = f: lib.genAttrs (import systems) (system: f nixpkgs.legacyPackages.${system});
    in
    {
      devShells = forEachSystem (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            golangci-lint
          ];
        };
      });

      formatter = forEachSystem (pkgs: pkgs.nixfmt-tree);

      checks = forEachSystem (pkgs: {
        dev-shell = pkgs.mkShell {
          packages = with pkgs; [
            go
            golangci-lint
          ];
        };
      });
    };
}
