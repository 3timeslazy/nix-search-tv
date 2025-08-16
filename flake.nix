{
  description = "nix-search-tv";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    unf = {
      url = "git+https://git.atagen.co/atagen/unf";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    agenix.url = "github:ryantm/agenix";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    unf,
    agenix,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};

        cmdPkg = "cmd/nix-search-tv";

        mkScript = name: text: pkgs.writeShellScriptBin name text;
        scripts = [
          (mkScript "build" "go build -o $DEV_DIR/bin $CMD_DIR")
          (mkScript "run" "$DEV_DIR/bin/nix-search-tv $@ --config $DEV_DIR/config.json")
          (mkScript "print-search" "run print")
          (mkScript "preview-search" "run preview $@")

          (mkScript "test-integrations" "build && NIX_SEARCH_TV_BIN=$DEV_DIR/bin/nix-search-tv go test --count 1 -v ./integrations/...")

          (mkScript "build-n-tv" "build && print-search | tv --preview-command 'preview-search {}'")
          (mkScript "build-n-fzf" "build && print-search | fzf --wrap --preview 'preview-search {}' --preview-window=wrap --scheme=history")
        ];

        mkModuleOpts = module:
          unf.lib.json {
            inherit self;
            pkgs = nixpkgs.legacyPackages.${system};

            # not all modules can be evaluated that easy. If your module
            # does not evaluate, try checking this NüschtOS file:
            #   https://github.com/NuschtOS/search.nuschtos.de/blob/main/flake.nix
            modules = [module];
          };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_23
            gopls
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

        packages.agenix-opts = mkModuleOpts agenix.nixosModules.default;

        packages.default = pkgs.buildGo123Module {
          pname = "nix-search-tv";
          version = self.rev or "unknown";
          src = self;

          # If `nix shell` fails with "go: inconsistent vendoring", that's
          # likely due to outdated `vendorHash`.
          #
          # To find the new hash, uncomment below:
          # vendorHash = nixpkgs.lib.fakeHash;
          vendorHash = "sha256-RcDoQvXgyWEQWCBHgk9/ms4RoWcKYPte77eONOTkn5k=";

          subPackages = [cmdPkg];

          meta = {
            description = "A tool integrating television and nix-search packages";
            homepage = "https://github.com/3timeslazy/nix-search-tv";
            mainProgram = "nix-search-tv";
          };
        };
      }
    );
}
