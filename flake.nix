{
  description = "SQLite extension and shell with jq query support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        inherit (pkgs) lib;

        # SQLite 3.53.0 amalgamation — provides shell.c, sqlite3.c, and headers.
        # On first build this will fail and print the correct hash; paste it here.
        sqliteAmalgamation = pkgs.fetchzip {
          url = "https://www.sqlite.org/2026/sqlite-amalgamation-3530000.zip";
          hash = "sha256-Ai10AJEWtasenkFmBGzlZ3S1pf+AE2fvRP8qak4SoMQ=";
        };

        # Common Go module settings shared by both packages.
        # vendorHash: run `nix build .#sqlite-jq` once; replace with the value
        # printed in "got: sha256-..." in the error output.
        goAttrs = {
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-LViDOF/UzybIGq73ydfdoT0j7DN4IglL3+4u84Bc2rE=";
        };

      in {
        packages = {
          # Loadable extension (.dylib on macOS, .so on Linux).
          sqlite-jq = pkgs.buildGoModule (goAttrs // {
            pname = "sqlite-jq";

            buildPhase =
              let ext = if pkgs.stdenv.isDarwin then "dylib" else "so";
              in ''
                runHook preBuild
                go build -buildmode=c-shared -o sqlite_jq.${ext} ./*.go
                runHook postBuild
              '';

            installPhase = ''
              runHook preInstall
              mkdir -p $out/lib
              find . -maxdepth 1 \( -name 'sqlite_jq.so' -o -name 'sqlite_jq.dylib' \) \
                -exec install -Dm755 {} $out/lib/ \;
              install -Dm644 sqlite_jq.h $out/include/sqlite_jq.h 2>/dev/null || true
              runHook postInstall
            '';
          });

          # Standalone sqlite3 shell with jq built in — no .load required.
          sqlite3-jq = pkgs.buildGoModule (goAttrs // {
            pname = "sqlite3-jq";

            buildPhase = ''
              runHook preBuild

              # Make the amalgamation available at the relative path init_shim.c expects.
              mkdir -p sqlite
              cp ${sqliteAmalgamation}/{sqlite3.c,sqlite3.h,sqlite3ext.h,shell.c} sqlite/

              # Build Go static archive.
              go build -buildmode=c-archive -o sqlite_jq.a ./*.go

              # Link the standalone shell.
              $CC -o sqlite3-jq \
                -DSQLITE_SHELL_INIT_PROC=register_jq_extension \
                sqlite/sqlite3.c \
                sqlite/shell.c \
                standalone/init_shim.c \
                sqlite_jq.a \
                $(go env CGO_LDFLAGS) \
                -lpthread \
                ${lib.optionalString pkgs.stdenv.isLinux "-ldl"} \
                -Wno-deprecated-declarations

              runHook postBuild
            '';

            installPhase = ''
              runHook preInstall
              install -Dm755 sqlite3-jq $out/bin/sqlite3-jq
              runHook postInstall
            '';
          });

          default = self.packages.${system}.sqlite3-jq;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools        # goimports, etc.
            golangci-lint
            sqlite-interactive
            gojq
          ];
        };
      }
    );
}
