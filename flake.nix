{
  description = "Zerops CLI utility";

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
        nativeBuildInputs = with pkgs; [ go ];
        shellHook = ''
          go mod vendor
          git add vendor/.
          echo -e '\033[0;33mprepared vendor files\033[0m'
        '';
      };
    };
  };
}
