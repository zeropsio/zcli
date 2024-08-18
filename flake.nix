{
  description = "Arbel's zcli flake";

  inputs = {
      nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
      flake-parts.url = "github:hercules-ci/flake-parts";
    };

  outputs = { self, nixpkgs, ... }@inputs:
  inputs.flake-parts.lib.mkFlake { inherit inputs; } {
    flake = {
    };
    systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" "x86_64-darwin" ];
    perSystem = { config, pkgs, system, ... }: {
      packages.default = (import ./default.nix { inherit pkgs self; });
      devShells.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [ got ];
      };
    };
  };
}
