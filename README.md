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

## Usage

`nix-search-tv` does not do the search by itself, but rather integrates
with other general purpose fuzzy finders, such as [television](https://github.com/alexpasmantier/television) and [fzf](https://github.com/junegunn/fzf)

### Television

Add `nix_channels.toml` file to your television config directory with the content below:

```toml
[[cable_channel]]
name = "nixpkgs"
source_command = "nix-search-tv print"
preview_command = "nix-search-tv preview {}"
```

### fzf

The most straightforward integration might look like:

```sh
alias ns="nix-search-tv print | fzf --preview 'nix-search-tv preview {}' --scheme history"
```

> [!NOTE]
> No matter how you use nix-search-tv with fzf, it's better to add `--scheme history`. That way, the options will be sorted, which makes the search experience better

More advanced integration might be found here in [nixpkgs.sh](./nixpkgs.sh). It is the same search but with the following shortcuts:

- Search only Nixpkgs or Home Manager
- Open package code declaration or homepage
- Search GitHub for snippets with the selected package/option

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

  "experimental": {
    // nix-search-tv can parse and index documentation pages generated by nixos-render-docs,
    // enabling search functionality for those pages.
    "render_docs_indexes": {
      "nvf": "https://notashelf.github.io/nvf/options.html"
    }
  }
}
```

## Searchable package registries

- [Nixpkgs](https://search.nixos.org/packages?channel=unstable)
- [Home Manager](https://github.com/nix-community/home-manager)
- [NixOS](https://search.nixos.org/options)
- [Darwin](https://github.com/LnL7/nix-darwin)
- [NUR](https://github.com/nix-community/NUR)

You can also search any documentation page generated by nixos-render-docs. See `experimental.render_docs_indexes` in [Configuration](#configuration) section

## Credits

This project was inspired and wouldn't exist without work done by [nix-search](https://github.com/diamondburned/nix-search) contributors.
