{
  description = "nix-search-tv";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};

        cmdPkg = "cmd/nix-search-tv";

        mkScript = name: text: pkgs.writeShellScriptBin name text;
        scripts = [
          (mkScript "build" "GOEXPERIMENT=jsonv2 go build -o $DEV_DIR/bin/ $CMD_DIR")
          (mkScript "run" "$DEV_DIR/bin/nix-search-tv $@ --config $DEV_DIR/config.json")
          (mkScript "print-search" "run print")
          (mkScript "preview-search" "run preview $@")

          (mkScript "test-integrations" "build && NIX_SEARCH_TV_BIN=$DEV_DIR/bin/nix-search-tv go test --count 1 -v ./tests/...")

          (mkScript "build-n-tv" "build && print-search | tv --preview-command 'preview-search {}'")
          (mkScript "build-n-fzf" "build && print-search | fzf --wrap --preview 'preview-search {}' --preview-window=wrap --scheme=history")
        ];
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_25
            brotli
            television
            fzf
            tmux
          ];

          buildInputs = scripts;

          shellHook = ''
            export PROJECT_ROOT=$(git rev-parse --show-toplevel)
            export DEV_DIR="$PROJECT_ROOT/.dev"
            export CMD_DIR="$PROJECT_ROOT/${cmdPkg}"
          '';
        };

        packages.default = pkgs.buildGo125Module {
          pname = "nix-search-tv";
          version = self.rev or "unknown";
          src = self;

          # If `nix shell` fails with "go: inconsistent vendoring", that's
          # likely due to outdated `vendorHash`.
          #
          # To find the new hash, uncomment below:
          # vendorHash = nixpkgs.lib.fakeHash;
          vendorHash = "sha256-8X+Vf2LovtnfQSYiVtDWp5R0aY0y2Tkhm7RF98cCYiw=";

          subPackages = [cmdPkg];

          env.GOEXPERIMENT = "jsonv2";

          meta = {
            description = "A tool integrating television and nix-search packages";
            homepage = "https://github.com/3timeslazy/nix-search-tv";
            mainProgram = "nix-search-tv";
          };
        };
      }
    );
}
