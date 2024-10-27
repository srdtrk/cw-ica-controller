{ pkgs ? import <nixpkgs> {}, lib ? pkgs.lib, stdenv ? pkgs.stdenv }:
let
  unstable = import
    (builtins.fetchTarball https://github.com/nixos/nixpkgs/tarball/ccc0c2126893dd20963580b6478d1a10a4512185)
    # reuse the current configuration
    { config = pkgs.config; };
in
  pkgs.mkShell {
    nativeBuildInputs = with pkgs.buildPackages; [ 
      just unstable.golangci-lint unstable.go rustup
    ];
    # Run a command after entering the shell
    shellHook = ''
      echo "Entering shell with stable rust"
      rustup toolchain install stable
    '';
}
