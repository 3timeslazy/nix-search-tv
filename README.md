# nix-search-tv-install

Fuzzy search for NixOS packages and install them easily with automatic insertion in `configuration.nix` and system rebuild.

The `configuration.nix` file has to respect the [nixfmt](https://github.com/NixOS/nixfmt) formatting.

Fork of [nix-search-tv](https://github.com/3timeslazy/nix-search-tv) with an `install` command. 

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

## Installation

### Nix Package

```nix
environment.systemPackages = [ 
    (nix-search-tv.overrideAttrs (oldAttrs: {
      src = fetchFromGitHub {
        owner = "gschurck";
        repo = "nix-search-tv-install";
        rev = "feat/add-install-command";
        hash = "<latest_sha256_hash>";
      };
    }))
]
```

### Flake

There are many ways how one can install a package from a flake, below is one:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nix-search-tv.url = "github:gschurck/nix-search-tv-install";
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

## Usage

`nix-search-tv` does not do the search by itself, but rather integrates
with other general purpose fuzzy finders, such as [television](https://github.com/alexpasmantier/television) and [fzf](https://github.com/junegunn/fzf)

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

[actions.install]
description = "Install package and rebuild system"
command = "sudo nix-search-tv install {}"
mode = "execute"

[keybindings]
enter = "actions.install"
```

or use the Home Manager option:

```nix
programs.nix-search-tv.enableTelevisionIntegration = true;
```

### fzf

The most straightforward integration might look like:

```sh
alias ns="nix-search-tv print | fzf --preview 'nix-search-tv preview {}' --scheme history | xargs -I {} sudo nix-search-tv install {}"
```

> [!NOTE]
> No matter how you use nix-search-tv with fzf, it's better to add `--scheme history`. That way, the options will be sorted, which makes the search experience better

More advanced integration might be found here in [nixpkgs.sh](./nixpkgs.sh). It is the same search but with the following shortcuts:

- Search only Nixpkgs or Home Manager
- Open package code declaration or homepage
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

## Configuration

By default, the configuration file is looked at `$XDG_CONFIG_HOME/nix-search-tv/config.json`

```jsonc
{
  // What indexes to search by default
  //
  // default:
  //   linux: [nixpkgs, "home-manager", "nur", "nixos"]
  //   darwin: [nixpkgs, "home-manager", "nur", "darwin"]
  "indexes": ["nixpkgs", "home-manager", "nur"],

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

## Searchable package registries

### Builtin

- [Nixpkgs](https://search.nixos.org/packages?channel=unstable)
- [Home Manager](https://github.com/nix-community/home-manager)
- [NixOS](https://search.nixos.org/options)
- [Darwin](https://github.com/LnL7/nix-darwin)
- [NUR](https://github.com/nix-community/NUR)

### Custom

#### Parse HTML

`nix-search-tv` can parse a documentation HTML page and extract options from it. How to tell if a page can be parsed? To understand that, check the links in the example below and if the documentation page looks exactly like one of them, it probably can be parsed.

```jsonc
{
  "render_docs_indexes": {
    // https://github.com/nix-community/plasma-manager
    "plasma": "https://nix-community.github.io/plasma-manager/options.xhtml",
  },
}
```

#### Parse options.json file

The point of this setting is to generate the options file at nix build time and point `nix-search-tv` to them. Internally, the tool compares previous and the new path and only re-indexes it if the path has changes.

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

## Credits

This project was inspired and wouldn't exist without work done by [nix-search](https://github.com/diamondburned/nix-search) contributors.
