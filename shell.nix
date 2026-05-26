# Compatibility shim for `nix-shell` users.
# For flake-aware Nix, use `nix develop` instead.
(builtins.getFlake (toString ./.)).devShells.${builtins.currentSystem}.default
