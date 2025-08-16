type: noogle-cli:
{lib, config, pkgs, system, ...}:
let
  inherit (lib) mkOption mkEnableOption types last head splitString concatLines mapAttrsToList mkIf replaceStrings optionalString optionals optional;
  options =  {
    programs.nix-search-tv-script = {
      enable = mkEnableOption "Wether to include the nix-serach-tv-script";
      outputPackage = mkOption {
        description = "when enabled this option provides the outputPackage with the settings";
        type = types.package;
      };
      outputPackageName = mkOption {
        type = types.str;
        description = "name of the package, use this to run the command";
        default = "nixpkgs-sh";
      };
      settings = {
        indexes = {
          real = mkOption {
             default = {
               "nixpkgs" = "ctrl-n";
               "home-manager" = "ctrl-h";
             };
             type = with types; attrsOf str;
        };
         pseudo = mkOption {
           default = {
             "all" = "ctrl-a";
           };
           type = with types; attrsOf str;
         };
        };
        noogleEnable = mkEnableOption "wether to enable noogle-cli integration";
        keys = {
          searchSnippet = mkOption {
            default = "ctrl-w";
            type = types.str;
          };
          openSource = mkOption {
            default = "ctrl-s";
            type = types.str;
          };
          openHomepage = mkOption {
             default = "ctrl-o";
             type = types.str;
          };
          nixShell = mkOption {
            default = "ctrl-i";
            type = types.str;
          };
          printPreview = mkOption {
            default = "alt-p";
            type = types.str;
          };
          noogle = mkOption {
            default = "ctrl-l";
            type = types.str;
          };

        };
        opener = mkOption {
           default = "xdg-open";
           type = with types; either package str;
        };
        stateFile = mkOption {
          default = "/tmp/nix-search-tv-fzf";
          type = types.path;
          
        };
      };

    };
  };
  cfg = config.programs.nix-search-tv-script; 
  script_text = builtins.readFile (
    "${  pkgs.applyPatches {
      src = ./.;
        patches = optional cfg.settings.noogleEnable 
          (pkgs.fetchpatch2 {
            name="noogle-integration";
            url="https://github.com/haennes/nix-search-tv/commit/e45c13254c5a5181d8c0af67fc18106fcdd8cb60.patch";
            hash="sha256-TKaW8IYd5m3+WYRJ1rJK4Srdb/opHJ5GS/G+CGwaluY=";
          })
        ;
        }
    }/nixpkgs.sh");
  script_without_config_tail= (last (splitString "# ========================================" script_text));
  script_without_config_head = (head (splitString "# === Change keybinds or add more here ===" script_text));
  map_indexes_to_config = idxs:  concatLines (mapAttrsToList (name: value: "\"${name} ${value}\"") idxs);
  out_pkg_list = [ cfg.outputPackage ];
  noogle-cli-pkg = noogle-cli.packages.${system}.noogle-cli;

in
{
  inherit options;
  config = mkIf cfg.enable  ({
    programs.nix-search-tv-script.outputPackage = pkgs.writeShellApplication {
      
      name = cfg.outputPackageName;
      excludeShellChecks = [ "SC2016" ];
      runtimeInputs = [ pkgs.fzf] ++ (optionals cfg.settings.noogleEnable (with pkgs; [glow noogle-cli-pkg jq]));
      text =
      script_without_config_head +
    ''
      declare -a INDEXES=(
          ${ map_indexes_to_config cfg.settings.indexes.real }

          # you can add any indexes combination here,
          # like `nixpkgs,nixos`
          ${ map_indexes_to_config cfg.settings.indexes.pseudo }

      )

      SEARCH_SNIPPET_KEY=${cfg.settings.keys.searchSnippet}
      OPEN_SOURCE_KEY=${cfg.settings.keys.openSource}
      OPEN_HOMEPAGE_KEY=${cfg.settings.keys.openHomepage}
      NIX_SHELL_KEY=${cfg.settings.keys.nixShell}
      PRINT_PREVIEW_KEY=${cfg.settings.keys.printPreview}
      ${optionalString cfg.settings.noogleEnable ''NOOGLE_KEY="${cfg.settings.keys.noogle}"''}

      OPENER="${cfg.settings.opener}"
    '' + (
      replaceStrings [
        "/tmp/nix-search-tv-fzf"
      ] [
        cfg.settings.stateFile
      ] script_without_config_tail);
    
    };
  } // (if type == "home-manager" then {
      home.packages = out_pkg_list;
    } else if types == "nixos" then {
      environment.systemPackages = out_pkg_list;
  } else (abort "select either home-manager or nixos as target")));
}
