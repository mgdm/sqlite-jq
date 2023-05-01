{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go
    pkgs.gotools
    pkgs.gopls
    pkgs.go-outline
    pkgs.gocode
    pkgs.gopkgs
    pkgs.gocode-gomod
    pkgs.godef
    pkgs.golint
    pkgs.sqlite-interactive
    pkgs.gojq
    pkgs.golangci-lint
  ];
}
