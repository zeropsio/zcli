{
  description = "Arbel's zcli flake";

  inputs = {
      nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    };

  outputs = { self, nixpkgs, ... }@inputs:
  let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};
  in
  {
    packages.${system}.default = (import ./default.nix { inherit pkgs self; });

    devShells = pkgs.mkShell {
        nativeBuildInputs = [ pkgs.go ];
      };
  };
}
