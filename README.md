# nix-search-tv

Fuzzy search for NixOS packages.

---

[![asciicast](https://asciinema.org/a/afNYMXrhoEwwh3wzOK7FbsFtW.svg)](https://asciinema.org/a/afNYMXrhoEwwh3wzOK7FbsFtW)

<div>
    <a href="https://codeberg.org/3timeslazy/nix-search-tv">
        <img alt="Get it on Codeberg" src="https://img.shields.io/badge/Codeberg-2184D0?style=for-the-badge&logo=Codeberg&logoColor=white" height="60">
    </a>
    <a href="https://github.com/3timeslazy/nix-search-tv">
        <img alt="Get it on GitHub" src="https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white" height="60">
    </a>
</div>

## Searchable Things

Out of the box, it is possible to search for things from:

- [Nixpkgs](https://search.nixos.org/packages?channel=unstable)
- [Home Manager](https://github.com/nix-community/home-manager)
- [NixOS](https://search.nixos.org/options)
- [Noogle](https://noogle.dev/)
- [Darwin](https://github.com/LnL7/nix-darwin)
- [NUR](https://github.com/nix-community/NUR)

Also, you can search for arbitrary modules options and/or web pages. See [Custom Search](#custom-search)

## Usage

`nix-search-tv` does not do the search by itself, but rather integrates
with other general purpose fuzzy finders, such as [television](https://github.com/alexpasmantier/television) and [fzf](https://github.com/junegunn/fzf). This way, you can use it by piping results into fzf, embed into NeoVim, Emacs or anything else. Some examples [here](#examples)

### Television

Add `nix.toml` file to your television cables directory with the content below:

```toml
[metadata]
name = "nix"
requirements = ["nix-search-tv"]

[source]
command = "nix-search-tv print"

[preview]
command = "nix-search-tv preview {}"
```

or use the Home Manager option:

```nix
programs.nix-search-tv.enableTelevisionIntegration = true;
```

### fzf

The most straightforward integration might look like:

```sh
alias ns="nix-search-tv print | fzf --preview 'nix-search-tv preview {}' --scheme history"
```

There is also a [nixpkgs.sh](./nixpkgs.sh) script which provides more advanced fzf integration. It is the same search but with the following shortcuts:

- Search only Nixpkgs or Home Manager
- Jump to source code or homepage
- Pipe preview to a pager to see documentation in full screen and select text
- Search GitHub for snippets with the selected package/option
- And more

You can install it like:

```sh
let
  ns = pkgs.writeShellScriptBin "ns" (builtins.readFile ./path/to/nixpkgs.sh);
in {
  environment.systemPackages = [ ns ]
}
```

## Installation

### Nix Package

```nix
environment.systemPackages = [ nix-search-tv ]
```

### Flake

There are many ways how one can install a package from a flake, below is one:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nix-search-tv.url = "github:3timeslazy/nix-search-tv";
  };

  outputs = {
    nixpkgs,
    nix-search-tv,
    ...
  }: {
    nixosConfigurations.system = nixpkgs.lib.nixosSystem {
      modules = [
        {
          environment.systemPackages = [
            nix-search-tv.packages.x86_64-linux.default
          ];
        }
      ];
    };
  };
}
```

### fzf + nixpkgs.sh

by [tonybtw.com](https://www.tonybtw.com/community/nix-search-tv/)

```nix
pkgs.writeShellApplication {
  name = "ns";
  runtimeInputs = with pkgs; [
    fzf
    nix-search-tv
  ];
  text = builtins.readFile "${pkgs.nix-search-tv.src}/nixpkgs.sh";
}
```

## Configuration

By default, the configuration file is looked at `$XDG_CONFIG_HOME/nix-search-tv/config.json`

```jsonc
{
  // What indexes to search by default
  //
  // default:
  //   linux: [nixpkgs, "home-manager", "nur", "nixos"]
  //   darwin: [nixpkgs, "home-manager", "nur", "darwin"]
  "indexes": ["nixpkgs", "nixos", "home-manager", "nur", "noogle"],

  // How often to look for updates and run
  // indexer again
  //
  // default: 1 week (168h)
  "update_interval": "3h2m1s",

  // Where to store the index files
  //
  // default: $XDG_CACHE_HOME/nix-search-tv
  "cache_dir": "path/to/cache/dir",

  // Whether to show the banner when waiting for
  // the indexing
  //
  // default: true
  "enable_waiting_message": true,

  // More about experimental below
  "experimental": {
    "render_docs_indexes": {
      "plasma": "https://nix-community.github.io/plasma-manager/options.xhtml",
    },
    "options_file": {
      "agenix": "<path to options.json>",
    },
  },
}
```

### Custom Search

#### Parse HTML

`nix-search-tv` can parse a documentation HTML page and extract options from it. How to tell if a page can be parsed? To understand that, check the links in the example below and if the documentation page looks exactly like one of them, it probably can be parsed.

```jsonc
{
  "render_docs_indexes": {
    // https://github.com/nix-community/plasma-manager
    "plasma": "https://nix-community.github.io/plasma-manager/options.xhtml",
    
    // Home Manager is also a valid page
    // https://nix-community.github.io/home-manager/options.xhtml
  },
}
```

#### Build-time options.json

The point of this search is to generate the options.json file at nix build time and point `nix-search-tv` to it. Internally, the tool compares previous and the new path and only re-indexes it if the path has changes.

```jsonc
{
  "options_file": {
    // https://github.com/ryantm/agenix
    "agenix": "<path to built options.json>",

    // https://jovian-experiments.github.io/Jovian-NixOS/index.html
    "jovian": "<path to built options.json>",
  },
}
```

Here's one way to generate the options.json files for `agenix` and `nixvim` using [unf](https://git.atagen.co/atagen/unf) and home-manager:

```nix
# flake.nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";

    agenix.url = "github:ryantm/agenix";
    nixvim.url = "github:nix-community/nixvim";

    unf = {
      url = "git+https://git.atagen.co/atagen/unf";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    nixpkgs,
    home-manager,
    ...
  } @ inputs : let
    # It's not required to use `unf`, but even though
    # it brings some dependencies, I use it because its easy.
    mkOpts = system: module:
      inputs.unf.lib.json {
        inherit self;
        pkgs = nixpkgs.legacyPackages.${system};

        # not all modules can be evaluated easily. If your module
        # does not evaluate, try checking this NüschtOS file:
        #   https://github.com/NuschtOS/search.nuschtos.de/blob/main/flake.nix
        modules = [module];
      };
  in {
    nixosConfigurations.hostname = nixpkgs.lib.nixosSystem rec {
      system = "x86_64-linux";
      modules = [
        home-manager.nixosModules.home-manager
        {
          # extraSpecialArgs is used here to pass the options files
          # that depend on the flake inputs to home-manager modules,
          # where configuration files are usually defined.
          home-manager.extraSpecialArgs = {
            inherit inputs;

            agenixOptions = mkOpts system inputs.agenix.nixosModules.default;

            # nixvim provides an options.json file already
            nixvimOptions = inputs.nixvim.packages.${system}.options-json + /share/doc/nixos/options.json
          };
        }
      ];
    };
  };
}

# home.nix
{
  pkgs,
  lib,
  ...
} @ args : {
  xdg.configFile."nix-search-tv/config.json".text = builtins.toJSON {
    experimental = {
      options_file = {
        agenix = "${args.agenixOptions}";
        nixvim = "${args.nixvimOptions}";
      };
    };
  };

  # or, with home-manager
  programs.nix-search-tv = {
    enable = true;
    settings = {
      experimental.options_file = {
        agenix = "${args.agenixOptions}";
        nixvim = "${args.nixvimOptions}";
      };
    };
  };
}
```

<!--TODO: add --json option -->

## Examples

### Custom fzf wrapper

Code and a demo here: https://github.com/3timeslazy/nix-search-tv/issues/20

### Agent Skill

Add nix-search-tv as an [LLM agent skill](https://github.com/0xferrous/agent-stuff/blob/1ded370267321bed7448392920401ca91c3365f6/skills/nix-search/SKILL.md)

### Niri

A niri shortcut to start nix-search-tv as a floating window

```kdl
Mod+N { spawn-sh "ghostty --title=nix-search-tv -e path/to/nixpkgs.sh"; }

window-rule {
    match title="nix-search-tv"
    open-floating true
    default-column-width { proportion 0.75; }
    default-window-height { proportion 0.75; }
}
```

## Credits

This project was inspired and wouldn't exist without work done by [nix-search](https://github.com/diamondburned/nix-search) contributors.
