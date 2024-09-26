{
  description = "Flake for development on the bosh-cli featuring a nix-shell";

  inputs = {
    nixpkgsRepo.url = github:NixOS/nixpkgs/nixos-24.05;
  };

  outputs = { self, nixpkgsRepo }:
    let
      nixpkgsLib = nixpkgsRepo.lib;
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgsLib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgsRepo { inherit system; });
    in {
      devShells = forAllSystems (system:
        let
          nixpkgs = nixpkgsFor.${system};
        in {
          default = nixpkgs.mkShell {
            buildInputs = with nixpkgs; [
              delve
              go
              gopls
              gotools
            ];
          };
        });
    };
}
